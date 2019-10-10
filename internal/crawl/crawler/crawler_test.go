package crawler

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"sigs.k8s.io/kustomize/internal/tools/doc"
	"sigs.k8s.io/kustomize/v3/pkg/pgmconfig"
)

const (
	kustomizeRepo = "https://github.com/kubernetes-sigs/kustomize"
)

// Simple crawler that forwards it's list of documents to a provided channel and
// returns it's error to the caller.
type testCrawler struct {
	matchPrefix string
	err         error
	docs        []doc.KustomizationDocument
	lukp        map[string]int
}

func (c testCrawler) Match(d *doc.Document) bool {
	return d != nil && strings.HasPrefix(d.ID(), c.matchPrefix)
}

func (c testCrawler) FetchDocument(ctx context.Context, d *doc.Document) error {
	if i, ok := c.lukp[d.ID()]; ok {
		d.DocumentData = c.docs[i].DocumentData
		return nil
	}
	for _, suffix := range pgmconfig.KustomizationFileNames {
		fmt.Println(d.ID(), "/", suffix)
		i, ok := c.lukp[d.ID()+"/"+suffix]
		if !ok {
			continue
		}
		d.FilePath += "/" + suffix
		d.DocumentData = c.docs[i].DocumentData
		return nil
	}
	return fmt.Errorf("Document %v does not exist for matcher: %s",
		d, c.matchPrefix)
}

func (c testCrawler) SetCreated(ctx context.Context, d *doc.Document) error {
	d.CreationTime = &time.Time{}
	return nil
}

func newCrawler(matchPrefix string, err error,
	docs []doc.KustomizationDocument) testCrawler {
	c := testCrawler{
		matchPrefix: matchPrefix,
		err:         err,
		docs:        docs,
		lukp:        make(map[string]int),
	}
	for i, d := range docs {
		c.lukp[d.ID()] = i
	}
	return c
}

