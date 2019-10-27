// Package crawler provides helper methods and defines an interface for lauching
// source repository crawlers that retrieve files from a source and forwards
// to a channel for indexing and retrieval.
package crawler

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/gomodule/redigo/redis"

	"sigs.k8s.io/kustomize/hack/crawl/doc"
)

var (
	logger = log.New(os.Stdout, "Crawler: ", log.LstdFlags|log.LUTC|log.Llongfile)
)

// Crawler forwards documents from source repositories to index and store them
// for searching. Each crawler is responsible for querying it's source of
// information, and forwarding files that have not been seen before or that need
// updating.
type Crawler interface {
	// Crawl returns when it is done processing. This method does not take
	// ownership of the channel. The channel is write only, and it
	// designates where the crawler should forward the documents.
	Crawl(ctx context.Context, output chan<- CrawledDocument) error

	// Get the document data given the FilePath, Repo, and Ref/Tag/Branch.
	FetchDocument(context.Context, *doc.Document) error
	// Write to the document what the created time is.
	SetCreated(context.Context, *doc.Document) error

	Match(*doc.Document) bool
}

type CrawledDocument interface {
	ID() string
	GetDocument() *doc.Document
	GetResources() ([]*doc.Document, error)
	WasCached() bool
}

type CrawlSeed []*doc.Document

type IndexFunc func(CrawledDocument, Crawler) error
type Converter func(*doc.Document) (CrawledDocument, error)

// Cleaner, more efficient, and more extensible crawler implementation.
// The seed must include the ids of each document in the index.
func CrawlFromSeed(ctx context.Context, seed CrawlSeed,
	crawlers []Crawler, conv Converter, indx IndexFunc) {

	seen := make(map[string]struct{})

	logIfErr := func(err error) {
		if err == nil {
			return
		}
		logger.Println("error: ", err)
	}

	stack := make(CrawlSeed, 0)

	findMatch := func(d *doc.Document) Crawler {
		for _, crawl := range crawlers {
			if crawl.Match(d) {
				return crawl
			}
		}

		return nil
	}

	addBranches := func(cdoc CrawledDocument, match Crawler) {
		if _, ok := seen[cdoc.ID()]; ok {
			return
		}

		seen[cdoc.ID()] = struct{}{}
		// Insert into index
		err := indx(cdoc, match)
		logIfErr(err)
		if err != nil {
			return
		}

		deps, err := cdoc.GetResources()
		logIfErr(err)
		if err != nil {
			return
		}
		for _, dep := range deps {
			if _, ok := seen[dep.ID()]; ok {
				continue
			}
			stack = append(stack, dep)
		}
	}

	doCrawl := func(docsPtr *CrawlSeed) {
		for len(*docsPtr) > 0 {
			back := len(*docsPtr) - 1
			next := (*docsPtr)[back]
			*docsPtr = (*docsPtr)[:back]

			match := findMatch(next)
			if match == nil {
				logIfErr(fmt.Errorf(
					"%v could not match any crawler", next))
				continue
			}

			err := match.FetchDocument(ctx, next)
			logIfErr(err)
			// If there was no change or there is an error, we don't have
			// to branch out, since the dependencies are already in the
			// index, or we cannot find the document.
			if err != nil || next.WasCached() {
				continue
			}

			cdoc, err := conv(next)
			logIfErr(err)
			if err != nil {
				continue
			}

			addBranches(cdoc, match)
		}
	}
	// Exploit seed to update bulk of corpus.
	logger.Printf("updating %d documents from seed\n", len(seed))
	doCrawl(&seed)
	// Traverse any new links added while updating corpus.
	logger.Printf("crawling %d new documents found in the seed\n", len(stack))
	doCrawl(&stack)

	ch := make(chan CrawledDocument, 1<<10)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for cdoc := range ch {
			if _, ok := seen[cdoc.ID()]; ok {
				continue
			}
			match := findMatch(cdoc.GetDocument())
			if match == nil {
				logIfErr(fmt.Errorf(
					"%v could not match any crawler", cdoc))
				continue
			}
			addBranches(cdoc, match)
		}
	}()

	// Exploration through APIs.
	errs := CRunner(ctx, ch, crawlers)
	if errs != nil {
		for _, err := range errs {
			logIfErr(err)
		}
	}
	close(ch)
	logger.Println("Processing the new documents from the crawlers' exploration.")
	wg.Wait()
	// Handle deps of newly discovered documents.
	logger.Printf("crawling the %d new documents from the crawlers' exploration.",
		len(stack))
	doCrawl(&stack)
}

// CRunner is a blocking function and only returns once all of the
// crawlers are finished with execution.
//
// This function uses the output channel to forward kustomization documents
// from a list of crawlers. The output is to be consumed by a database/search
// indexer for later retrieval.
//
// The return value is an array of errors in which each index represents the
// index of the crawler that emitted the error. Although the errors themselves
// can be nil, the array will always be exactly the size of the crawlers array.
//
// CRunner takes in a seed, which represents the documents stored in an
// index somewhere. The document data is not required to be populated. If there
// are many documents, this is preferable. The order of iteration over the seed
// is not garanteed, but the CRunner does guarantee that every element
// from the seed will be processed before any other documents from the
// crawlers.
func CRunner(ctx context.Context,
	output chan<- CrawledDocument, crawlers []Crawler) []error {

	errs := make([]error, len(crawlers))
	wg := sync.WaitGroup{}

	for i, crawler := range crawlers {
		// Crawler implementations get their own channels to prevent a
		// crawler from closing the main output channel.
		docs := make(chan CrawledDocument)
		wg.Add(2)

		// Forward all of the documents from this crawler's channel to
		// the main output channel.
		go func(docs <-chan CrawledDocument) {
			defer wg.Done()
			for d := range docs {
				output <- d
			}
		}(docs)

		// Run this crawler and capture its returned error.
		go func(idx int, crawler Crawler,
			docs chan<- CrawledDocument) {

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
