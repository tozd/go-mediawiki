package mediawiki

import (
	"context"

	"github.com/hashicorp/go-retryablehttp"
	"gitlab.com/tozd/go/errors"
)

// LatestWikidataEntitiesRun returns URL of the latest run of Wikidata entities JSON dump.
func LatestWikidataEntitiesRun(client *retryablehttp.Client) (string, errors.E) {
	return latestRun(
		client,
		"https://dumps.wikimedia.org/wikidatawiki/entities/",
		"https://dumps.wikimedia.org/wikidatawiki/entities/%s/wikidata-%s-all.json.bz2",
	)
}

// ProcessWikidataDump downloads (unless already saves), decompresses, decodes JSON,
// and calls processEntity on every entity in a Wikidata entities JSON dump.
func ProcessWikidataDump(
	ctx context.Context, config *ProcessDumpConfig,
	processEntity func(context.Context, Entity) errors.E,
) errors.E {
	return Process(ctx, &ProcessConfig{
		URL:                    config.URL,
		Path:                   config.Path,
		Client:                 config.Client,
		DecompressionThreads:   config.DecompressionThreads,
		DecodingThreads:        config.DecodingThreads,
		ItemsProcessingThreads: config.ItemsProcessingThreads,
		Process: func(ctx context.Context, i interface{}) errors.E {
			return processEntity(ctx, *(i.(*Entity)))
		},
		Progress:    config.Progress,
		Item:        &Entity{},
		FileType:    JSONArray,
		Compression: BZIP2,
	})
}
