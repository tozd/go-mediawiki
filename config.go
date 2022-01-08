package mediawiki

import (
	"context"

	"github.com/hashicorp/go-retryablehttp"
	"gitlab.com/tozd/go/x"
)

// ProcessDumpConfig is a configuration for high-level Process*Dump functions.
//
// URL and UserAgent are required.
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
