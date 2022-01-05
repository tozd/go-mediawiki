package mediawiki_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/tozd/go/errors"
	"gitlab.com/tozd/go/mediawiki"
)

func TestProcessWikipediaDumpLatest(t *testing.T) {
	cacheDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	articleCounter := int64(0)

	err := mediawiki.ProcessWikipediaDump(
		ctx,
		&mediawiki.ProcessDumpConfig{
			CacheDir:  cacheDir,
			UserAgent: "Unit test user agent (https://gitlab.com/tozd/go/mediawiki)",
		},
		func(a mediawiki.Article) errors.E {
			atomic.AddInt64(&articleCounter, int64(1))
			cancel()
			return nil
		},
	)
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		assert.Fail(t, "not a context error: %+v", err)
	}
	assert.Equal(t, int64(1), articleCounter)
}
