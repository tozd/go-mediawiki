package mediawiki

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/foolin/pagser"
	"github.com/hashicorp/go-retryablehttp"
	"gitlab.com/tozd/go/errors"
)

const (
	wikipediaRuns = "https://dumps.wikimedia.org/other/enterprise_html/runs/"
)

type runs struct {
	Links []string `pagser:"a->eachAttr(href)"`
}

func latestWikipediaRun(client *retryablehttp.Client, userAgent string) (string, errors.E) {
	res, err := client.Get(wikipediaRuns)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer res.Body.Close()

	p := pagser.New()

	var data runs
	err = p.ParseReader(&data, res.Body)
	if err != nil {
		return "", errors.WithStack(err)
	}

	for i, link := range data.Links {
		data.Links[i] = strings.TrimSuffix(link, "/")
	}

	lastDate := data.Links[len(data.Links)-1]

	return fmt.Sprintf("https://dumps.wikimedia.org/other/enterprise_html/runs/%s/enwiki-NS0-%s-ENTERPRISE-HTML.json.tar.gz", lastDate, lastDate), nil //nolint:lll
}

func ProcessWikipediaDump(
	ctx context.Context, config *ProcessDumpConfig,
	processArticle func(context.Context, Article) errors.E,
) errors.E {
	if config.UserAgent == "" {
		return errors.New("user agent is a required configuration option")
	}
	var client *retryablehttp.Client
	if config.Client != nil {
		client = config.Client
	} else {
		client = defaultClient
	}
	var err errors.E
	var url, cacheDir, cacheGlob string
	var cacheFilename func(*http.Response) (string, errors.E)
	if config.URL != "" {
		url = config.URL
	} else {
		url, err = latestWikipediaRun(client, config.UserAgent)
		if err != nil {
			return errors.Wrap(err, "unable to determine the latest English Wikipedia dump run")
		}
	}
	filename := path.Base(url)
	cacheGlob = filename
	cacheFilename = func(_ *http.Response) (string, errors.E) {
		return filename, nil
	}
	if config.CacheDir != "" {
		cacheDir = config.CacheDir
	} else {
		cacheDir = "."
	}
	return Process(ctx, &ProcessConfig{
		URL:                    url,
		CacheDir:               cacheDir,
		CacheGlob:              cacheGlob,
		CacheFilename:          cacheFilename,
		Client:                 client,
		DecompressionThreads:   config.DecompressionThreads,
		JSONDecodeThreads:      config.JSONDecodeThreads,
		ItemsProcessingThreads: config.ItemsProcessingThreads,
		UserAgent:              config.UserAgent,
		Process: func(ctx context.Context, i interface{}) errors.E {
			return processArticle(ctx, *(i.(*Article)))
		},
		Item:        &Article{}, //nolint:exhaustivestruct
		DumpType:    NDJSON,
		Compression: GZIP,
	})
}
