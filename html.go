package mediawiki

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/foolin/pagser"
	"github.com/hashicorp/go-retryablehttp"
	"gitlab.com/tozd/go/errors"
)

var runRegex = regexp.MustCompile(`^(\d{8})/$`)

// Runs is used for parsing links of dump runs.
type runs struct {
	Links []string `pagser:"a->eachAttr(href)"`
}

func latestRun(ctx context.Context, client *retryablehttp.Client, runURL, fileFormat string) (string, errors.E) {
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, runURL, nil)
	if err != nil {
		errE := errors.WithMessage(err, "new request")
		errors.Details(errE)["url"] = runURL
		return "", errE
	}
	resp, err := client.Do(req)
	if err != nil {
		errE := errors.WithMessage(err, "do")
		errors.Details(errE)["url"] = runURL
		return "", errE
	}
	defer resp.Body.Close()              //nolint:errcheck
	defer io.Copy(io.Discard, resp.Body) //nolint:errcheck

	p := pagser.New()

	var data runs
	err = p.ParseReader(&data, resp.Body)
	if err != nil {
		errE := errors.WithMessage(err, "parse")
		errors.Details(errE)["url"] = runURL
		return "", errE
	}

	// We start with the last link.
	for i := len(data.Links) - 1; i >= 0; i-- {
		link := data.Links[i]
		match := runRegex.FindStringSubmatch(link)
		if match != nil {
			lastDate := match[1]
			url := fmt.Sprintf(fileFormat, lastDate, lastDate)

			// It can happen that the file is missing in the dump directory. So we check.
			resp, err := client.Head(url)
			if err != nil {
				errE := errors.WithMessage(err, "head")
				errors.Details(errE)["url"] = url
				return "", errE
			}
			defer resp.Body.Close()              //nolint:errcheck
			defer io.Copy(io.Discard, resp.Body) //nolint:errcheck
			if resp.StatusCode == http.StatusOK {
				return url, nil
			}
		}
	}

	return "", errors.WithDetails(ErrNotFound, "url", runURL)
}
