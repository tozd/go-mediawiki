package mediawiki

import (
	"archive/tar"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/cosnicolaou/pbzip2"
	"github.com/hashicorp/go-retryablehttp"
	gzip "github.com/klauspost/pgzip"
	"gitlab.com/tozd/go/errors"
	"gitlab.com/tozd/go/x"
)

const (
	progressPrintRate = 30 * time.Second
	staleReadTimeout  = 60 * time.Second
)

type DumpType int

const (
	JSONArray DumpType = iota
	NDJSON
)

type Compression int

const (
	BZIP2 Compression = iota
	GZIP
)

type ProcessConfig struct {
	URL                    string
	CacheDir               string
	CacheGlob              string
	CacheFilename          func(*http.Response) (string, errors.E)
	Client                 *retryablehttp.Client
	DecompressionThreads   int
	JSONDecodeThreads      int
	ItemsProcessingThreads int
	UserAgent              string
	Process                func(context.Context, interface{}) errors.E
	Progress               func(context.Context, x.Progress)
	Item                   interface{}
	DumpType               DumpType
	Compression            Compression
}

func getDumpJSONs(
	ctx context.Context, config *ProcessConfig, wg *sync.WaitGroup,
	output chan<- json.RawMessage, errs chan<- errors.E,
) {
	defer wg.Done()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	matches, err := filepath.Glob(filepath.Join(config.CacheDir, config.CacheGlob))
	if err != nil {
		errs <- errors.WithStack(err)
		return
	}

	var compressedReader io.Reader
	var compressedSize int64
	var timer *time.Timer

	if len(matches) == 1 {
		// If we file is already cached, we use it.
		compressedFile, err := os.Open(matches[0]) //nolint:govet
		if err != nil {
			errs <- errors.WithStack(err)
			return
		}
		defer compressedFile.Close()
		compressedReader = compressedFile
		compressedSize, err = compressedFile.Seek(0, io.SeekEnd)
		if err != nil {
			errs <- errors.WithStack(err)
			return
		}
		_, err = compressedFile.Seek(0, io.SeekStart)
		if err != nil {
			errs <- errors.WithStack(err)
			return
		}
	} else if len(matches) > 1 {
		errs <- errors.Errorf(`too many cached files matching pattern "%s" in "%s": %d`, config.CacheGlob, config.CacheDir, len(matches)) //nolint:lll
		return
	} else {
		// Otherwise we download the file and cache it.
		req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, config.URL, nil) //nolint:govet
		if err != nil {
			errs <- errors.WithStack(err)
			return
		}
		if config.UserAgent != "" {
			req.Header.Set("User-Agent", config.UserAgent)
		}
		downloadReader, errE := x.NewRetryableResponse(config.Client, req)
		if errE != nil {
			errs <- errE
			return
		}
		defer downloadReader.Close()
		compressedSize = downloadReader.Size()
		filename, errE := config.CacheFilename(downloadReader.Response)
		if errE != nil {
			errs <- errE
			return
		}
		p := filepath.Join(config.CacheDir, filename)
		compressedFile, err := os.Create(p)
		if err != nil {
			errs <- errors.WithStack(err)
			return
		}
		defer func() {
			info, err := os.Stat(p)
			if err != nil || downloadReader.Size() != info.Size() {
				// Incomplete file. Delete.
				_ = os.Remove(p)
			}
		}()
		defer compressedFile.Close()
		compressedReader = io.TeeReader(downloadReader, compressedFile)
		// TODO: Better error message when canceled.
		//       See: https://github.com/golang/go/issues/26356
		timer = time.AfterFunc(staleReadTimeout, cancel)
		defer timer.Stop()
	}

	countingReader := &x.CountingReader{Reader: compressedReader}
	ticker := x.NewTicker(ctx, countingReader, compressedSize, progressPrintRate)
	defer ticker.Stop()
	go func() {
		compressedRead := int64(0)
		for progress := range ticker.C {
			if timer != nil && compressedRead != progress.Count {
				timer.Reset(staleReadTimeout)
				compressedRead = progress.Count
			}
			if config.Progress != nil {
				config.Progress(ctx, progress)
			}
		}
	}()

	var decompressedReader io.Reader
	switch config.Compression {
	case BZIP2:
		decompressedReader = pbzip2.NewReader(
			ctx, countingReader,
			pbzip2.DecompressionOptions(
				pbzip2.BZConcurrency(config.DecompressionThreads),
			),
		)
	case GZIP:
		gzipReader, err := gzip.NewReader(countingReader) //nolint:govet
		if err != nil {
			errs <- errors.WithStack(err)
			return
		}
		defer gzipReader.Close()

		decompressedReader = tar.NewReader(gzipReader)
	default:
		panic(errors.Errorf("unknown compression: %d", config.Compression))
	}

	for {
		if config.Compression == GZIP {
			// Go to the first or next file in gzip/tar.
			_, err = decompressedReader.(*tar.Reader).Next()
			if err != nil {
				// When there are no more files in gzip/tar, Next returns io.EOF.
				if !errors.Is(err, io.EOF) {
					errs <- errors.WithStack(err)
				}
				return
			}
		}

		decoder := json.NewDecoder(decompressedReader)

		if config.DumpType == JSONArray {
			// Read open bracket.
			_, err = decoder.Token()
			if err != nil {
				errs <- errors.WithStack(err)
				return
			}
		}

		for decoder.More() {
			var raw json.RawMessage
			err = decoder.Decode(&raw)
			if err != nil {
				errs <- errors.WithStack(err)
				return
			}
			if err = ctx.Err(); err != nil {
				errs <- errors.WithStack(err)
				return
			}
			output <- raw
		}

		if config.DumpType == JSONArray {
			// Read closing bracket.
			_, err = decoder.Token()
			if err != nil {
				errs <- errors.WithStack(err)
				return
			}
		}

		if config.Compression != GZIP {
			// Only gzip/tar has multiple files.
			break
		}
	}
}

