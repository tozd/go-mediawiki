package mediawiki

import (
	"context"

	"github.com/hashicorp/go-retryablehttp"
	"gitlab.com/tozd/go/x"
)

// ProcessDumpConfig is a configuration for high-level Process*Dump functions.
//
// Client is required.
//
// Client should set User-Agent header with contact information, e.g.:
//
//     client := retryablehttp.NewClient()
//     client.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retry int) {
//     	req.Header.Set("User-Agent", "My bot (user@example.com)")
//     }
type ProcessDumpConfig struct {
	URL                    string
	CacheDir               string
	Client                 *retryablehttp.Client
	DecompressionThreads   int
	DecodingThreads        int
	ItemsProcessingThreads int
	Progress               func(context.Context, x.Progress)
}
