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

func TestProcessWikidataDumpLatest(t *testing.T) {
	cacheDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	entityCounter := int64(0)

	err := mediawiki.ProcessWikidataDump(
		ctx,
		&mediawiki.ProcessDumpConfig{
			CacheDir:  cacheDir,
			UserAgent: "Unit test user agent (https://gitlab.com/tozd/go/mediawiki)",
		},
		func(a mediawiki.Entity) errors.E {
			atomic.AddInt64(&entityCounter, int64(1))
			cancel()
			return nil
		},
	)
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		assert.Fail(t, "not a context error: %+v", err)
	}
	assert.Equal(t, int64(1), entityCounter)
}
