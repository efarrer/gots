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

func TestGetNilFieldNames(t *testing.T) {
	i := 0
	names := config.GetNilFieldNames(struct {
		a *int
		b *int
		c *int
		d []string
		e []string
	}{
		a: &i,
		b: nil,
		c: &i,
		d: nil,
		e: []string{},
	})

	require.Equal(t, []string{"b", "d"}, names)
}

func TestFilterSlice(t *testing.T) {
	resp := config.FilterSlice([]string{"A", "B", "C"}, func(s string) bool { return s == "B" })

	require.Equal(t, []string{"A", "C"}, resp)
}
