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
	commonsTestDump = "https://gitlab.com/tozd/go/mediawiki/-/raw/main/testdata/commons-testdata-mediainfo.json.bz2"
)

func TestProcessCommonsDumpLatest(t *testing.T) {
	t.Parallel()

	client := retryablehttp.NewClient()
	client.RequestLogHook = func(_ retryablehttp.Logger, req *http.Request, _ int) {
		req.Header.Set("User-Agent", testUserAgent)
	}

	url, errE := mediawiki.LatestCommonsEntitiesRun(context.Background(), client)
	require.NoError(t, errE, "% -+#.1v", errE)

	cacheDir := t.TempDir()
	dumpPath := filepath.Join(cacheDir, path.Base(url))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	entityCounter := int64(0)

	errE = mediawiki.ProcessCommonsEntitiesDump(
		ctx,
		&mediawiki.ProcessDumpConfig{
			URL:    url,
			Path:   dumpPath,
			Client: client,
		},
		func(_ context.Context, a mediawiki.Entity) errors.E {
			atomic.AddInt64(&entityCounter, int64(1))
			cancel()
			b, errE := x.MarshalWithoutEscapeHTML(a)
			if errE != nil {
				return errors.Wrapf(errE, "cannot marshal json: %+v", a)
			}
			var c mediawiki.Entity
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
		assert.Fail(t, "not a context error", "% -+#.1v", errE)
	}
	assert.LessOrEqual(t, int64(1), entityCounter)
}

func TestProcessCommonsDumpExplicit(t *testing.T) {
	t.Parallel()

	client := retryablehttp.NewClient()
	client.RequestLogHook = func(_ retryablehttp.Logger, req *http.Request, _ int) {
		req.Header.Set("User-Agent", testUserAgent)
	}

	cacheDir := t.TempDir()
	dumpPath := filepath.Join(cacheDir, path.Base(commonsTestDump))

	assert.NoFileExists(t, dumpPath)

	entityCounter := int64(0)

	errE := mediawiki.ProcessCommonsEntitiesDump(
		context.Background(),
		&mediawiki.ProcessDumpConfig{
			URL:    commonsTestDump,
			Path:   dumpPath,
			Client: client,
		},
		func(_ context.Context, a mediawiki.Entity) errors.E {
			atomic.AddInt64(&entityCounter, int64(1))
			b, errE := x.MarshalWithoutEscapeHTML(a)
			if errE != nil {
				return errors.Wrapf(errE, "cannot marshal json: %+v", a)
			}
			var c mediawiki.Entity
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
	require.NoError(t, errE, "% -+#.1v", errE)
	assert.Equal(t, int64(10), entityCounter)

	assert.FileExists(t, dumpPath)

	info, err := os.Stat(dumpPath)
	require.NoError(t, err)
	assert.Equal(t, int64(3525), info.Size())

	entityCounter = int64(0)

	errE = mediawiki.ProcessCommonsEntitiesDump(
		context.Background(),
		&mediawiki.ProcessDumpConfig{
			Path: dumpPath,
		},
		func(_ context.Context, a mediawiki.Entity) errors.E {
			atomic.AddInt64(&entityCounter, int64(1))
			b, errE := x.MarshalWithoutEscapeHTML(a)
			if errE != nil {
				return errors.Wrapf(errE, "cannot marshal json: %+v", a)
			}
			var c mediawiki.Entity
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
	require.NoError(t, errE, "% -+#.1v", errE)
	assert.Equal(t, int64(10), entityCounter)
}
