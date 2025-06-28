package builder_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/efarrer/gots/config/builder"
	"github.com/stretchr/testify/require"
)

func TestCompute(t *testing.T) {
	t.Run("uses the existing value if set", func(t *testing.T) {
		value := "test value"
		b := builder.New(os.Stdin, builder.AppTypeGo)

		result := builder.Compute(b, &value, func() (string, error) { return value, nil }, builder.AppTypeGo)

		require.Equal(t, value, *result)
		require.False(t, b.NeedsConfig())
	})

	t.Run("Just sets NeedsConfig for dry-run", func(t *testing.T) {
		b := builder.New(os.Stdin, builder.AppTypeGo).DryRun()

		result := builder.Compute(b, nil, func() (string, error) { return "test value", nil }, builder.AppTypeGo)

		require.Nil(t, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("respects AppType", func(t *testing.T) {
		b := builder.New(os.Stdin, builder.AppTypeGo)

		result := builder.Compute(b, nil, func() (string, error) { return "some value", nil }, builder.AppType("none"))

		require.Nil(t, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("sets the computed value", func(t *testing.T) {
		value := "test value"
		b := builder.New(os.Stdin, builder.AppTypeGo)

		result := builder.Compute(b, nil, func() (string, error) { return value, nil }, builder.AppTypeGo)

		require.Equal(t, value, *result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("respects AppType", func(t *testing.T) {
		b := builder.New(os.Stdin, builder.AppTypeGo)

		result := builder.Compute(b, nil, func() (string, error) { return "", errors.New("some error") }, builder.AppTypeGo)

		require.Nil(t, result)
		require.True(t, b.NeedsConfig())
	})
}

func TestRequest(t *testing.T) {
	t.Run("uses the existing value if set", func(t *testing.T) {
		value := 80
		b := builder.New(os.Stdin, builder.AppTypeGo)

		result := builder.Request(b, &value, 8080, "", builder.AppTypeGo)

		require.Equal(t, value, *result)
		require.False(t, b.NeedsConfig())
	})

	t.Run("Just sets NeedsConfig for dry-run", func(t *testing.T) {
		b := builder.New(os.Stdin, builder.AppTypeGo).DryRun()

		result := builder.Request(b, nil, 8080, "", builder.AppTypeGo)

		require.Nil(t, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("respects AppType", func(t *testing.T) {
		b := builder.New(os.Stdin, builder.AppTypeGo)

		result := builder.Request(b, nil, 8080, "", builder.AppType("none"))

		require.Nil(t, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("Handles strings", func(t *testing.T) {
		value := "teststring"
		expectedValue := value

		b := builder.New(strings.NewReader(value), builder.AppTypeGo)

		result := builder.Request(b, nil, "8080", "", builder.AppTypeGo)

		require.Equal(t, expectedValue, *result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("Handles int", func(t *testing.T) {
		value := "738839"
		expectedValue := 738839

		b := builder.New(strings.NewReader(value), builder.AppTypeGo)

		result := builder.Request(b, nil, 8080, "", builder.AppTypeGo)

		require.Equal(t, expectedValue, *result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("Handles bool (true)", func(t *testing.T) {
		value := "yes"
		expectedValue := true

		b := builder.New(strings.NewReader(value), builder.AppTypeGo)

		result := builder.Request(b, nil, false, "", builder.AppTypeGo)

		require.Equal(t, expectedValue, *result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("Handles bool (false)", func(t *testing.T) {
		value := "no"
		expectedValue := false

		b := builder.New(strings.NewReader(value), builder.AppTypeGo)

		result := builder.Request(b, nil, false, "", builder.AppTypeGo)

		require.Equal(t, expectedValue, *result)
		require.True(t, b.NeedsConfig())
	})
}

func TestRequestSlice(t *testing.T) {
	t.Run("uses the existing value if set", func(t *testing.T) {
		value := []int{80}
		b := builder.New(os.Stdin, builder.AppTypeGo)

		result := builder.RequestSlice(b, value, []int{8080}, "", []string{""}, builder.AppTypeGo)

		require.Equal(t, value, result)
		require.False(t, b.NeedsConfig())
	})

	t.Run("Just sets NeedsConfig for dry-run", func(t *testing.T) {
		b := builder.New(os.Stdin, builder.AppTypeGo).DryRun()

		result := builder.RequestSlice(b, nil, []int{8080}, "", nil, builder.AppTypeGo)

		require.Nil(t, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("respects AppType", func(t *testing.T) {
		b := builder.New(os.Stdin, builder.AppTypeGo)

		result := builder.RequestSlice(b, nil, []int{8080}, "", nil, builder.AppType("none"))

		require.Nil(t, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("Handles strings", func(t *testing.T) {
		value := "8080"
		expectedValue := []string{"8080"}

		b := builder.New(strings.NewReader(value), builder.AppTypeGo)

		result := builder.RequestSlice(b, nil, []string{"8080"}, "", []string{"Port %d: "}, builder.AppTypeGo)

		require.Equal(t, expectedValue, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("Handles int", func(t *testing.T) {
		value := "738839"
		expectedValue := []int{738839}

		b := builder.New(strings.NewReader(value), builder.AppTypeGo)

		result := builder.RequestSlice(b, nil, []int{8080}, "", []string{"Port %d: "}, builder.AppTypeGo)

		require.Equal(t, expectedValue, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("Handles bool (true)", func(t *testing.T) {
		value := "yes"
		expectedValue := []bool{true}

		b := builder.New(strings.NewReader(value), builder.AppTypeGo)

		result := builder.RequestSlice(b, nil, []bool{false}, "", []string{"Port %d: "}, builder.AppTypeGo)

		require.Equal(t, expectedValue, result)
		require.True(t, b.NeedsConfig())
	})

	t.Run("Handles bool (false)", func(t *testing.T) {
		value := "no"
		expectedValue := []bool{false}

		b := builder.New(strings.NewReader(value), builder.AppTypeGo)

		result := builder.RequestSlice(b, nil, []bool{false}, "", []string{"Port %d: "}, builder.AppTypeGo)

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
