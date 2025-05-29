package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed Dockerfile.template
var dockerfileTemplate string

//go:embed serve.config.template
var serveConfigTemplate string

//go:embed docker-compose.yaml.template
var dockerComposeTemplate string

//go:embed gots-run.template
var gotsRunTemplate string

const configPath = "./.gots"

// Volume represents a docker volume
type Volume struct {
	DockerDir string
	HostDir   string
}

// Config the gots configuration
type Config struct {
	ExecName       string
	ExecArgs       []string
	CompileCommand []string
	Port           *int
	Funnel         *bool
	DockerVolumes  []Volume
	WorkDir        string
	mutated        bool
	dryRun         bool // When set to true the user isn't prompted for configuration settings
}

// GetCmd checks for a directory structure of ./cmd/<name> and if so it returns <name>. If not it returns ""
func GetCmd() string {
	dirPath := "./cmd/"
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		return ""
	}

	if !fileInfo.IsDir() {
		return ""
	}

	var cmd string
	err = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != dirPath { // Exclude the directory itself
			cmd = filepath.Base(path)
		}
		return nil
	})
	if err != nil {
		return ""
	}
	return cmd
}

// Load loads the .gots (if it exists)
func Load() *Config {
	file, err := os.Open(configPath)
	if err != nil {
		return &Config{}
	}
	defer file.Close()

	cfg := Config{}
	err = json.NewDecoder(file).Decode(&cfg)
	if err != nil {
		return &Config{}
	}
	return &cfg
}

// ValidateComplete validates that the configuration is complete
func (c *Config) ValidateComplete() bool {
	c.dryRun = true
	c.RequestMissingConfiguration()
	return c.mutated
}

// RequestMissingConfiguration prompts the user for missing configuration parameters
func (c *Config) RequestMissingConfiguration() {
	c.requestCurWorkDir().
		requestExecName().
		requestCompileCommand().
		requestExecArgs().
		requestPort().
		requestFunnel().
		requestVolumes()
}

func (c *Config) requestCurWorkDir() *Config {
	if c.WorkDir != "" {
		return c
	}
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get working directory\n")
		os.Exit(1)
	}
	c.WorkDir = wd

	return c
}

func (c *Config) requestExecName() *Config {
	if c.ExecName != "" {
		return c
	}
	c.mutated = true
	if c.dryRun {
		return c
	}

	// Try and auto discover the ExecName
	c.ExecName = GetCmd()
	if c.ExecName != "" {
		return c
	}

	fmt.Print("Enter the name of the executable: ")
	fmt.Scanf("%s", &c.ExecName)

	fmt.Print("Enter the path to the directory that contains the main.go (e.g. ./cmd/foo): ")
	compilePath := ""
	fmt.Scanf("%s", &compilePath)
	c.CompileCommand = []string{"go", "build", compilePath}
	return c
}

func (c *Config) requestCompileCommand() *Config {
	if c.CompileCommand != nil {
		return c
	}
	c.mutated = true
	if c.dryRun {
		return c
	}

	cmdName := GetCmd()
	var compileCommand []string = nil
	if cmdName != "" {
		compileCommand = []string{"go", "build", "./cmd/" + cmdName}
	}

	c.CompileCommand = compileCommand
	return c
}

func (c *Config) requestExecArgs() *Config {
	if c.ExecArgs != nil {
		return c
	}
	c.mutated = true
	if c.dryRun {
		return c
	}

	c.ExecArgs = []string{}
	fmt.Printf("Enter the command line arguments to pass to \"%s\". Hit enter after each argument.\n", c.ExecName)
	for i := 0; true; i++ {
		fmt.Printf("Arg %d: ", i)
		args := ""
		fmt.Scanf("%s", &args)
		if args == "" {
			return c
		}
		c.ExecArgs = append(c.ExecArgs, args)
	}
	return c
}

func (c *Config) requestPort() *Config {
	if c.Port != nil {
		return c
	}
	c.mutated = true
	if c.dryRun {
		return c
	}

	fmt.Print("What TCP port is used by the application (default 80): ")
	port := 80
	fmt.Scanf("%d", &port)
	c.Port = &port
	return c
}

