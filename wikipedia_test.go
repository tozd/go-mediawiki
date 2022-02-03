package mediawiki_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/tozd/go/errors"
	"gitlab.com/tozd/go/x"

	"gitlab.com/tozd/go/mediawiki"
)

const (
	wikipediaTestDump = "https://gitlab.com/tozd/go/mediawiki/-/raw/main/testdata/enwiki-NS0-testdata-ENTERPRISE-HTML.json.tar.gz"
)

func TestProcessWikipediaDumpLatest(t *testing.T) {
	client := retryablehttp.NewClient()
	client.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retry int) {
		req.Header.Set("User-Agent", testUserAgent)
	}

	cacheDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	articleCounter := int64(0)

	err := mediawiki.ProcessWikipediaDump(
		ctx,
		&mediawiki.ProcessDumpConfig{
			Client:   client,
			CacheDir: cacheDir,
		},
		func(_ context.Context, a mediawiki.Article) errors.E {
			atomic.AddInt64(&articleCounter, int64(1))
			cancel()
			b, errE := x.MarshalWithoutEscapeHTML(a)
			if errE != nil {
				return errors.Wrapf(errE, "cannot marshal json: %+v", a)
			}
			var c mediawiki.Article
			err := json.Unmarshal(b, &c)
			if err != nil {
				return errors.Wrapf(err, "cannot unmarshal json: %s", string(b))
			}
			d, err := x.MarshalWithoutEscapeHTML(c)
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
	client := retryablehttp.NewClient()
	client.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retry int) {
		req.Header.Set("User-Agent", testUserAgent)
	}

	cacheDir := t.TempDir()

	articleCounter := int64(0)

	errE := mediawiki.ProcessWikipediaDump(
		context.Background(),
		&mediawiki.ProcessDumpConfig{
			URL:      wikipediaTestDump,
			Client:   client,
			CacheDir: cacheDir,
		},
		func(_ context.Context, a mediawiki.Article) errors.E {
			atomic.AddInt64(&articleCounter, int64(1))
			b, errE := x.MarshalWithoutEscapeHTML(a)
			if errE != nil {
				return errors.Wrapf(errE, "cannot marshal json: %+v", a)
			}
			var c mediawiki.Article
			err := json.Unmarshal(b, &c)
			if err != nil {
				return errors.Wrapf(err, "cannot unmarshal json: %s", string(b))
			}
			d, err := x.MarshalWithoutEscapeHTML(c)
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
