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

	"gitlab.com/tozd/go/mediawiki"
)

const (
	wikidataTestDump = "https://gitlab.com/tozd/go/mediawiki/-/raw/main/testdata/wikidata-testdata-all.json.bz2"
)

func TestProcessWikidataDumpLatest(t *testing.T) {
	client := retryablehttp.NewClient()
	client.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retry int) {
		req.Header.Set("User-Agent", testUserAgent)
	}

	cacheDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	entityCounter := int64(0)

	err := mediawiki.ProcessWikidataDump(
		ctx,
		&mediawiki.ProcessDumpConfig{
			Client:   client,
			CacheDir: cacheDir,
		},
		func(_ context.Context, a mediawiki.Entity) errors.E {
			atomic.AddInt64(&entityCounter, int64(1))
			cancel()
			b, err := json.Marshal(a)
			if err != nil {
				return errors.Wrapf(err, "cannot marshal json: %+v", a)
			}
			var c mediawiki.Entity
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
	assert.LessOrEqual(t, int64(1), entityCounter)
}

func TestProcessWikidataDumpExplicit(t *testing.T) {
	client := retryablehttp.NewClient()
	client.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retry int) {
		req.Header.Set("User-Agent", testUserAgent)
	}

	cacheDir := t.TempDir()

	entityCounter := int64(0)

	errE := mediawiki.ProcessWikidataDump(
		context.Background(),
		&mediawiki.ProcessDumpConfig{
			URL:      wikidataTestDump,
			Client:   client,
			CacheDir: cacheDir,
		},
		func(_ context.Context, a mediawiki.Entity) errors.E {
			atomic.AddInt64(&entityCounter, int64(1))
			b, err := json.Marshal(a)
			if err != nil {
				return errors.Wrapf(err, "cannot marshal json: %+v", a)
			}
			var c mediawiki.Entity
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
	assert.Equal(t, int64(10), entityCounter)

	dumpPath := filepath.Join(cacheDir, path.Base(wikidataTestDump))
	assert.FileExists(t, dumpPath)

	info, err := os.Stat(dumpPath)
	require.NoError(t, err)
	assert.Equal(t, int64(209393), info.Size())
}
