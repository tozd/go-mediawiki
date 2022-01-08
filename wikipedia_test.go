package mediawiki_test

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			UserAgent: testUserAgent,
		},
		func(_ context.Context, a mediawiki.Article) errors.E {
			atomic.AddInt64(&articleCounter, int64(1))
			cancel()
			b, err := json.Marshal(a)
			if err != nil {
				return errors.Wrapf(err, "cannot marshal json: %+v", a)
			}
			var c mediawiki.Article
			err = json.Unmarshal(b, &c)
			if err != nil {
				return errors.Wrapf(err, "cannot unmarshal json: %s", string(b))
			}
			d, err := json.Marshal(c)
			if err != nil {
				return errors.Wrapf(err, "cannot marshal json again: %+v", c)
			}
			bStr := string(b)
			dStr := string(d)
			// We have to use JSONEq instead of Equal so that empty slice is equal to nil slice.
			assert.JSONEq(t, bStr, dStr)
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
			UserAgent: testUserAgent,
		},
		func(_ context.Context, a mediawiki.Article) errors.E {
			atomic.AddInt64(&articleCounter, int64(1))
			b, err := json.Marshal(a)
			if err != nil {
				return errors.Wrapf(err, "cannot marshal json: %+v", a)
			}
			var c mediawiki.Article
			err = json.Unmarshal(b, &c)
			if err != nil {
				return errors.Wrapf(err, "cannot unmarshal json: %s", string(b))
			}
			d, err := json.Marshal(c)
			if err != nil {
				return errors.Wrapf(err, "cannot marshal json again: %+v", c)
			}
			bStr := string(b)
			dStr := string(d)
			// We have to use JSONEq instead of Equal so that empty slice is equal to nil slice.
			assert.JSONEq(t, bStr, dStr)
			return nil
		},
	)
	assert.NoError(t, errE)
	assert.Equal(t, int64(10), articleCounter)

	dumpPath := filepath.Join(cacheDir, path.Base(wikipediaTestDump))
	assert.FileExists(t, dumpPath)

	info, err := os.Stat(dumpPath)
	require.NoError(t, err)
	assert.Equal(t, int64(64819), info.Size())
}
