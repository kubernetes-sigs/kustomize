// Package crawler provides helper methods and defines an interface for lauching
// source repository crawlers that retrieve files from a source and forwards
// to a channel for indexing and retrieval.
package crawler

import (
	"context"
	"fmt"
	"sync"

	"sigs.k8s.io/kustomize/internal/search/doc"
)

// Crawler forwards documents from source repositories to index and store them
// for searching. Each crawler is responsible for querying it's source of
// information, and forwarding files that have not been seen before or that need
// updating.
type Crawler interface {
	// Crawl returns when it is done processing. This method does not take
	// ownership of the channel. The channel is write only, and it
	// designates where the crawler should forward the documents.
	Crawl(ctx context.Context, output chan<- *doc.KustomizationDocument) error
}

// CrawlerRunner is a blocking function and only returns once all of the
// crawlers are finished with execution.
//
// This function uses the output channel to forward kustomization documents
// from a list of crawlers. The output is to be consumed by a database/search
// indexer for later retrieval.
//
// The return value is an array of errors in which each index represents the
// index of the crawler that emitted the error. Although the errors themselves
// can be nil, the array will always be exactly the size of the crawlers array.
func CrawlerRunner(ctx context.Context,
	output chan<- *doc.KustomizationDocument, crawlers []Crawler) []error {

	errs := make([]error, len(crawlers))
	wg := sync.WaitGroup{}

	for i, crawler := range crawlers {
		// Crawler implementations get their own channels to prevent a
		// crawler from closing the main output channel.
		docs := make(chan *doc.KustomizationDocument)
		wg.Add(2)

		// Forward all of the documents from this crawler's channel to
		// the main output channel.
		go func(docs <-chan *doc.KustomizationDocument) {
			defer wg.Done()
			for doc := range docs {
				output <- doc
			}
		}(docs)

		// Run this crawler and capture its returned error.
		go func(idx int, crawler Crawler,
			docs chan<- *doc.KustomizationDocument) {

			defer func() {
				wg.Done()
				if r := recover(); r != nil {
					errs[idx] = fmt.Errorf(
						"%+v panicked: %v, additional error %v",
						crawler, r, errs[idx],
					)
				}
			}()
			defer close(docs)
			errs[idx] = crawler.Crawl(ctx, docs)
		}(i, crawler, docs) // Copies the index and the crawler
	}

	wg.Wait()
	return errs
}
