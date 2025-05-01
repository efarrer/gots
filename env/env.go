package env

import (
	"fmt"
	"os"
	"os/exec"
)

func ValidateEnv() {
	docker := "docker"
	_, err := exec.LookPath(docker)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not find '%s': %v\n", docker, err)
		os.Exit(1)
	}

	tailscale := "tailscale"
	_, err = exec.LookPath(tailscale)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not find '%s': %v\n", tailscale, err)
		os.Exit(1)
	}

	jq := "jq"
	_, err = exec.LookPath(jq)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not find '%s': %v\n", jq, err)
		os.Exit(1)
	}
}
