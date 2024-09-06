package mediawiki

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/tozd/go/x"
)

func TestTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		time      string
		precision TimePrecision
	}{
		{"+1994-01-01T00:00:00Z", Minute},
		{"+1952-00-00T00:00:00Z", Year},
		{"+0001-00-00T00:00:00Z", Year},
		{"-0001-00-00T00:00:00Z", Year},
		{"+11994-01-01T00:00:00Z", Minute},
		{"+11952-00-00T00:00:00Z", Year},
		{"+10001-00-00T00:00:00Z", Year},
		{"-10001-00-00T00:00:00Z", Year},
		{"-13798000000-00-00T00:00:00Z", Year},
	}
	for _, test := range tests {
		t.Run(test.time, func(t *testing.T) {
			t.Parallel()

			p, err := parseTime(test.time)
			require.NoError(t, err, "% -+#.1v", err)
			s := formatTime(p, test.precision)
			assert.Equal(t, test.time, s)
		})
	}
}

func TestAmount(t *testing.T) {
	t.Parallel()

	tests := []string{
		"+123.34",
		"-123.34",
		"+0.3333333333333333333333333333333333333333333333333333333333333333333333333333",
		"-2.0000000000000000000000000000000000000000000000000000000000000000000000000001",
		"+0",
	}
	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			t.Parallel()

			in := `"` + test + `"`
			var a Amount
			err := json.Unmarshal([]byte(in), &a)
			require.NoError(t, err)
			out, err := x.MarshalWithoutEscapeHTML(a)
			require.NoError(t, err, "% -+#.1v", err)
			assert.Equal(t, in, string(out))
		})
	}
}
