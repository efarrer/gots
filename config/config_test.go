package config_test

import (
	"testing"

	"github.com/efarrer/gots/config"
	"github.com/stretchr/testify/require"
)

func TestVolumesToStringsToVolumes(t *testing.T) {
	expected := []config.Volume{
		{
			DockerDir: "dd0",
			HostDir:   "hd0",
		}, {
			DockerDir: "dd1",
			HostDir:   "hd1",
		}, {
			DockerDir: "dd2",
			HostDir:   "hd2",
		}, {
			DockerDir: "dd3",
			HostDir:   "hd3",
		},
	}

	res := config.StringsToVolumes(config.VolumesToStrings(expected))
	require.Equal(t, expected, res)
}

func Test(t *testing.T) {

	/*
		c := config.Config{}
		builder := config.Builder(&c)
		builder.Compute(&c.WorkDir, GetWorkDir(), AppTypeGo)

		builder.Compute(&c.ExecName, GetCmd(), AppTypeGo)
		builder.Request(&c.ExecName, "", "Enter the name of the executable:", AppTypeGo)

		builder.Request(&c.CompilePath, "", "Enter the path to the directory that contains the main.go (e.g. ./cmd/foo): ", AppTypeGo)

		builder.Request(&c.Port, 80, "What TCP port is used by the application (default 80):", AppTypeGo)
	*/

}