func decodeJSONs(
	ctx context.Context, config *ProcessConfig, wg *sync.WaitGroup,
	input <-chan json.RawMessage, output chan<- interface{}, errs chan<- errors.E,
) {
	defer wg.Done()

	itemType := reflect.TypeOf(config.Item).Elem()

	for {
		select {
		case raw, ok := <-input:
			if !ok {
				return
			}
			e := reflect.New(itemType).Interface()
			errE := x.UnmarshalWithoutUnknownFields(raw, &e)
			if errE != nil {
				errs <- errors.Wrapf(errE, "cannot decode json: %s", raw)
				return
			}
			if err := ctx.Err(); err != nil {
				errs <- errors.WithStack(err)
				return
			}
			output <- e
		case <-ctx.Done():
			errs <- errors.WithStack(ctx.Err())
			return
		}
	}
}

func processItems(
	ctx context.Context, config *ProcessConfig, wg *sync.WaitGroup,
	input <-chan interface{}, errs chan<- errors.E,
) {
	defer wg.Done()

	for {
		select {
		case i, ok := <-input:
			if !ok {
				return
			}
			err := config.Process(ctx, i)
			if err != nil {
				errs <- err
				return
			}
		case <-ctx.Done():
			errs <- errors.WithStack(ctx.Err())
			return
		}
	}
}

func Process(ctx context.Context, config *ProcessConfig) errors.E {
	if config.DecompressionThreads == 0 {
		config.DecompressionThreads = runtime.GOMAXPROCS(0)
	}
	if config.JSONDecodeThreads == 0 {
		config.JSONDecodeThreads = runtime.GOMAXPROCS(0)
	}
	if config.ItemsProcessingThreads == 0 {
		config.ItemsProcessingThreads = runtime.GOMAXPROCS(0)
	}

	// We call cancel on SIGINT or SIGTERM signal and on any
	// error from goroutines. The expectation is that all
	// goroutines return soon afterwards.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// mainWg counts groups of same goroutines.
	var mainWg sync.WaitGroup
	// mainWgChan is closed when mainWg is done.
	mainWgChan := make(chan struct{})

	// Call cancel on SIGINT or SIGTERM signal.
	go func() {
		c := make(chan os.Signal, 1)
		defer close(c)

		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(c)

		// We wait for a signal or that the context is canceled
		// or that all goroutines are done.
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		case <-mainWgChan:
		}
	}()

	errs := make(chan errors.E, 1+config.JSONDecodeThreads+config.ItemsProcessingThreads)
	defer close(errs)

	jsons := make(chan json.RawMessage, config.JSONDecodeThreads)
	items := make(chan interface{}, config.ItemsProcessingThreads)

	var getDumpJSONsWg sync.WaitGroup
	mainWg.Add(1)
	getDumpJSONsWg.Add(1)
	go getDumpJSONs(ctx, config, &getDumpJSONsWg, jsons, errs)
	go func() {
		getDumpJSONsWg.Wait()
		mainWg.Done()
		// All goroutines using jsons channel as output are done,
		// we can close the channel.
		close(jsons)
	}()

	var decodeJSONsWg sync.WaitGroup
	mainWg.Add(1)
	for w := 0; w < config.JSONDecodeThreads; w++ {
		decodeJSONsWg.Add(1)
		go decodeJSONs(ctx, config, &decodeJSONsWg, jsons, items, errs)
	}
	go func() {
		decodeJSONsWg.Wait()
		mainWg.Done()
		// All goroutines using items channel as output are done,
		// we can close the channel.
		close(items)
	}()

	var processItemWg sync.WaitGroup
	mainWg.Add(1)
	for w := 0; w < config.ItemsProcessingThreads; w++ {
		processItemWg.Add(1)
		go processItems(ctx, config, &processItemWg, items, errs)
	}
	go func() {
		processItemWg.Wait()
		mainWg.Done()
	}()

	// When mainWg is done, we close mainWgChan.
	// This means that all goroutines are done.
	go func() {
		mainWg.Wait()
		close(mainWgChan)
	}()

	allErrors := []errors.E{}
WAIT:
	for {
		// We cancel the context on any error, but we also store it.
		// We also wait for all goroutines to return. The expectation
		// is that they return all when they are all successful, or
		// when there was an error and we canceled the context.
		select {
		case err := <-errs:
			allErrors = append(allErrors, err)
			cancel()
		case <-mainWgChan:
			break WAIT
		}
	}

	if len(allErrors) > 0 {
		// If there is any non-context-canceled error, return it.
		// TODO: What if there are multiple such errors?
		for _, err := range allErrors {
			if !errors.Is(err, context.Canceled) {
				return err
			}
		}

		// Otherwise return any error, i.e., the first.
		return allErrors[0]
	}

	return nil
}
