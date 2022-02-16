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

	url, errE := mediawiki.LatestWikipediaRun(client, "enwiki", 0)
	require.NoError(t, errE)

	cacheDir := t.TempDir()
	dumpPath := filepath.Join(cacheDir, path.Base(url))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	articleCounter := int64(0)

	errE = mediawiki.ProcessWikipediaDump(
		ctx,
		&mediawiki.ProcessDumpConfig{
			URL:    url,
			Path:   dumpPath,
			Client: client,
		},
		func(_ context.Context, a mediawiki.Article) errors.E {
			atomic.AddInt64(&articleCounter, int64(1))
			cancel()
			b, errE := x.MarshalWithoutEscapeHTML(a) //nolint:govet
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
	if !errors.Is(errE, context.DeadlineExceeded) && !errors.Is(errE, context.Canceled) {
		assert.Fail(t, "not a context error: %+v", errE)
	}
	assert.LessOrEqual(t, int64(1), articleCounter)
}

func TestProcessWikipediaDumpExplicit(t *testing.T) {
	client := retryablehttp.NewClient()
	client.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retry int) {
		req.Header.Set("User-Agent", testUserAgent)
	}

	cacheDir := t.TempDir()
	dumpPath := filepath.Join(cacheDir, path.Base(wikipediaTestDump))

	articleCounter := int64(0)

	errE := mediawiki.ProcessWikipediaDump(
		context.Background(),
		&mediawiki.ProcessDumpConfig{
			URL:    wikipediaTestDump,
			Path:   dumpPath,
			Client: client,
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

	assert.FileExists(t, dumpPath)

	info, err := os.Stat(dumpPath)
	require.NoError(t, err)
	assert.Equal(t, int64(64819), info.Size())

	articleCounter = int64(0)

	errE = mediawiki.ProcessWikipediaDump(
		context.Background(),
		&mediawiki.ProcessDumpConfig{
			Path: dumpPath,
		},
		func(_ context.Context, a mediawiki.Article) errors.E {
			atomic.AddInt64(&articleCounter, int64(1))
			b, errE := x.MarshalWithoutEscapeHTML(a) //nolint:govet
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
}
