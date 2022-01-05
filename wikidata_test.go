package mediawiki_test

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/tozd/go/errors"

	"gitlab.com/tozd/go/mediawiki"
)

const (
	wikidataTestDump = "https://gitlab.com/tozd/go/mediawiki/-/raw/main/testdata/wikidata-testdata-all.json.bz2"
)

func TestProcessWikidataDumpLatest(t *testing.T) {
	cacheDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	entityCounter := int64(0)

	err := mediawiki.ProcessWikidataDump(
		ctx,
		&mediawiki.ProcessDumpConfig{ //nolint:exhaustivestruct
			CacheDir:  cacheDir,
			UserAgent: "Unit test user agent (https://gitlab.com/tozd/go/mediawiki)",
		},
		func(_ context.Context, a mediawiki.Entity) errors.E {
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

func TestProcessWikidataDumpExplicit(t *testing.T) {
	cacheDir := t.TempDir()

	entityCounter := int64(0)

	errE := mediawiki.ProcessWikidataDump(
		context.Background(),
		&mediawiki.ProcessDumpConfig{ //nolint:exhaustivestruct
			URL:       wikidataTestDump,
			CacheDir:  cacheDir,
			UserAgent: "Unit test user agent (https://gitlab.com/tozd/go/mediawiki)",
		},
		func(_ context.Context, a mediawiki.Entity) errors.E {
			atomic.AddInt64(&entityCounter, int64(1))
			return nil
		},
	)
	assert.NoError(t, errE)
	assert.Equal(t, int64(10), entityCounter)

	dumpPath := filepath.Join(cacheDir, path.Base(wikidataTestDump))
	assert.FileExists(t, dumpPath)

	info, err := os.Stat(dumpPath)
	assert.NoError(t, err)
	assert.Equal(t, int64(209393), info.Size())
}
