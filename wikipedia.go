package mediawiki

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-retryablehttp"
	"gitlab.com/tozd/go/errors"
)

// LatestWikipediaRun returns URL of the latest run of Wikimedia Enterprise HTML dump.
// Use "enwiki" for English Wikipedia and namespace 0 for its articles.
func LatestWikipediaRun(ctx context.Context, client *retryablehttp.Client, language string, namespace int) (string, errors.E) {
	format := fmt.Sprintf("https://dumps.wikimedia.org/other/enterprise_html/runs/%%s/%s-NS%d-%%s-ENTERPRISE-HTML.json.tar.gz", language, namespace)
	return latestRun(
		ctx,
		client,
		"https://dumps.wikimedia.org/other/enterprise_html/runs/",
		format,
	)
}

// LatestWikipediaImageMetadataRun returns URL of the latest run of Wikipedia image table dump.
// Use "enwiki" for English Wikipedia.
func LatestWikipediaImageMetadataRun(ctx context.Context, client *retryablehttp.Client, language string) (string, errors.E) {
	format := fmt.Sprintf("https://dumps.wikimedia.org/enwiki/%%s/%s-%%s-image.sql.gz", language)
	return latestRun(
		ctx,
		client,
		fmt.Sprintf("https://dumps.wikimedia.org/%s/", language),
		format,
	)
}

// ProcessWikipediaDump downloads (unless already saves), decompresses, decodes JSON,
// and calls processArticle on every article in a Wikimedia Enterprise HTML dump.
func ProcessWikipediaDump(
	ctx context.Context, config *ProcessDumpConfig,
	processArticle func(context.Context, Article) errors.E,
) errors.E {
	return Process(ctx, &ProcessConfig[Article]{
		URL:                    config.URL,
		Path:                   config.Path,
		Client:                 config.Client,
		DecompressionThreads:   config.DecompressionThreads,
		DecodingThreads:        config.DecodingThreads,
		ItemsProcessingThreads: config.ItemsProcessingThreads,
		Process:                processArticle,
		Progress:               config.Progress,
		FileType:               NDJSON,
		Compression:            GZIPTar,
	})
}
