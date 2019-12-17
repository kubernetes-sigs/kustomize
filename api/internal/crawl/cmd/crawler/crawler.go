package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"sigs.k8s.io/kustomize/api/internal/crawl/crawler"
	"sigs.k8s.io/kustomize/api/internal/crawl/crawler/github"
	"sigs.k8s.io/kustomize/api/internal/crawl/doc"
	"sigs.k8s.io/kustomize/api/internal/crawl/httpclient"
	"sigs.k8s.io/kustomize/api/internal/crawl/index"

	"github.com/gomodule/redigo/redis"
)

const (
	githubAccessTokenVar = "GITHUB_ACCESS_TOKEN"
	redisCacheURL        = "REDIS_CACHE_URL"
	redisKeyURL          = "REDIS_KEY_URL"
	retryCount           = 3
)

type CrawlMode int
const (
	CrawlUnknown CrawlMode = iota
	// Crawl all the kustomization files in all the repositories of a Github user
	CrawlUser
	// Crawl all the kustomization files in a Github repo
	CrawlRepo
	// Crawl all the documents in the index
	CrawlIndex
	// Crawl all the kustomization files on Github
	CrawlGithub
	// Crawl all the documents in the index and crawling all the kustomization files on Github
	CrawlIndexAndGithub
)

func NewCrawlMode(s string) CrawlMode {
	switch s {
	case "github-user":
		return CrawlUser
	case "github-repo":
		return CrawlRepo
	case "":
		return CrawlIndexAndGithub
	case "index":
		return CrawlIndex
	case "github":
		return CrawlGithub
	default:
		return CrawlUnknown
	}
}

func Usage() {
	fmt.Printf("Usage: %s [mode] [githubUser|githubRepo]\n", os.Args[0])
	fmt.Printf("\tmode can be one of [github-user, github-repo, index, github]\n")
	fmt.Printf("%s: crawl all the documents in the index and crawling all the kustomization files on Github\n", os.Args[0])
	fmt.Printf("%s index: crawl all the documents in the index\n", os.Args[0])
	fmt.Printf("%s gihub: crawl all the kustomization files on Github\n", os.Args[0])
	fmt.Printf("%s github-user <github-user>: Crawl all the kustomization files in all the repositories of a Github user\n", os.Args[0])
	fmt.Printf("\tFor example, %s github-user kubernetes-sigs\n", os.Args[0])
	fmt.Printf("%s github-repo <github-repo>: Crawl all the kustomization files in a Github repo\n", os.Args[0])
	fmt.Printf("\tFor example, %s github-repo kubernetes-sigs/kustomize\n", os.Args[0])
}

func main() {
	githubToken := os.Getenv(githubAccessTokenVar)
	if githubToken == "" {
		fmt.Printf("Must set the variable '%s' to make github requests.\n",
			githubAccessTokenVar)
		return
	}

	ctx := context.Background()
	idx, err := index.NewKustomizeIndex(ctx)
	if err != nil {
		fmt.Printf("Could not create an index: %v\n", err)
		return
	}

	cacheURL := os.Getenv(redisCacheURL)
	cache, err := redis.DialURL(cacheURL)
	clientCache := &http.Client{}
	if err != nil {
		fmt.Printf("Error: redis could not make a connection: %v\n", err)
	} else {
		clientCache = httpclient.NewClient(cache)
	}

	// docConverter takes in a plain document and processes it for the index.
	docConverter := func(d *doc.Document) (crawler.CrawledDocument, error) {
		kdoc := doc.KustomizationDocument{
			Document: *d,
		}

		err := kdoc.ParseYAML()
		return &kdoc, err
	}

	// Index updates the value in the index.
	indexFunc := func(cdoc crawler.CrawledDocument, mode index.Mode) error {
		switch d := cdoc.(type) {
		case *doc.KustomizationDocument:
			switch mode {
			case index.Delete:
				fmt.Println("Deleting: ", d)
				return idx.Delete(d.ID())
			default:
				fmt.Println("Inserting: ", d)
				return idx.Put(d.ID(), d)
			}
		default:
			return fmt.Errorf("type %T not supported", d)
		}
	}

	// seen tracks the IDs of all the documents in the index.
	// This helps avoid indexing a given document multiple times.
	seen := make(map[string]struct{})

	var mode CrawlMode
	if len(os.Args) == 1 {
		mode = CrawlIndexAndGithub
	} else {
		mode = NewCrawlMode(os.Args[1])
	}

	ghCrawlerConstructor := func(user, repo string) crawler.Crawler {
		if user != "" {
			return 	github.NewCrawler(githubToken, retryCount, clientCache,
				github.QueryWith(
					github.Filename("kustomization.yaml"),
					github.Filename("kustomization.yml"),
					github.User(user)),
			)
		} else if repo != "" {
			return github.NewCrawler(githubToken, retryCount, clientCache,
				github.QueryWith(
					github.Filename("kustomization.yaml"),
					github.Filename("kustomization.yml"),
					github.Repo(repo)),
			)
		} else {
			return github.NewCrawler(githubToken, retryCount, clientCache,
				github.QueryWith(
					github.Filename("kustomization.yaml"),
					github.Filename("kustomization.yml")),
			)
		}
	}

	seedDocs := make(crawler.CrawlSeed, 0)

	// get all the documents in the index
	getSeedDocsFunc := func() {
		query := []byte(`{ "query":{ "match_all":{} } }`)
		it := idx.IterateQuery(query, 10000, 60*time.Second)
		for it.Next() {
			for _, hit := range it.Value().Hits.Hits {
				seedDocs = append(seedDocs, hit.Document.Copy())
			}
		}
		if err := it.Err(); err != nil {
			fmt.Printf("Error iterating: %v\n", err)
		}
	}

	switch mode {
	case CrawlIndexAndGithub:
		getSeedDocsFunc()
		crawlers := []crawler.Crawler{ghCrawlerConstructor("", "")}
		crawler.CrawlFromSeed(ctx, seedDocs, crawlers, docConverter, indexFunc, seen)
		crawler.CrawlGithub(ctx, crawlers, docConverter, indexFunc, seen)
	case CrawlIndex:
		getSeedDocsFunc()
		crawlers := []crawler.Crawler{ghCrawlerConstructor("", "")}
		crawler.CrawlFromSeed(ctx, seedDocs, crawlers, docConverter, indexFunc, seen)
	case CrawlGithub:
		crawlers := []crawler.Crawler{ghCrawlerConstructor("", "")}
		crawler.CrawlGithub(ctx, crawlers, docConverter, indexFunc, seen)
	case CrawlUser:
		if len(os.Args) < 3 {
			Usage()
			log.Fatalf("Please specify a github user!")
		}
		crawlers := []crawler.Crawler{ghCrawlerConstructor(os.Args[2], "")}
		crawler.CrawlGithub(ctx, crawlers, docConverter, indexFunc, seen)
	case CrawlRepo:
		if len(os.Args) < 3 {
			Usage()
			log.Fatalf("Please specify a github repo!")
		}
		crawlers := []crawler.Crawler{ghCrawlerConstructor("", os.Args[2])}
		crawler.CrawlGithub(ctx, crawlers, docConverter, indexFunc, seen)
	case CrawlUnknown:
		Usage()
		log.Fatalf("The crawler mode must be one of [github-user, github-repo, index, github]")
	}
}
