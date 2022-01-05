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
	wikipediaTestDump = "https://gitlab.com/tozd/go/mediawiki/-/raw/main/testdata/enwiki-NS0-testdata-ENTERPRISE-HTML.json.tar.gz" //nolint:lll
)

func TestProcessWikipediaDumpLatest(t *testing.T) {
	cacheDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	articleCounter := int64(0)

	err := mediawiki.ProcessWikipediaDump(
		ctx,
		&mediawiki.ProcessDumpConfig{ //nolint:exhaustivestruct
			CacheDir:  cacheDir,
			UserAgent: "Unit test user agent (https://gitlab.com/tozd/go/mediawiki)",
		},
		func(_ context.Context, a mediawiki.Article) errors.E {
			atomic.AddInt64(&articleCounter, int64(1))
			cancel()
			return nil
		},
	)
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		assert.Fail(t, "not a context error: %+v", err)
	}
	assert.LessOrEqual(t, int64(1), articleCounter)
}

func TestProcessWikipediaDumpExplicit(t *testing.T) {
	cacheDir := t.TempDir()

	articleCounter := int64(0)

	errE := mediawiki.ProcessWikipediaDump(
		context.Background(),
		&mediawiki.ProcessDumpConfig{ //nolint:exhaustivestruct
			URL:       wikipediaTestDump,
			CacheDir:  cacheDir,
			UserAgent: "Unit test user agent (https://gitlab.com/tozd/go/mediawiki)",
		},
		func(_ context.Context, a mediawiki.Article) errors.E {
			atomic.AddInt64(&articleCounter, int64(1))
			return nil
		},
	)
	assert.NoError(t, errE)
	assert.Equal(t, int64(10), articleCounter)

	dumpPath := filepath.Join(cacheDir, path.Base(wikipediaTestDump))
	assert.FileExists(t, dumpPath)

	info, err := os.Stat(dumpPath)
	assert.NoError(t, err)
	assert.Equal(t, int64(64819), info.Size())
}
