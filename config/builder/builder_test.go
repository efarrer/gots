package builder_test

import (
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/efarrer/gots/config/builder"
	"github.com/stretchr/testify/require"
)

func Ptr[T any](t T) *T {
	return &t
}

func TestGetFieldValueByName(t *testing.T) {
	str := struct {
		A *int
		B int
		C string
	}{
		A: Ptr(8),
		B: 9,
		C: "Hi",
	}

	require.Equal(t, Ptr(8), builder.GetFieldValueByName[*int](str, "A"))
	require.Equal(t, 9, builder.GetFieldValueByName[int](str, "B"))
	require.Equal(t, "Hi", builder.GetFieldValueByName[string](str, "C"))
}

func TestGetFieldTags(t *testing.T) {
	str := struct {
		A int
		B int `gots:""`
		C int `gots:"a"`
		D int `gots:"a,b,c"`
	}{}

	require.Equal(t, []string{}, builder.GetFieldTags(str, "A").ToSlice())
	require.Equal(t, []string{""}, builder.GetFieldTags(str, "B").ToSlice())
	require.Equal(t, []string{"a"}, builder.GetFieldTags(str, "C").ToSlice())
	res := builder.GetFieldTags(str, "D").ToSlice()
	slices.Sort(res)
	require.Equal(t, []string{"a", "b", "c"}, res)
}

