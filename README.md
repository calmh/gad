GitHub Auto Deployer
====================

This is a simple program to perform deploys on web hooks. It has the
following redeeming features:

 - Single self contained binary - easy to deploy.

 - Supports GitHub HMAC authentication to validate incoming requests.

Deploying
---------

 - Drop the binary somewhere useful, lets say `/usr/local/bin` for this
   guide.

 - Decide on how to make sure it stays running. This is platform
   dependent, but upstart, svrun, systemd, SMF and so on are all valid
   choices.

 - Generate a secret key. This will be used to authenticate requests
   from GitHub.

 - Set configuration through environment variables:

   - `GAD_LISTEN_ADDRESS`: the address to listen for HTTP requests on

   - `GAD_DEPLOY_COMMAND`: the command to run when performing deploys

   - `GAD_GITHUB_SECRET`: the secret generated above, used to authenticate
     requests from GitHub

 - Start it up.

 - Configure GitHub.

   - Add a WebHook.

   - Set the "Payload URL" to the address and listen port of the machine
     to receive the push hooks.

   - Set the "Secret" to the secret generated above.

   - Test it; you can see requests and responses under "Recent
     Deliveries".

Example
-------

A simple script to correctly start `gad` might then look like this:

```sh
cd /srv/www/my-site-to-deploy
export GAD_LISTEN_ADDRESS=:8876
export GAD_DEPLOY_COMMAND="git pull"
export GAD_GITHUB_SECRET=86486a6c-b488-11e4-9ff9-679075158ddc
exec /usr/local/bin/gad
```

Note that we `cd` to the working directory where the deploy command will
be run.

This would then correspond to the payload URL
`http://your-site-address:8876/` and the secret
`86486a6c-b488-11e4-9ff9-679075158ddc`.

For security reasons, the command is not run through a shell, so
wildcards and semicolon separated commands can't be used un
`GAD_DEPLOY_COMMAND`. You you need that, create a shell script for the
deploy and set this as the deploy command.
