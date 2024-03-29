package mediawiki

import (
	"context"
	"fmt"
	"strings"

	"github.com/elliotchance/phpserialize"
	"github.com/hashicorp/go-retryablehttp"
	"gitlab.com/tozd/go/errors"
	"gitlab.com/tozd/go/x"
)

// LatestCommonsEntitiesRun returns URL of the latest run of Wikimedia Commons entities JSON dump.
func LatestCommonsEntitiesRun(ctx context.Context, client *retryablehttp.Client) (string, errors.E) {
	return latestRun(
		ctx,
		client,
		"https://dumps.wikimedia.org/commonswiki/entities/",
		"https://dumps.wikimedia.org/commonswiki/entities/%s/commons-%s-mediainfo.json.bz2",
	)
}

// ProcessCommonsEntitiesDump downloads (unless already saved), decompresses, decodes JSON,
// and calls processEntity on every entity in a Wikimedia Commons entities JSON dump.
func ProcessCommonsEntitiesDump(
	ctx context.Context, config *ProcessDumpConfig,
	processEntity func(context.Context, Entity) errors.E,
) errors.E {
	return Process(ctx, &ProcessConfig[commonsEntity]{
		URL:                    config.URL,
		Path:                   config.Path,
		Client:                 config.Client,
		DecompressionThreads:   config.DecompressionThreads,
		DecodingThreads:        config.DecodingThreads,
		ItemsProcessingThreads: config.ItemsProcessingThreads,
		Process: func(ctx context.Context, i commonsEntity) errors.E {
			return processEntity(ctx, Entity(i))
		},
		Progress:    config.Progress,
		FileType:    JSONArray,
		Compression: BZIP2,
	})
}

func convertToStringMaps(value interface{}) interface{} {
	switch v := value.(type) {
	case []interface{}:
		for i, el := range v {
			v[i] = convertToStringMaps(el)
		}
	case map[interface{}]interface{}:
		return convertToStringMapsMap(v)
	}
	return value
}

func convertToStringMapsMap(m map[interface{}]interface{}) map[string]interface{} {
	out := make(map[string]interface{})

	for key, value := range m {
		out[fmt.Sprint(key)] = convertToStringMaps(value)
	}

	return out
}

// LatestCommonsImageMetadataRun returns URL of the latest run of Wikimedia Commons image table dump.
func LatestCommonsImageMetadataRun(ctx context.Context, client *retryablehttp.Client) (string, errors.E) {
	return latestRun(
		ctx,
		client,
		"https://dumps.wikimedia.org/commonswiki/",
		"https://dumps.wikimedia.org/commonswiki/%s/commonswiki-%s-image.sql.gz",
	)
}

// DecodeImageMetadata decodes image and other uploaded files metadata column in
// image table. See: https://www.mediawiki.org/wiki/Manual:Image_table
func DecodeImageMetadata(metadata interface{}) (map[string]interface{}, errors.E) {
	if metadata == "" || metadata == "0" || metadata == "-1" {
		return make(map[string]interface{}), nil
	}

	m, ok := metadata.(string)
	if !ok {
		errE := errors.WithMessage(ErrUnexpectedType, "metadata")
		errors.Details(errE)["expected"] = "string" //nolint:goconst
		errors.Details(errE)["type"] = fmt.Sprintf("%T", metadata)
		return nil, errE
	}
	if strings.HasPrefix(m, "{") {
		var d map[string]interface{}
		errE := x.Unmarshal([]byte(m), &d)
		if errE != nil {
			return nil, errE
		}
		return d, nil
	}

	var d map[interface{}]interface{}
	err := phpserialize.Unmarshal([]byte(m), &d)
	if err != nil {
		return nil, errors.WithMessage(err, "phpserialize unmarshal")
	}
	return convertToStringMapsMap(d), nil
}
