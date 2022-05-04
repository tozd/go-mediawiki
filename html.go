package mediawiki

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
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
		return "", errors.WithStack(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body) //nolint:errcheck

	p := pagser.New()

	var data runs
	err = p.ParseReader(&data, resp.Body)
	if err != nil {
		return "", errors.WithStack(err)
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
				return "", errors.WithStack(err)
			}
			defer resp.Body.Close()
			defer io.Copy(ioutil.Discard, resp.Body) //nolint:errcheck
			if resp.StatusCode == http.StatusOK {
				return url, nil
			}
		}
	}

	return "", errors.New("not found")
}