func TestCompute(t *testing.T) {
	type govalue struct {
		Value *string `gots:"go"`
	}
	type novalue struct {
		Value *string `gots:"nope"`
	}

	t.Run("uses the existing value if set", func(t *testing.T) {
		value := govalue{Value: Ptr("test value")}
		b := builder.New(os.Stdin, builder.AppTypeGo)

		result := builder.Compute(b, value, "Value", func() (string, error) { return *value.Value, nil })

		require.Equal(t, value.Value, result)
		require.False(t, b.NeedsConfig())
	})

	t.Run("Just sets NeedsConfig for dry-run", func(t *testing.T) {
		nilvalue := govalue{}
		b := builder.New(os.Stdin, builder.AppTypeGo).DryRun()

		result := builder.Compute(b, nilvalue, "Value", func() (string, error) { return "test value", nil })

		require.Nil(t, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("respects AppType", func(t *testing.T) {
		nonilvalue := novalue{}
		b := builder.New(os.Stdin, builder.AppTypeGo)

		result := builder.Compute(b, nonilvalue, "Value", func() (string, error) { return "some value", nil })

		require.Nil(t, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("sets the computed value", func(t *testing.T) {
		value := "test value"
		nilvalue := govalue{}
		b := builder.New(os.Stdin, builder.AppTypeGo)

		result := builder.Compute(b, nilvalue, "Value", func() (string, error) { return value, nil })

		require.Equal(t, value, *result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("respects AppType", func(t *testing.T) {
		nilvalue := govalue{}
		b := builder.New(os.Stdin, builder.AppTypeGo)

		result := builder.Compute(b, nilvalue, "Value", func() (string, error) { return "", errors.New("some error") })

		require.Nil(t, result)
		require.True(t, b.NeedsConfig())
	})
}

func TestRequest(t *testing.T) {
	type gointvalue struct {
		Value *int `gots:"go"`
	}
	type nointvalue struct {
		Value *int
	}
	type gostringvalue struct {
		Value *string `gots:"go"`
	}
	type goboolvalue struct {
		Value *bool `gots:"go"`
	}
	t.Run("uses the existing value if set", func(t *testing.T) {
		intvalue := gointvalue{Value: Ptr(80)}
		value := 80
		b := builder.New(os.Stdin, builder.AppTypeGo)

		result := builder.Request(b, intvalue, "Value", 8080, "")

		require.Equal(t, value, *result)
		require.False(t, b.NeedsConfig())
	})

	t.Run("Just sets NeedsConfig for dry-run", func(t *testing.T) {
		nilintvalue := gointvalue{}
		b := builder.New(os.Stdin, builder.AppTypeGo).DryRun()

		result := builder.Request[int](b, nilintvalue, "Value", 0, "")

		require.Nil(t, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("respects AppType", func(t *testing.T) {
		nonilintvalue := nointvalue{}
		b := builder.New(os.Stdin, builder.AppTypeGo)

		result := builder.Request[int](b, nonilintvalue, "Value", 0, "")

		require.Nil(t, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("Handles strings", func(t *testing.T) {
		nilstringvalue := gostringvalue{}
		value := "teststring"
		expectedValue := value

		b := builder.New(strings.NewReader(value), builder.AppTypeGo)

		result := builder.Request[string](b, nilstringvalue, "Value", "", "")

		require.Equal(t, expectedValue, *result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("Handles int", func(t *testing.T) {
		nilintvalue := gointvalue{}
		value := "738839"
		expectedValue := 738839

		b := builder.New(strings.NewReader(value), builder.AppTypeGo)

		result := builder.Request[int](b, nilintvalue, "Value", 0, "")

		require.Equal(t, expectedValue, *result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("Handles bool (true)", func(t *testing.T) {
		nilboolvalue := goboolvalue{}
		value := "yes"
		expectedValue := true

		b := builder.New(strings.NewReader(value), builder.AppTypeGo)

		result := builder.Request(b, nilboolvalue, "Value", false, "")

		require.Equal(t, expectedValue, *result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("Handles bool (false)", func(t *testing.T) {
		nilboolvalue := goboolvalue{}
		value := "no"
		expectedValue := false

		b := builder.New(strings.NewReader(value), builder.AppTypeGo)

		result := builder.Request(b, nilboolvalue, "Value", false, "")

		require.Equal(t, expectedValue, *result)
		require.True(t, b.NeedsConfig())
	})
}

func TestRequestSlice(t *testing.T) {
	type gointslicevalue struct {
		Value []int `gots:"go"`
	}
	type nointslicevalue struct {
		Value []int `gots:"nope"`
	}
	type gostringslicevalue struct {
		Value []string `gots:"go"`
	}
	type goboolslicevalue struct {
		Value []bool `gots:"go"`
	}
	t.Run("uses the existing value if set", func(t *testing.T) {
		intslicevalue := gointslicevalue{Value: []int{80}}
		value := []int{80}
		b := builder.New(os.Stdin, builder.AppTypeGo)

		result := builder.RequestSlice(b, intslicevalue, "Value", []int{8080}, "", []string{""})

		require.Equal(t, value, result)
		require.False(t, b.NeedsConfig())
	})

	t.Run("Just sets NeedsConfig for dry-run", func(t *testing.T) {
		nilintslicevalue := gointslicevalue{}
		b := builder.New(os.Stdin, builder.AppTypeGo).DryRun()

		result := builder.RequestSlice(b, nilintslicevalue, "Value", []int{8080}, "", nil)

		require.Nil(t, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("respects AppType", func(t *testing.T) {
		nonilintslicevalue := nointslicevalue{}
		b := builder.New(os.Stdin, builder.AppTypeGo)

		result := builder.RequestSlice(b, nonilintslicevalue, "Value", []int{8080}, "", nil)

		require.Nil(t, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("Handles strings", func(t *testing.T) {
		nilstringslicevalue := gostringslicevalue{}
		value := "8080"
		expectedValue := []string{"8080"}

		b := builder.New(strings.NewReader(value), builder.AppTypeGo)

		result := builder.RequestSlice(b, nilstringslicevalue, "Value", []string{"8080"}, "", []string{"Port %d: "})

		require.Equal(t, expectedValue, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("Handles int", func(t *testing.T) {
		nilintslicevalue := gointslicevalue{}
		value := "738839"
		expectedValue := []int{738839}

		b := builder.New(strings.NewReader(value), builder.AppTypeGo)

		result := builder.RequestSlice(b, nilintslicevalue, "Value", []int{8080}, "", []string{"Port %d: "})

		require.Equal(t, expectedValue, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("Handles bool (true)", func(t *testing.T) {
		nilboolslicevalue := goboolslicevalue{}
		value := "yes"
		expectedValue := []bool{true}

		b := builder.New(strings.NewReader(value), builder.AppTypeGo)

		result := builder.RequestSlice(b, nilboolslicevalue, "Value", []bool{false}, "", []string{"Port %d: "})

		require.Equal(t, expectedValue, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("Handles bool (false)", func(t *testing.T) {
		nilboolslicevalue := goboolslicevalue{}
		value := "no"
		expectedValue := []bool{false}

		b := builder.New(strings.NewReader(value), builder.AppTypeGo)

		result := builder.RequestSlice(b, nilboolslicevalue, "Value", []bool{false}, "", []string{"Port %d: "})

		require.Equal(t, expectedValue, result)
		require.True(t, b.NeedsConfig())
	})
}

func TestGetWorkDir(t *testing.T) {
	require.NotEmpty(t, builder.GetWorkDir())
}

func TestGetCmd(t *testing.T) {
	require.Empty(t, builder.GetCmd())
}
