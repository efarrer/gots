package builder

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type AppType string

const (
	AppTypeGo = "go"
)

// A Builder sets the needed files in a config.Config
type Builder struct {
	at          AppType
	needsConfig bool
	dryRun      bool
	input       io.Reader
}

// New creates a new Builder for the given AppType
func New(reader io.Reader, at AppType) *Builder {
	return &Builder{
		at:          at,
		needsConfig: false,
		dryRun:      false,
		input:       reader,
	}
}

func (b *Builder) DryRun() *Builder {
	b.dryRun = true
	return b
}

func (b *Builder) NeedsConfig() bool {
	return b.needsConfig
}

// Compute a value. Return nil if it can't be computed
func Compute[V comparable](b *Builder, val *V, fn func() (V, error), ats ...AppType) *V {
	if val != nil {
		return val
	}
	b.needsConfig = true
	if b.dryRun {
		return val
	}

	for _, at := range ats {
		if at == b.at {
			v, err := fn()
			if err != nil {
				return val
			}
			val = &v
			return val
		}
	}
	return val
}

// Request the user provide a value
func Request[V any](b *Builder, val *V, def V, request string, ats ...AppType) *V {
	var vals []V
	if val == nil {
		vals = nil
	} else {
		vals = append(vals, *val)
	}
	res := RequestSlice(b, vals, []V{def}, request, nil, ats...)
	if len(res) == 1 {
		return &res[0]
	}
	return nil
}

// Request the user provide a slice of value
func RequestSlice[V any](b *Builder, vals []V, def []V, request string, subrequests []string, ats ...AppType) []V {
	if vals != nil {
		return vals
	}
	b.needsConfig = true
	if b.dryRun {
		return vals
	}

	matches := false
	for _, at := range ats {
		if at == b.at {
			matches = true
			break
		}
	}
	if !matches {
		return vals
	}

	format := ""
	switch any(def).(type) {
	case []string:
		format = "%s"
	case []bool: // Reads in a string then converts to a bool
		format = "%s"
	case []int:
		format = "%d"
	default:
		panic(fmt.Sprintf("Need a format specifier for %#v\n", def))
	}

	// Find the minimal number of expected responses.
	minExpectedSubresponses := len(subrequests)

	var thisVals []V
	fmt.Println(request)
	for i := 0; true; i++ {
		// Find the right subrequest to print and print it
		if len(subrequests) > 0 {
			subrequest := subrequests[i%len(subrequests)]
			fmt.Printf(subrequest, i)
		}

		var thisVal V
		switch any(def).(type) {
		case []bool:
			yOrN := ""
			count, _ := fmt.Fscanf(b.input, format, &yOrN)
			if count == 0 {
				return vals
			}
			if strings.HasPrefix(strings.ToLower(yOrN), "y") {
				// Use reflection to work around the fact that we know that val is a bool, but the type system does not.
				reflect.ValueOf(&thisVal).Elem().SetBool(true)
			} else {
				// Use reflection to work around the fact that we know that val is a bool, but the type system does not.
				reflect.ValueOf(&thisVal).Elem().SetBool(false)
			}
		default:
			count, _ := fmt.Fscanf(b.input, format, &thisVal)
			if count == 0 {
				// Never read anything so just return the default
				if len(vals) == 0 {
					return def
				}
				// We did read stuff in the past so return it
				return vals
			}
		}
		thisVals = append(thisVals, thisVal)

		// We are just reading a single value (not a slice) so return it
		if minExpectedSubresponses == 0 {
			return thisVals
		}

		// If we've gotten a set of responses then save them.
		if len(thisVals) == minExpectedSubresponses {
			vals = append(vals, thisVals...)
			thisVals = []V{}
		}
	}
	return vals
}

// GetWorkDir returns the current working directory.
func GetWorkDir() string {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get working directory\n")
		os.Exit(1)
	}
	return wd
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
