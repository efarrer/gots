package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/efarrer/gots/config"
	"github.com/efarrer/gots/env"
	"github.com/efarrer/gots/run"
)

const TS_AUTHKEY_ERR = 9 // Magic number that is used to know if we need to set TS_AUTHKEY (see config/gots-run.template)

func main() {
	configFlag := false
	generateFlag := false
	startFlag := false
	stopFlag := false
	restartFlag := false
	flag.BoolVar(&configFlag, "config", false, "Creates the .gots configuration file, based on user input.")
	flag.BoolVar(&generateFlag, "generate", false, "Creates the Docker files and scripts to run executable in Docker with Tailscale.")
	flag.BoolVar(&startFlag, "start", false, "Start the command in Docker with Tailscale.")
	flag.BoolVar(&stopFlag, "stop", false, "Stop the Docker containers.")
	flag.BoolVar(&restartFlag, "restart", false, "Stop then start the Docker containers..")
	flag.Parse()

	env.ValidateEnv()

	if restartFlag {
		startFlag = true
		stopFlag = true
	}

	if !configFlag && !startFlag && !generateFlag && !stopFlag {
		flag.Usage()
	}

	cfg := config.Load()

	// Config
	if configFlag {
		cfg.Migrate()

		err := cfg.RequestMissingConfiguration()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			os.Exit(1)
		}
		err = cfg.Save()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			os.Exit(1)
		}
		return
	}

	// Validate for Generate or Run
	if generateFlag || startFlag || stopFlag {
		if !cfg.ValidateComplete() {
			fmt.Fprintf(os.Stderr, "Configuration is not complete re-run gots with -config\n")
			return
		}
	}

	// Generate
	if generateFlag {
		err := cfg.Generate("./")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			os.Exit(1)
		}
		return
	}

	// Make a temp dir
	tempDir, err := os.MkdirTemp("", "gots")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create temp dir %s\n", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a subdirectory so the docker containers have consistent names
	tempDir = path.Join(tempDir, *cfg.ExecName)
	err = os.Mkdir(tempDir, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create %s %s\n", tempDir, err)
		return
	}

	// Copy .gots to temp dir
	data, err := os.ReadFile(".gots")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read .gots %s\n", err)
		return
	}
	err = os.WriteFile(path.Join(tempDir, ".gots"), data, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write .gots %s\n", err)
		return
	}

	// Change to temp dir
	err = os.Chdir(tempDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to change to %s dir %s\n", tempDir, err)
		return
	}

	// Generate files in temp dir
	cfg.Generate("./")

	// Run
	if startFlag {
		stdout, stderr, err := run.RunWithOutput("./gots-run")
		if err != nil {
			var exitError *exec.ExitError
			if errors.As(err, &exitError) {
				if exitError.ExitCode() == TS_AUTHKEY_ERR {
					fmt.Fprintf(os.Stderr, "TS_AUTHKEY environment variable must be set\n")
					return
				}
			}
			fmt.Fprintf(os.Stderr, "Unable to execute gots-run %s\n", err)
			fmt.Fprintf(os.Stderr, "Stdout:\n")
			fmt.Fprintf(os.Stderr, "%s\n", stdout)
			fmt.Fprintf(os.Stderr, "Stderr:\n")
			fmt.Fprintf(os.Stderr, "%s\n", stderr)
		}
		return
	}

	// Stop
	if stopFlag {
		stdout, stderr, err := run.RunWithOutput("./gots-run", "-stop")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to execute gots-run -stop %s\n", err)
			fmt.Fprintf(os.Stderr, "Stdout:\n")
			fmt.Fprintf(os.Stderr, "%s\n", stdout)
			fmt.Fprintf(os.Stderr, "Stderr:\n")
			fmt.Fprintf(os.Stderr, "%s\n", stderr)
		}
		return
	}
}
