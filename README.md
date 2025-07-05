# gots
The easiest way to run your program in Tailscale.

# Installation
    > go install github.com/efarrer/gots@latest

# Usage
It takes only two steps to run Go application in your Tailscale network. From the applications code directory run:

    > gots -config <target type>
    > gots -start

# Argument Details
## -config
Runs the configuration wizard and outputs the .gots file with the Docker/Tailscale parameters.
## -start
Runs the application in Tailscale. The first time this is used the TS_AUTHKEY env var must be set with a Tailscale auth key.
## -stop
Stops the application.
## -generate
Generates Docker config files and a script to run the command in Tailscale. Useful if you want to add additional customizations.
## -update
Updates the docker images used to run the application.

# Prerequisits
* Docker
* Tailscale
* bash
* jq
* Go compiler (for go target type).

