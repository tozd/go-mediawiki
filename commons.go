package mediawiki

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"gitlab.com/tozd/go/errors"
)

const (
	latestCommonsMediaInfo = "https://dumps.wikimedia.org/commonswiki/entities/latest-mediainfo.json.bz2"
)

// ProcessCommonsDump downloads (unless already cached), decompresses, decodes JSON,
// and calls processEntity on every entity in a Wikimedia Commons entities JSON dump.
func ProcessCommonsDump(
	ctx context.Context, config *ProcessDumpConfig,
	processEntity func(context.Context, Entity) errors.E,
) errors.E {
	if config.Client == nil {
		return errors.New("client is a required configuration option")
	}
	var url, cacheGlob string
	var cacheFilename func(*http.Response) (string, errors.E)
	if config.URL != "" {
		url = config.URL
		filename := path.Base(url)
		cacheGlob = filename
		cacheFilename = func(_ *http.Response) (string, errors.E) {
			return filename, nil
		}
	} else {
		url = latestCommonsMediaInfo
		cacheGlob = "commons-*-mediainfo.json.bz2"
		cacheFilename = func(resp *http.Response) (string, errors.E) {
			lastModifiedStr := resp.Header.Get("Last-Modified")
			if lastModifiedStr == "" {
				return "", errors.Errorf("missing Last-Modified header in response")
			}
			lastModified, err := http.ParseTime(lastModifiedStr)
			if err != nil {
				return "", errors.WithStack(err)
			}
			return fmt.Sprintf("commons-%s-mediainfo.json.bz2", lastModified.UTC().Format("20060102")), nil
		}
	}
	return Process(ctx, &ProcessConfig{
		URL:                    url,
		CacheDir:               config.CacheDir,
		CacheGlob:              cacheGlob,
		CacheFilename:          cacheFilename,
		Client:                 config.Client,
		DecompressionThreads:   config.DecompressionThreads,
		JSONDecodeThreads:      config.JSONDecodeThreads,
		ItemsProcessingThreads: config.ItemsProcessingThreads,
		Process: func(ctx context.Context, i interface{}) errors.E {
			return processEntity(ctx, Entity(*(i.(*commonsEntity))))
		},
		Progress:    config.Progress,
		Item:        &commonsEntity{},
		FileType:    JSONArray,
		Compression: BZIP2,
	})
}
