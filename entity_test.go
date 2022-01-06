package mediawiki

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTime(t *testing.T) {
	tests := []struct {
		time      string
		precision TimePrecision
	}{
		{"+1994-01-01T00:00:00Z", Minute},
		{"+1952-00-00T00:00:00Z", Year},
		{"+0001-00-00T00:00:00Z", Year},
		{"-0001-00-00T00:00:00Z", Year},
	}
	for _, test := range tests {
		t.Run(test.time, func(t *testing.T) {
			p, err := parseTime(test.time)
			require.NoError(t, err)
			s := formatTime(p, test.precision)
			assert.Equal(t, test.time, s)
		})
	}
}
