// Command update runs the update of the test data.
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-retryablehttp"
	"gitlab.com/tozd/go/errors"

	"gitlab.com/tozd/go/mediawiki"
)

const (
	maxEntries = 10
	userAgent  = "go-mediawiki user agent (https://gitlab.com/tozd/go/mediawiki)"
)

func CommonsEntities(ctx context.Context, client *retryablehttp.Client) errors.E {
	url, errE := mediawiki.LatestCommonsEntitiesRun(ctx, client)
	if errE != nil {
		return errE
	}

	return entities(ctx, client, url, "commons-testdata-mediainfo.json")
}

func WikidataEntities(ctx context.Context, client *retryablehttp.Client) errors.E {
	url, errE := mediawiki.LatestWikidataEntitiesRun(ctx, client)
	if errE != nil {
		return errE
	}

	return entities(ctx, client, url, "wikidata-testdata-all.json")
}

func entities(ctx context.Context, client *retryablehttp.Client, url, output string) errors.E {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	file, err := os.Create(filepath.Clean(output))
	if err != nil {
		return errors.WithStack(err)
	}
	defer file.Close() //nolint:errcheck

	entriesCount := 0
	_, _ = file.WriteString("[\n")

	errE := mediawiki.Process(ctx, &mediawiki.ProcessConfig[json.RawMessage]{ //nolint:exhaustruct
		URL:                    url,
		Client:                 client,
		DecompressionThreads:   1,
		DecodingThreads:        1,
		ItemsProcessingThreads: 1,
		Process: func(_ context.Context, j json.RawMessage) errors.E {
			if entriesCount > maxEntries {
				cancel()
				return nil
			}
			entriesCount++

			_, _ = file.Write(j)
			if entriesCount < maxEntries {
				_, _ = file.WriteString(",")
			}
			_, _ = file.WriteString("\n")
			if entriesCount == maxEntries {
				cancel()
			}
			return nil
		},
		FileType:    mediawiki.JSONArray,
		Compression: mediawiki.BZIP2,
	})
	if errE != nil && !errors.Is(errE, context.Canceled) {
		return errE
	}
	_, _ = file.WriteString("]\n")

	return nil
}

func Wikipedia(ctx context.Context, client *retryablehttp.Client) errors.E {
	url, errE := mediawiki.LatestWikipediaRun(ctx, client, "enwiki", 0)
	if errE != nil {
		return errE
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	file, err := os.Create("enwiki_namespace_0_0.ndjson")
	if err != nil {
		return errors.WithStack(err)
	}
	defer file.Close() //nolint:errcheck

	entriesCount := 0
	errE = mediawiki.Process(ctx, &mediawiki.ProcessConfig[json.RawMessage]{ //nolint:exhaustruct
		URL:                    url,
		Client:                 client,
		DecompressionThreads:   1,
		DecodingThreads:        1,
		ItemsProcessingThreads: 1,
		Process: func(_ context.Context, j json.RawMessage) errors.E {
			if entriesCount > maxEntries {
				cancel()
				return nil
			}
			entriesCount++

			_, _ = file.Write(j)
			_, _ = file.WriteString("\n")
			if entriesCount == maxEntries {
				cancel()
			}
			return nil
		},
		FileType:    mediawiki.NDJSON,
		Compression: mediawiki.GZIPTar,
	})
	if errE != nil && !errors.Is(errE, context.Canceled) {
		return errE
	}

	return nil
}

func main() {
	client := retryablehttp.NewClient()
	client.RequestLogHook = func(_ retryablehttp.Logger, req *http.Request, _ int) {
		req.Header.Set("User-Agent", userAgent)
	}

	errE := CommonsEntities(context.Background(), client)
	if errE != nil {
		log.Fatalf("% -+#.1v", errE)
	}

	errE = WikidataEntities(context.Background(), client)
	if errE != nil {
		log.Fatalf("% -+#.1v", errE)
	}

	errE = Wikipedia(context.Background(), client)
	if errE != nil {
		log.Fatalf("% -+#.1v", errE)
	}
}
