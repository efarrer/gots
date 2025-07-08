package builder

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

const (
	AppTypeGo          = "go"
	AppTypeDockerImage = "dockerimage"
	AppTypeDockerFile  = "dockerfile"
)

// GetFieldValueByName retrieves the field value from a struct by name
func GetFieldValueByName[T any](s interface{}, fieldName string) T {
	var zeroValue T // Get the zero value of type T

	// Get the reflect.Value of the input.
	val := reflect.ValueOf(s)

	// If it's a pointer, dereference it to get the underlying struct value.
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Ensure we are dealing with a struct.
	if val.Kind() != reflect.Struct {
		panic(fmt.Sprintf("Warning: Input is not a struct or a pointer to a struct. Got %v\n", val.Kind()))
		return zeroValue
	}

	// Get the field by its name.
	field := val.FieldByName(fieldName)

	// Check if the field was found and is valid.
	if !field.IsValid() {
		panic(fmt.Sprintf("Warning: Field '%s' not found or is unexported in the struct.\n", fieldName))
		return zeroValue
	}

	// Check if the field's type is assignable to the generic type T.
	// This is where generics provide compile-time type safety.
	if !field.CanConvert(reflect.TypeOf(zeroValue)) {
		panic(fmt.Sprintf("Warning: Field '%s' type (%v) is not assignable to expected type %T.\n",
			fieldName, field.Type(), zeroValue))
		return zeroValue
	}

	// Convert the field's value to the generic type T and return it.
	return field.Convert(reflect.TypeOf(zeroValue)).Interface().(T)
}

// GetGotsTags returns the value of the "gots" tags.
func GetFieldTags(s interface{}, fieldName string) mapset.Set[string] {
	tags := mapset.NewSet[string]()

	// Get the reflect.Value of the struct
	val := reflect.ValueOf(s)

	// If it's a pointer, dereference it
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Ensure we are dealing with a struct
	if val.Kind() != reflect.Struct {
		panic(fmt.Sprintf("Warning: Input is not a struct or a pointer to a struct. Got %v\n", val.Kind()))
		return tags
	}

	// Get the reflect.Type of the struct
	typ := val.Type()

	// Iterate over the fields of the struct
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Name == fieldName {
			// Found the field, now extract its tags
			if tagValue, ok := field.Tag.Lookup("gots"); ok {
				tags.Append(strings.Split(tagValue, ",")...)
			}
			return tags
		}
	}

	panic(fmt.Sprintf("Warning: Field '%s' not found in the struct.\n", fieldName))
	return tags
}

// A Builder sets the needed files in a config.Config
type Builder struct {
	at          string
	needsConfig bool
	dryRun      bool
	input       io.Reader
}

// New creates a new Builder for the given app type
func New(reader io.Reader, at string) *Builder {
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
func Compute[V comparable](b *Builder, strct any, fldName string, fn func() (V, error)) *V {
	val := GetFieldValueByName[*V](strct, fldName)
	ats := GetFieldTags(strct, fldName)
	if val != nil {
		return val
	}
	b.needsConfig = true
	if b.dryRun {
		return val
	}

	if ats.Contains(b.at) {
		v, err := fn()
		if err != nil {
			return val
		}
		val = &v
		return val
	}
	return val
}

// Request the user provide a value
func Request[V any](b *Builder, strct any, fldName string, def V, request string) *V {
	val := GetFieldValueByName[*V](strct, fldName)
	ats := GetFieldTags(strct, fldName)
	var vals []V
	if val == nil {
		vals = nil
	} else {
		vals = append(vals, *val)
	}
	res := RequestSliceRaw(b, vals, []V{def}, request, nil, ats)
	if len(res) == 1 {
		return &res[0]
	}
	return nil
}

// RequestSlice asks the user provide a slice of value
func RequestSlice[V any](b *Builder, strct any, fldName string, def []V, request string, subrequests []string) []V {
	vals := GetFieldValueByName[[]V](strct, fldName)
	ats := GetFieldTags(strct, fldName)
	return RequestSliceRaw(b, vals, def, request, subrequests, ats)
}

func RequestSliceRaw[V any](b *Builder, vals []V, def []V, request string, subrequests []string, ats mapset.Set[string]) []V {
	if vals != nil {
		return vals
	}
	b.needsConfig = true
	if b.dryRun {
		return vals
	}

	if !ats.Contains(b.at) {
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
