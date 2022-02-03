package mediawiki

import (
	"archive/tar"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
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

type FileType int

const (
	JSONArray FileType = iota
	NDJSON
	SQLDump
)

type Compression int

const (
	NoCompression Compression = iota
	Tar
	BZIP2
	BZIP2Tar
	GZIP
	GZIPTar
)

// ProcessConfig is a configuration for low-level Process function.
//
// URL, Client, Process, Item, FileType, and Compression are required.
//
// Client should set User-Agent header with contact information, e.g.:
//
//     client := retryablehttp.NewClient()
//     client.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retry int) {
//     	req.Header.Set("User-Agent", "My bot (user@example.com)")
//     }
type ProcessConfig struct {
	URL                    string
	CacheDir               string
	CacheGlob              string
	CacheFilename          func(*http.Response) (string, errors.E)
	Client                 *retryablehttp.Client
	DecompressionThreads   int
	DecodingThreads        int
	ItemsProcessingThreads int
	Process                func(context.Context, interface{}) errors.E
	Progress               func(context.Context, x.Progress)
	Item                   interface{}
	FileType               FileType
	Compression            Compression
}

func getFileRows(
	ctx context.Context, config *ProcessConfig, wg *sync.WaitGroup,
	output chan<- []byte, errs chan<- errors.E,
) {
	defer wg.Done()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var matches []string
	var err error
	if config.CacheDir != "" && config.CacheGlob != "" {
		matches, err = filepath.Glob(filepath.Join(config.CacheDir, config.CacheGlob))
		if err != nil {
			errs <- errors.WithStack(err)
			return
		}
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
		errs <- errors.Errorf(`too many cached files matching pattern "%s" in "%s": %d`, config.CacheGlob, config.CacheDir, len(matches))
		return
	} else {
		// Otherwise we download the file and cache it.
		req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, config.URL, nil) //nolint:govet
		if err != nil {
			errs <- errors.WithStack(err)
			return
		}
		downloadReader, errE := x.NewRetryableResponse(config.Client, req)
		if errE != nil {
			errs <- errE
			return
		}
		defer downloadReader.Close()
		compressedSize = downloadReader.Size()
		if config.CacheDir != "" && config.CacheFilename != nil {
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
		} else {
			compressedReader = downloadReader
		}
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
	case BZIP2, BZIP2Tar:
		decompressedReader = pbzip2.NewReader(
			ctx, countingReader,
			pbzip2.DecompressionOptions(
				pbzip2.BZConcurrency(config.DecompressionThreads),
			),
		)
	case GZIP, GZIPTar:
		gzipReader, err := gzip.NewReader(countingReader) //nolint:govet
		if err != nil {
			errs <- errors.WithStack(err)
			return
		}
		defer gzipReader.Close()
		decompressedReader = gzipReader
	case NoCompression, Tar:
		decompressedReader = countingReader
	default:
		panic(errors.Errorf("unknown compression: %d", config.Compression))
	}

	if config.Compression == Tar || config.Compression == GZIPTar || config.Compression == BZIP2Tar {
		decompressedReader = tar.NewReader(decompressedReader)
	}

	for {
		if config.Compression == Tar || config.Compression == GZIPTar || config.Compression == BZIP2Tar {
			// Go to the first or next file in gzip/tar.
			_, err = decompressedReader.(*tar.Reader).Next()
			if err != nil {
				// When there are no more files in gzip/tar, Next returns io.EOF.
				if errors.Is(err, io.EOF) {
					// Make sure the whole file is written out to compressedFile.
					_, _ = io.Copy(io.Discard, compressedReader)
				} else {
					errs <- errors.WithStack(err)
				}
				return
			}
		}

		decoder := json.NewDecoder(decompressedReader)

		if config.FileType == JSONArray {
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
			select {
			case <-ctx.Done():
				errs <- errors.WithStack(ctx.Err())
				return
			case output <- raw:
			}
		}

		if config.FileType == JSONArray {
			// Read closing bracket.
			_, err = decoder.Token()
			if err != nil {
				errs <- errors.WithStack(err)
				return
			}
		}

		if config.Compression != Tar && config.Compression != GZIPTar && config.Compression != BZIP2Tar {
			// Only tar can have multiple files.
			break
		}
	}

	// Make sure the whole file is written out to compressedFile.
	_, _ = io.Copy(io.Discard, compressedReader)
}

func decodeRows(
	ctx context.Context, config *ProcessConfig, wg *sync.WaitGroup,
	input <-chan []byte, output chan<- interface{}, errs chan<- errors.E,
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
			select {
			case <-ctx.Done():
				errs <- errors.WithStack(ctx.Err())
				return
			case output <- e:
			}
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

// Process is a low-level function which decompresses a file (supports Compression compressions),
// extacts JSONs or SQL statements from it (stored in FileType types), decodes JSONs or SQL statements, and
// calls Process callback on each decoded JSON or SQL statement. All that in parallel fashion, controlled by
// DecompressionThreads, DecodingThreads, and ItemsProcessingThreads. File is downloaded from a HTTP URL and is
// processed already during download. Downloaded file is optionally cached to local storage (to CacheDir
// directory, with filename as determined by CacheFile) and followup calls to Process use a cached file
// (if it matches CacheGlob, which should match at most one file).
func Process(ctx context.Context, config *ProcessConfig) errors.E {
	if config.DecompressionThreads == 0 {
		config.DecompressionThreads = runtime.GOMAXPROCS(0)
	}
	if config.DecodingThreads == 0 {
		config.DecodingThreads = runtime.GOMAXPROCS(0)
	}
	if config.ItemsProcessingThreads == 0 {
		config.ItemsProcessingThreads = runtime.GOMAXPROCS(0)
	}

	// We call cancel on any error from goroutines. The expectation is that all
	// goroutines return soon afterwards.
	// TODO: Use golang.org/x/sync/errgroup instead?
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// mainWg counts groups of same goroutines.
	var mainWg sync.WaitGroup
	// mainWgChan is closed when mainWg is done.
	mainWgChan := make(chan struct{})

	errs := make(chan errors.E, 1+config.DecodingThreads+config.ItemsProcessingThreads)
	defer close(errs)

	rows := make(chan []byte, config.DecodingThreads)
	items := make(chan interface{}, config.ItemsProcessingThreads)

	var getFileRowsWg sync.WaitGroup
	mainWg.Add(1)
	getFileRowsWg.Add(1)
	go getFileRows(ctx, config, &getFileRowsWg, rows, errs)
	go func() {
		getFileRowsWg.Wait()
		mainWg.Done()
		// All goroutines using rows channel as output are done,
		// we can close the channel.
		close(rows)
	}()

	var decodeRowsWg sync.WaitGroup
	mainWg.Add(1)
	for w := 0; w < config.DecodingThreads; w++ {
		decodeRowsWg.Add(1)
		go decodeRows(ctx, config, &decodeRowsWg, rows, items, errs)
	}
	go func() {
		decodeRowsWg.Wait()
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
		// This function is closing errs in defer, so we do not have
		// to check if the channel is closed.
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
