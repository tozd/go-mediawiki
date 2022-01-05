package mediawiki

import (
	"github.com/hashicorp/go-retryablehttp"
)

type ProcessDumpConfig struct {
	DumpURL                string
	CacheDir               string
	Client                 *retryablehttp.Client
	DecompressionThreads   int
	JSONDecodeThreads      int
	ItemsProcessingThreads int
	UserAgent              string
}
