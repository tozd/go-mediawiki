package mediawiki_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"gitlab.com/tozd/go/errors"

	"gitlab.com/tozd/go/mediawiki"
)

const (
	testFilesBaseURL = "https://gitlab.com/tozd/go/mediawiki/-/raw/main/testdata/"
	testUserAgent    = "Unit test user agent (https://gitlab.com/tozd/go/mediawiki)"
)

func TestCompression(t *testing.T) {
	client := retryablehttp.NewClient()

	tests := []struct {
		name        string
		compression mediawiki.Compression
		dumpType    mediawiki.FileType
		items       int
	}{
		{"enwiki-NS0-testdata-ENTERPRISE-HTML.json.tar", mediawiki.Tar, mediawiki.NDJSON, 10},
		{"enwiki-NS0-testdata-ENTERPRISE-HTML.json.tar.bz2", mediawiki.BZIP2Tar, mediawiki.NDJSON, 10},
		{"enwiki-NS0-testdata-ENTERPRISE-HTML.json.tar.gz", mediawiki.GZIPTar, mediawiki.NDJSON, 10},
		{"wikidata-testdata-all.json", mediawiki.NoCompression, mediawiki.JSONArray, 10},
		{"wikidata-testdata-all.json.bz2", mediawiki.BZIP2, mediawiki.JSONArray, 10},
		{"wikidata-testdata-all.json.gz", mediawiki.GZIP, mediawiki.JSONArray, 10},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cacheDir := t.TempDir()

			itemCounter := int64(0)

			err := mediawiki.Process(context.Background(), &mediawiki.ProcessConfig{ //nolint:exhaustivestruct
				URL:       testFilesBaseURL + test.name,
				CacheDir:  cacheDir,
				CacheGlob: test.name,
				CacheFilename: func(_ *http.Response) (string, errors.E) {
					return test.name, nil
				},
				Client:    client,
				UserAgent: testUserAgent,
				Process: func(ctx context.Context, i interface{}) errors.E {
					atomic.AddInt64(&itemCounter, int64(1))
					return nil
				},
				Item:        new(interface{}),
				FileType:    test.dumpType,
				Compression: test.compression,
			})
			assert.NoError(t, err)
			assert.Equal(t, int64(test.items), itemCounter)
		})
	}
}
