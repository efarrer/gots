package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/efarrer/gots/config/builder"
	"github.com/efarrer/gots/config/compute"
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

// Deref safely derefs a pointer. For nil returns the zero value
func Deref[A any](pa *A) A {
	if pa == nil {
		var ret A
		return ret
	}
	return *pa
}

// Volume represents a docker volume
type Volume struct {
	DockerDir string
	HostDir   string
}

func VolumesToStrings(vs []Volume) []string {
	if vs == nil {
		return nil
	}
	ret := []string{}
	for _, v := range vs {
		ret = append(ret, v.DockerDir, v.HostDir)
	}

	return ret
}

func StringsToVolumes(strs []string) []Volume {
	ret := []Volume{}
	for i := 0; i < len(strs); i += 2 {
		ret = append(ret, Volume{DockerDir: strs[i], HostDir: strs[i+1]})
	}
	return ret
}

// GetNilFieldNames iterates over a struct and returns the names of fields
// that are not nil (for pointer types) or are not their zero value (for non-pointer types).
func GetNilFieldNames(s interface{}) []string {
	var fieldNames []string
	val := reflect.ValueOf(s)

	// Ensure we're dealing with a struct
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fieldNames
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		// Check for nil for pointer types
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				fieldNames = append(fieldNames, fieldType.Name)
			}
			// Check for empty slices
		} else if field.Kind() == reflect.Slice {
			if field.IsNil() {
				fieldNames = append(fieldNames, fieldType.Name)
			}
		}

	}
	return fieldNames
}

// FilterSlice removes elements from a slice that satisfy the filter function.
// The filter function should return true for elements to be removed.
func FilterSlice[T any](slice []T, filter func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if !filter(item) { // If the filter returns false, keep the item
			result = append(result, item)
		}
	}
	return result
}

// Config the gots configuration
type Config struct {
	Type                     string
	DockerImage              *string  `gots:"go,dockerimage" json:"DockerImage,omitempty"`
	DockerHostname           *string  `gots:"go,dockerimage" json:"DockerHostname,omitempty"`
	ExecName                 *string  `gots:"go" json:"ExecName,omitempty"`
	ExecArgs                 []string `gots:"go" json:"ExecArgs,omitempty"`
	DeprecatedCompileCommand []string `json:"CompileCommand,omitempty"` // Deprecated
	GoCompilePath            *string  `gots:"go" json:"GoCompilePath,omitempty"`
	Port                     *int     `gots:"go,dockerimage" json:"Port,omitempty"`
	Funnel                   *bool    `gots:"go,dockerimage" json:"Funnel,omitempty"`
	DockerVolumes            []Volume `gots:"go,dockerimage" json:"DockerVolumes,omitempty"`
	WorkDir                  *string  `gots:"go,dockerimage" json:"WorkDir,omitempty"`
}

func (c Config) GoCompilePathSafe() string {
	return Deref(c.GoCompilePath)
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
	// Make sure all fields are set
	names := FilterSlice(
		GetNilFieldNames(*c),
		func(field string) bool {
			// Ignore fields that aren't for this app type
			if !builder.GetFieldTags(c, field).Contains(c.Type) {
				return true
			}
			return false
		})

	return len(names) == 0
}

// Migrate performs any migrations that are needed
func (c *Config) Migrate() {
	if len(c.DeprecatedCompileCommand) > 0 {
		c.GoCompilePath = &c.DeprecatedCompileCommand[len(c.DeprecatedCompileCommand)-1]
		c.DeprecatedCompileCommand = nil
	}
	if c.Type == "" {
		c.Type = builder.AppTypeGo
	}

	// For go both the DockerImage and the DockerHostname are the same as the exec name
	if c.Type == builder.AppTypeGo && c.DockerImage == nil {
		c.DockerImage = c.ExecName
	}
	if c.Type == builder.AppTypeGo && c.DockerHostname == nil {
		c.DockerHostname = c.ExecName
	}
	if c.Type == "image" {
		c.Type = builder.AppTypeDockerImage
	}
}

