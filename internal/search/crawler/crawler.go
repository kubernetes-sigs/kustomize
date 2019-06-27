package crawler

import (
	"context"
	"log"
	"os"
	"sync"

	"sigs.k8s.io/kustomize/v3/internal/search/doc"
)

// The crawler.Interface retrieves doc.KustomizationDocument
type Interface interface {
	// Crawl returns when it is done processing. This method does not take ownership of
	// the channel.
	Crawl(context.Context, chan<- *doc.KustomizationDocument) error
}

// This function uses the output channel to forward kustomization documents from a list of crawlers.
// The output can be consumed by an
//
// The return value is an array of errors in which each index represents the index of the crawler that emitted
// the error. Although the errors themselves can be nil, the array will always be exactly the size of the
// crawlers array.
//
// Crawl is a blocking function and only returns once all of the crawlers are finished with execution.
func Crawl(ctx context.Context, output chan<- *doc.KustomizationDocument, crawlers []Interface) []error {
	logger := log.New(os.Stdin, "Crawl:", log.LUTC|log.Ltime|log.Ldate)

	errs := make([]error, len(crawlers))

	wg := sync.WaitGroup{}

	for i, crawler := range crawlers {
		// Crawler implementations get their own channels to prevent one crawler from closing the output stream.
		docs := make(chan *doc.KustomizationDocument)
		wg.Add(2)

		// Forward all of the documents from this crawler's channel to the main output channel.
		go func(docs <-chan *doc.KustomizationDocument) {
			defer wg.Done()
			for doc := range docs {
				output <- doc
			}
		}(docs)

		// Run this crawler and capture it's returned error.
		go func(idx int, crawler Interface, docs chan<- *doc.KustomizationDocument) {
			defer func() {
				wg.Done()
				if r := recover(); r != nil {
					logger.Printf("%+v panicked: %v", crawler, r)
				}
			}()
			defer close(docs)
			errs[idx] = crawler.Crawl(ctx, docs)
		}(i, crawler, docs) // Copies the index and the crawler
	}

	wg.Wait()
	return errs
}