// Crawl implements the Crawler interface for testing.
func (c testCrawler) Crawl(ctx context.Context,
	output chan<- CrawlerDocument) error {

	for i, d := range c.docs {
		isResource := true
		for _, suffix := range pgmconfig.KustomizationFileNames {
			if strings.HasSuffix(d.FilePath, suffix) {
				isResource = false
				break
			}
		}
		if isResource {
			continue
		}
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
	fmt.Println("testing CrawlerRunner")
	tests := []struct {
		tc   []Crawler
		errs []error
		docs sortableDocs
	}{
		{
			tc: []Crawler{
				testCrawler{
					docs: []doc.KustomizationDocument{
						{Document: doc.Document{
							FilePath: "crawler1/doc1/kustomization.yaml",
						}},
						{Document: doc.Document{
							FilePath: "crawler1/doc2/kustomization.yaml",
						}},
						{Document: doc.Document{
							FilePath: "crawler1/doc3/kustomization.yaml",
						}},
					},
				},
				testCrawler{err: errors.New("crawler2")},
				testCrawler{},
				testCrawler{
					docs: []doc.KustomizationDocument{
						{Document: doc.Document{
							FilePath: "crawler4/doc1/kustomization.yaml",
						}},
						{Document: doc.Document{
							FilePath: "crawler4/doc2/kustomization.yaml",
						}},
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
				{Document: doc.Document{
					FilePath: "crawler1/doc1/kustomization.yaml",
				}},
				{Document: doc.Document{
					FilePath: "crawler1/doc2/kustomization.yaml",
				}},
				{Document: doc.Document{
					FilePath: "crawler1/doc3/kustomization.yaml",
				}},
				{Document: doc.Document{
					FilePath: "crawler4/doc1/kustomization.yaml",
				}},
				{Document: doc.Document{
					FilePath: "crawler4/doc2/kustomization.yaml",
				}},
			},
		},
	}

	for _, test := range tests {
		output := make(chan CrawlerDocument)
		wg := sync.WaitGroup{}
		wg.Add(1)

		// Run the Crawler runner with a list of crawlers.
		go func() {
			defer close(output)
			defer wg.Done()

			errs := CrawlerRunner(context.Background(),
				output, test.tc)

			// Check that errors are returned as they should be.
			if !reflect.DeepEqual(errs, test.errs) {
				t.Errorf("Expected errs (%v) to equal (%v)",
					errs, test.errs)
			}

		}()

		// Iterate over the output channel of Crawler runner.
		returned := make(sortableDocs, 0, len(test.docs))
		for o := range output {
			d, ok := o.(*doc.KustomizationDocument)
			if !ok || d == nil {
				t.Errorf("%T not expected type (%T)",
					o, d)
			}
			returned = append(returned, *d)
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

func TestCrawlFromSeed(t *testing.T) {
	fmt.Println("testing CrawlFromSeed")

	tests := []struct {
		seed    CrawlerSeed
		matcher string
		corpus  []doc.KustomizationDocument
	}{
		{
			seed: CrawlerSeed{
				{
					RepositoryURL: kustomizeRepo,
					FilePath:      "examples/helloWorld/kustomization.yaml",
				},
				{
					RepositoryURL: kustomizeRepo,
					FilePath:      "examples/other/kustomization.yaml",
				},
			},
			matcher: kustomizeRepo,
			corpus: []doc.KustomizationDocument{
				// Visited from the seed, will be ignored in the crawl.
				{Document: doc.Document{
					RepositoryURL: kustomizeRepo,
					FilePath:      "examples/helloWorld/kustomization.yaml",
					DocumentData: `
resources:
- deployment.yaml
`,
				}},
				// Also visited from the seed as a relative resource.
				{Document: doc.Document{
					RepositoryURL: kustomizeRepo,
					FilePath:      "examples/helloWorld/deployment.yaml",
					DocumentData: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello
`,
				}},
				// Visited from the seed. Has a remote import.
				{Document: doc.Document{
					RepositoryURL: kustomizeRepo,
					FilePath:      "examples/other/kustomization.yaml",
					DocumentData: `
resources:
- https://github.com/kubernetes-sigs/kustomize/examples/other/overlay
- service.yaml
`,
				}},
				// Imported as a base from the seed.
				{Document: doc.Document{
					RepositoryURL: kustomizeRepo,
					FilePath:      "examples/other/overlay/kustomization.yaml",
					DocumentData: `
resources:
- https://github.com/kubernetes-sigs/kustomize/examples/seedcrawl1
- https://github.com/kubernetes-sigs/kustomize/examples/seedcrawl2
`,
				}},
				// Imported as a resource from the seed.
				{Document: doc.Document{
					RepositoryURL: kustomizeRepo,
					FilePath:      "examples/other/service.yaml",
				}},
				// Visited from crawling seed.
				{Document: doc.Document{
					RepositoryURL: kustomizeRepo,
					FilePath:      "examples/seedcrawl1/kustomization.yml",
				}},
				// Visited from crawling seed.
				{Document: doc.Document{
					RepositoryURL: kustomizeRepo,
					FilePath:      "examples/seedcrawl2/kustomization.yaml",
					DocumentData: `
resources:
- ../base
- job.yaml
`,
				}},
				// Visited from crawling seed.
				{Document: doc.Document{
					RepositoryURL: kustomizeRepo,
					FilePath:      "examples/base/kustomization.yml",
				}},
				// Visited from crawling seed imported as resource.
				{Document: doc.Document{
					RepositoryURL: kustomizeRepo,
					FilePath:      "examples/seedcrawl2/job.yaml",
				}},
				// Visited from the crawler runner.
				{Document: doc.Document{
					RepositoryURL: kustomizeRepo,
					FilePath:      "examples/other/base/kustomization.yaml",
					DocumentData: `
resources:
- ../app
`,
				}},
				// Visited from the crawler runner.
				{Document: doc.Document{
					RepositoryURL: kustomizeRepo,
					FilePath:      "examples/other/app/kustomization.yaml",
					DocumentData: `
resources:
- resource.yaml
`,
				}},
				// Visited from crawling runner imported as resource.
				{Document: doc.Document{
					RepositoryURL: kustomizeRepo,
					FilePath:      "examples/other/app/resource.yaml",
				}},
			},
		},
	}

	for _, tc := range tests {
		cr := newCrawler(tc.matcher, nil, tc.corpus)
		visited := make(map[string]int)
		CrawlFromSeed(context.Background(), tc.seed, []Crawler{cr},
			func(d *doc.Document) (CrawlerDocument, error) {
				return &doc.KustomizationDocument{
					Document: *d,
				}, nil
			},
			func(d CrawlerDocument, cr Crawler) error {
				visited[d.ID()]++
				return nil
			},
		)
		if lv, lc := len(visited), len(tc.corpus); lv != lc {
			t.Errorf("error: %d of %d documents visited.", lv, lc)
			t.Errorf("\nvisited (%v)\nexpected (%v).", visited, cr.lukp)
		}
		for id, cnt := range visited {
			if cnt != 1 {
				t.Errorf("%s not visited once (%d)", id, cnt)
			}
		}
	}
}
