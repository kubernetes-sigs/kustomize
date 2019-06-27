package crawler

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"sync"
	"testing"

	"sigs.k8s.io/kustomize/v3/internal/search/doc"
)

var (
	errTest = errors.New("This error is expected")
)

type testCrawler struct {
	docs []doc.KustomizationDocument
	err  error
}

func (c testCrawler) Crawl(ctx context.Context, output chan<- *doc.KustomizationDocument) error {
	for i := range c.docs {
		output <- &c.docs[i]
	}
	return c.err
}

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

func TestCrawlerInterface(t *testing.T) {
	tests := []struct {
		tc   []Interface
		errs []error
		docs sortableDocs
	}{
		{
			tc: []Interface{
				testCrawler{
					docs: []doc.KustomizationDocument{
						{FilePath: "crawler1/doc1"},
						{FilePath: "crawler1/doc2"},
						{FilePath: "crawler1/doc3"},
					},
				},
				testCrawler{err: errTest},
				testCrawler{},
				testCrawler{
					docs: []doc.KustomizationDocument{
						{FilePath: "crawler4/doc1"},
						{FilePath: "crawler4/doc2"},
					},
					err: errTest,
				},
			},
			errs: []error{nil, errTest, nil, errTest},
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
		go func() {
			defer close(output)
			defer wg.Done()

			errs := Crawl(context.Background(), output, test.tc)

			if !reflect.DeepEqual(errs, test.errs) {
				t.Errorf("Expected errors (%v) to be equal to (%v)\n", errs, test.errs)
			}

		}()

		returned := make(sortableDocs, 0, len(test.docs))
		for doc := range output {
			returned = append(returned, *doc)
		}

		sort.Sort(returned)
		if !reflect.DeepEqual(returned, test.docs) {
			t.Errorf("Expected docs (%v) to be equal to (%v)\n", returned, test.docs)
		}

		wg.Wait()
	}
}
