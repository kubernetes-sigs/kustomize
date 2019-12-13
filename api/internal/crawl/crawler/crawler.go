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

	"sigs.k8s.io/kustomize/api/internal/crawl/doc"
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
	// Get all the Documents directly referred in a Document.
	GetResources() ([]*doc.Document, error)
	WasCached() bool
}

type CrawlSeed []*doc.Document

type IndexFunc func(CrawledDocument, Crawler) error
type Converter func(*doc.Document) (CrawledDocument, error)

func logIfErr(err error) {
	if err == nil {
		return
	}
	logger.Println("error: ", err)
}

func findMatch(d *doc.Document, crawlers []Crawler) Crawler {
	for _, crawl := range crawlers {
		if crawl.Match(d) {
			return crawl
		}
	}
	return nil
}

func addBranches(cdoc CrawledDocument, match Crawler, indx IndexFunc,
	seen map[string]struct{}, stack *CrawlSeed) {

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
		*stack = append(*stack, dep)
	}
}

func doCrawl(ctx context.Context, docsPtr *CrawlSeed, crawlers []Crawler, conv Converter, indx IndexFunc,
	seen map[string]struct{}, stack *CrawlSeed) {
	docCount := 0
	// During the execution of the for loop, more Documents may be added into (*docsPtr).
	for len(*docsPtr) > 0 {
		// get the last Document in (*docPtr), which will be crawled in this iteration.
		tail := (*docsPtr)[len(*docsPtr)-1]

		// remove the last Document in (*docPtr)
		*docsPtr = (*docsPtr)[:(len(*docsPtr) - 1)]

		if _, ok := seen[tail.ID()]; ok {
			continue
		}
		docCount++

		match := findMatch(tail, crawlers)
		if match == nil {
			logIfErr(fmt.Errorf(
				"%v could not match any crawler", tail))
			continue
		}

		logger.Println("Crawling ", tail.RepositoryURL, tail.FilePath)
		err := match.FetchDocument(ctx, tail)
		logIfErr(err)
		// If there was no change or there is an error, we don't have
		// to branch out, since the dependencies are already in the
		// index, or we cannot find the document.
		if err != nil || tail.WasCached() {
			if tail.WasCached() {
				logger.Println(tail.RepositoryURL, tail.FilePath, "is cached already")
			}
			continue
		}

		logIfErr(match.SetCreated(ctx, tail))

		cdoc, err := conv(tail)
		logIfErr(err)

		addBranches(cdoc, match, indx, seen, stack)
	}
	logger.Printf("%d documents were crawled by doCrawl\n", docCount)
}

// CrawlFromSeed updates all the documents in seed, and crawls all the new
// documents referred in the seed.
func CrawlFromSeed(ctx context.Context, seed CrawlSeed, crawlers []Crawler,
	conv Converter, indx IndexFunc, seen map[string]struct{}) {

	// stack tracks the documents directly referred in other documents.
	stack := make(CrawlSeed, 0)

	// Exploit seed to update bulk of corpus.
	logger.Printf("updating %d documents from seed\n", len(seed))
	// each unique document in seed will be crawled once.
	doCrawl(ctx, &seed, crawlers, conv, indx, seen, &stack)

	// Traverse any new documents added while updating corpus.
	logger.Printf("crawling %d new documents found in the seed\n", len(stack))
	// While crawling each document in stack, the documents directly referred in the document
	// will be added into stack.
	// After this statement is done, stack will become empty.
	doCrawl(ctx, &stack, crawlers, conv, indx, seen, &stack)
}

// CrawlGithubRunner is a blocking function and only returns once all of the
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
// CrawlGithubRunner takes in a seed, which represents the documents stored in an
// index somewhere. The document data is not required to be populated. If there
// are many documents, this is preferable. The order of iteration over the seed
// is not guaranteed, but the CrawlGithub does guarantee that every element
// from the seed will be processed before any other documents from the
// crawlers.
func CrawlGithubRunner(ctx context.Context, output chan<- CrawledDocument,
	crawlers []Crawler) []error {

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

// CrawlGithub crawls all the kustomization files on Github.
func CrawlGithub(ctx context.Context, crawlers []Crawler, conv Converter,
	indx IndexFunc, seen map[string]struct{}) {
	// stack tracks the documents directly referred in other documents.
	stack := make(CrawlSeed, 0)

	// ch is channel where all the crawlers sends the crawled documents to.
	ch := make(chan CrawledDocument, 1<<10)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for cdoc := range ch {
			if _, ok := seen[cdoc.ID()]; ok {
				continue
			}
			match := findMatch(cdoc.GetDocument(), crawlers)
			if match == nil {
				logIfErr(fmt.Errorf(
					"%v could not match any crawler", cdoc))
				continue
			}
			addBranches(cdoc, match, indx, seen, &stack)
		}
	}()

	logger.Println("processing the documents found from crawling github")
	if errs := CrawlGithubRunner(ctx, ch, crawlers); errs != nil {
		for _, err := range errs {
			logIfErr(err)
		}
	}
	close(ch)
	wg.Wait()

	// Handle deps of newly discovered documents.
	logger.Printf("crawling the %d new documents referred by other documents",
		len(stack))
	doCrawl(ctx, &stack, crawlers, conv, indx, seen, &stack)
}