// RequestMissingConfiguration prompts the user for missing configuration parameters
func (c *Config) RequestMissingConfiguration() error {
	// Grab the original configuration to see if anything changed
	origConfiguration := *c

	b := builder.New(os.Stdin, c.Type)

	c.ExecName = builder.Compute(b, c, "ExecName", compute.GetCmd)
	c.ExecName = builder.Request[string](b, c, "ExecName", "", "Enter the name of the executable: ")
	// For go both the DockerImage and the DockerHostname are the same as the exec name
	if c.Type == builder.AppTypeGo {
		c.DockerHostname = c.ExecName
		c.DockerImage = c.ExecName
	}
	c.DockerHostname = builder.Request(b, c, "DockerHostname", "", "Enter the hostname to use in the docker container: ")
	c.DockerImage = builder.Request(b, c, "DockerImage", "", "Enter the name of the docker image to execute: ")
	c.WorkDir = builder.Compute(b, c, "WorkDir", compute.Getwd)
	c.GoCompilePath = builder.Compute(b, c, "GoCompilePath", compute.ComputeGoCompilePath(c.ExecName))
	c.GoCompilePath = builder.Request(b, c, "GoCompilePath", "", "Enter the path to the directory that contains the main.go (e.g. ./cmd/foo): ")
	c.Port = builder.Request[int](b, c, "Port", 80, "What TCP port is used by the application (default 80): ")
	c.ExecArgs = builder.RequestSlice(b, c, "ExecArgs", []string{},
		fmt.Sprintf("Enter the command line arguments to pass to \"%s\". Hit enter after each argument.\n", Deref(c.ExecName)),
		[]string{"Arg %d: "},
	)
	c.Funnel = builder.Request(b, c, "Funnel", false, "Should a Tailscale funnel be started? (y/n): ")

	// DockerVolumes is special in that we want to use a struct not []string so the docker/host paths are unambiguous
	{
		ats := builder.GetFieldTags(c, "DockerVolumes")
		c.DockerVolumes = StringsToVolumes(builder.RequestSliceRaw(b, VolumesToStrings(c.DockerVolumes), []string{},
			fmt.Sprintf("Enter the volumes to mount in the Docker container\n"),
			[]string{
				"Docker dir (absolute path) %d: ",
				"Host dir (absolute path) %d: ",
			},
			ats,
		))
	}

	changed := ""
	if Deref(origConfiguration.DockerImage) != Deref(c.DockerImage) {
		changed += fmt.Sprintf("Docker image: %s\n", *c.DockerImage)
	}
	if Deref(origConfiguration.ExecName) != Deref(c.ExecName) {
		changed += fmt.Sprintf("Executable: %s\n", *c.ExecName)
	}
	if strings.Join(origConfiguration.ExecArgs, "") != strings.Join(c.ExecArgs, "") {
		changed += fmt.Sprintf("Executable arguments: %s\n", strings.Join(c.ExecArgs, ", "))
	}
	if Deref(origConfiguration.GoCompilePath) != Deref(c.GoCompilePath) {
		changed += fmt.Sprintf("Go main.go path %s\n", *c.GoCompilePath)
	}
	if Deref(origConfiguration.Port) != Deref(c.Port) {
		changed += fmt.Sprintf("Listening port %d\n", *c.Port)
	}
	if Deref(origConfiguration.Funnel) != Deref(c.Funnel) {
		changed += fmt.Sprintf("Start a Tailscale funnel: %t\n", *c.Funnel)
	}
	if fmt.Sprintf("%v", origConfiguration.DockerVolumes) != fmt.Sprintf("%v", c.DockerVolumes) {
		for _, vol := range c.DockerVolumes {
			changed += fmt.Sprintf("Volume: %s:%s\n", vol.DockerDir, vol.HostDir)
		}
	}

	if changed != "" {
		fmt.Printf("\n**********************************\n")
		fmt.Println(changed)
		fmt.Printf("**********************************\n\n")
		fmt.Print("Are these changes correct? (y/n): ")
		yOrN := ""
		fmt.Scanf("%s", &yOrN)
		if !strings.HasPrefix(strings.ToLower(yOrN), "y") {
			return fmt.Errorf("Configuration is not correct. Cowardly quitting\n")
		}
	}

	return nil
}

// Save saves the configuration to the .gots file
func (c *Config) Save() error {
	jsonData, err := json.MarshalIndent(*c, "", "  ")
	if err != nil {
		return fmt.Errorf("Unable to JSONify config\n")
	}
	err = os.WriteFile(configPath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("Unable to save config\n")
	}

	fmt.Printf("\nConfig saved to %s\n", configPath)
	return nil
}

type templates struct {
	dstFileName     string
	srcTemplateName string
	srcTemplate     string
}

// Generate creates all files needed to execute the executable in docker (Dockerfile, docker-compose.yaml, etc.)
func (c *Config) Generate(dstDir string) error {
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
			return fmt.Errorf("Unable to open %s\n", t.dstFileName)
		}
		defer file.Close()

		templ, err := template.New(t.dstFileName).Parse(t.srcTemplate)
		if err != nil {
			return fmt.Errorf("Unable to parse %s\n", t.srcTemplateName)
		}

		err = templ.Execute(file, *c)
		if err != nil {
			return fmt.Errorf("Unable to execute template %s %w\n", t.srcTemplateName, err)
		}

		if t.dstFileName == "gots-run" {
			os.Chmod(t.dstFileName, 0755)
			if err != nil {
				return fmt.Errorf("Unable chmod %s\n", t.dstFileName)
			}
		}
	}

	return nil
}
