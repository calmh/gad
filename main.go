package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var (
	listenAddr    = os.Getenv("GAD_LISTEN_ADDRESS")
	deployCommand = os.Getenv("GAD_DEPLOY_COMMAND")
	secret        = os.Getenv("GAD_GITHUB_SECRET")

	deployTrigger = runDeployer()
)

func main() {
	log.SetOutput(os.Stdout)

	if listenAddr == "" {
		listenAddr = ":8080"
	}
	if deployCommand == "" {
		deployCommand = "git pull"
	}

	log.Println("Listening on", listenAddr)
	log.Printf(`Will run "%s" to deploy`, deployCommand)
	if secret == "" {
		log.Println("Accepting all POSTs without authentication")
	} else {
		log.Println("Using GitHub HMAC authentication")
	}

	http.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

// handleRequest is called when a request comes in. It verifies that it is a
// POST request with the expected parameters and performs a deploy or returns
// an error to the client, as appropriate.
func handleRequest(w http.ResponseWriter, r *http.Request) {
	// We only expect POST requests here.
	if r.Method != "POST" {
		http.Error(w, "POST Expected", http.StatusMethodNotAllowed)
		return
	}

	// If there's a secret set, we except the signature to match.
	if secret != "" {
		// The signature is a SHA1 HMAC of the request body.
		mac := hmac.New(sha1.New, []byte(secret))
		io.Copy(mac, r.Body)
		sig := fmt.Sprintf("sha1=%x", mac.Sum(nil))

		// If it doesn't match the included header, return 401 Unauthorized
		// and abort.
		if hubSig := r.Header.Get("X-Hub-Signature"); hubSig != sig {
			log.Println("Incorrect signature; %s != %s", hubSig, sig)
			http.Error(w, "Incorrect Secret", http.StatusUnauthorized)
			return
		}
	}

	// Perform a deploy
	resultChan := make(chan error)
	select {
	case deployTrigger <- resultChan:
		// A deploy was triggered.
		err := <-resultChan
		if err != nil {
			log.Println("Deploy error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 200 OK is default

	default:
		// A deploy is already in progress. We return 409 Conflict to indicate
		// that we're not doing another deploy in parallell.
		http.Error(w, "Deploy in Progress", http.StatusConflict)
		return
	}
}

// runDeployer starts the routine that handles the actual deployments and
// returns and channel. A deployment is requested by sending an error channel
// to this channel. A read from the error channel will result in either nil
// (deployment successfull) or an error. We do this on a separate routine
// because we want at most one deployment running at any given instance.
func runDeployer() chan chan error {
	trigger := make(chan chan error)
	go func() {
		for resultChan := range trigger {
			resultChan <- performDeploy()
		}
	}()
	return trigger
}

// performDeploy run the actual deployment command. If it returns a nonzero
// exit status, an error is returned that describes the exit status and
// contains any output from the command.
func performDeploy() error {
	parts := strings.Fields(deployCommand)
	cmd := exec.Command(parts[0], parts[1:]...)
	bs, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s\n---\n%s", err.Error(), bs)
	}
	return nil
}
