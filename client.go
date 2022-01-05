package mediawiki

import (
	"github.com/hashicorp/go-retryablehttp"
)

var defaultClient = retryablehttp.NewClient()
