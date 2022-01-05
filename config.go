package mediawiki

import (
	"context"

	"github.com/hashicorp/go-retryablehttp"
	"gitlab.com/tozd/go/x"
)

type ProcessDumpConfig struct {
	URL                    string
	CacheDir               string
	Client                 *retryablehttp.Client
	DecompressionThreads   int
	JSONDecodeThreads      int
	ItemsProcessingThreads int
	UserAgent              string
	Progress               func(context.Context, x.Progress)
}
