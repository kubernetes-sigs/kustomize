package crawler

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"sync"
	"testing"

	"sigs.k8s.io/kustomize/internal/search/doc"
)

// Simple crawler that forwards it's list of documents to a provided channel and
// returns it's error to the caller.
type testCrawler struct {
	docs []doc.KustomizationDocument
	err  error
}

// Crawl implements the Crawler interface for testing.
func (c testCrawler) Crawl(ctx context.Context,
	output chan<- *doc.KustomizationDocument) error {

	for i := range c.docs {
		output <- &c.docs[i]
	}
	return c.err
}

// Used to make sure that we're comparing documents in order. This is needed
// since these documents will be sent concurrently.
type sortableDocs []doc.KustomizationDocument

func (s sortableDocs) Less(i, j int) bool {
	return s[i].FilePath < s[j].FilePath
}

func (s sortableDocs) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortableDocs) Len() int {
	return len(s)
}

func TestCrawlerRunner(t *testing.T) {
	tests := []struct {
		tc   []Crawler
		errs []error
		docs sortableDocs
	}{
		{
			tc: []Crawler{
				testCrawler{
					docs: []doc.KustomizationDocument{
						{FilePath: "crawler1/doc1"},
						{FilePath: "crawler1/doc2"},
						{FilePath: "crawler1/doc3"},
					},
				},
				testCrawler{err: errors.New("crawler2")},
				testCrawler{},
				testCrawler{
					docs: []doc.KustomizationDocument{
						{FilePath: "crawler4/doc1"},
						{FilePath: "crawler4/doc2"},
					},
					err: errors.New("crawler4"),
				},
			},
			errs: []error{
				nil,
				errors.New("crawler2"),
				nil,
				errors.New("crawler4"),
			},
			docs: sortableDocs{
				{FilePath: "crawler1/doc1"},
				{FilePath: "crawler1/doc2"},
				{FilePath: "crawler1/doc3"},
				{FilePath: "crawler4/doc1"},
				{FilePath: "crawler4/doc2"},
			},
		},
	}

	for _, test := range tests {
		output := make(chan *doc.KustomizationDocument)
		wg := sync.WaitGroup{}
		wg.Add(1)

		// Run the Crawler runner with a list of crawlers.
		go func() {
			defer close(output)
			defer wg.Done()

			errs := CrawlerRunner(context.Background(), output,
				test.tc)

			// Check that errors are returned as they should be.
			if !reflect.DeepEqual(errs, test.errs) {
				t.Errorf("Expected errs (%v) to equal (%v)",
					errs, test.errs)
			}

		}()

		// Iterate over the output channel of Crawler runner.
		returned := make(sortableDocs, 0, len(test.docs))
		for doc := range output {
			returned = append(returned, *doc)
		}

		// Check that all documents are received.
		sort.Sort(returned)
		if !reflect.DeepEqual(returned, test.docs) {
			t.Errorf("Expected docs (%v) to equal (%v)\n",
				returned, test.docs)
		}

		wg.Wait()
	}
}