func (c *Config) requestFunnel() *Config {
	if c.Funnel != nil {
		return c
	}
	c.mutated = true
	if c.dryRun {
		return c
	}

	fmt.Print("Should a Tailscale funnel be started? (y/n): ")
	yOrN := ""
	fmt.Scanf("%s", &yOrN)
	if strings.HasPrefix(strings.ToLower(yOrN), "y") {
		val := true
		c.Funnel = &val
	} else {
		val := false
		c.Funnel = &val
	}
	return c
}

func (c *Config) requestVolumes() *Config {
	if c.DockerVolumes != nil {
		return c
	}
	c.mutated = true
	if c.dryRun {
		return c
	}

	c.DockerVolumes = []Volume{}
	fmt.Printf("Enter the volumes to mount in the Docker container\n")
	for i := 0; true; i++ {
		fmt.Printf("Docker dir (absolute path) %d: ", i)
		dockerDir := ""
		fmt.Scanf("%s", &dockerDir)
		if dockerDir == "" {
			return c
		}

		fmt.Printf("Host dir (absolute path) %d: ", i)
		hostDir := ""
		fmt.Scanf("%s", &hostDir)
		if hostDir == "" {
			return c
		}
		c.DockerVolumes = append(c.DockerVolumes, Volume{DockerDir: dockerDir, HostDir: hostDir})
	}
	return c
}

// String outputs a human friendly representation of the configuration.
func (c *Config) String() {
	fmt.Printf("Executable: %s\n", c.ExecName)
	fmt.Printf("Executable arguments: %s\n", strings.Join(c.ExecArgs, ", "))
	fmt.Printf("Command to compile %s: %s\n", c.ExecName, strings.Join(c.CompileCommand, " "))
	fmt.Printf("Start a Tailscale funnel: %t\n", *c.Funnel)
	for _, vol := range c.DockerVolumes {
		fmt.Printf("Volume: %s:%s\n", vol.DockerDir, vol.HostDir)
	}
}

// ConfirmConfiguration prints the configuraiton and asks the user if it is correct.
func (c *Config) ConfirmConfiguration() {
	if !c.mutated {
		return
	}
	if c.dryRun {
		return
	}

	fmt.Printf("\n\n**********************************\n\n")
	c.String()

	fmt.Print("Are these correct? (y/n): ")
	yOrN := ""
	fmt.Scanf("%s", &yOrN)
	if !strings.HasPrefix(strings.ToLower(yOrN), "y") {
		fmt.Fprintf(os.Stderr, "Configuration is not correct. Cowardly quitting\n")
		os.Exit(1)
	}
}

// Save saves the configuration to the .gots file
func (c *Config) Save() {
	if !c.mutated {
		return
	}
	if c.dryRun {
		return
	}

	jsonData, err := json.MarshalIndent(*c, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to JSONify config\n")
		os.Exit(1)
	}
	err = os.WriteFile(configPath, jsonData, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to save config\n")
		os.Exit(1)
	}

	fmt.Printf("\nConfig saved to %s\n", configPath)
}

type templates struct {
	dstFileName     string
	srcTemplateName string
	srcTemplate     string
}

// Generate creates all files needed to execute the executable in docker (Dockerfile, docker-compose.yaml, etc.)
func (c *Config) Generate(dstDir string) {
	for _, t := range []templates{
		{
			dstFileName:     "Dockerfile",
			srcTemplateName: "Dockerfile.template",
			srcTemplate:     dockerfileTemplate,
		}, {
			dstFileName:     "serve.config",
			srcTemplateName: "serve.config.template",
			srcTemplate:     serveConfigTemplate,
		}, {
			dstFileName:     "docker-compose.yaml",
			srcTemplateName: "docker-compose.yaml.template",
			srcTemplate:     dockerComposeTemplate,
		}, {
			dstFileName:     "gots-run",
			srcTemplateName: "gots-run.template",
			srcTemplate:     gotsRunTemplate,
		},
	} {
		file, err := os.Create(t.dstFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to open %s\n", t.dstFileName)
			os.Exit(1)
		}
		defer file.Close()

		templ, err := template.New(t.dstFileName).Parse(t.srcTemplate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to parse %s\n", t.srcTemplateName)
			os.Exit(1)
		}

		err = templ.Execute(file, *c)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable execute %s\n", t.srcTemplateName)
			os.Exit(1)
		}

		if t.dstFileName == "gots-run" {
			os.Chmod(t.dstFileName, 0755)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable chmod %s\n", t.dstFileName)
				os.Exit(1)
			}
		}
	}
}
