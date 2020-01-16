package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"sigs.k8s.io/kustomize/api/internal/crawl/utils"

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
	case "index+github":
		return CrawlIndexAndGithub
	case "index":
		return CrawlIndex
	case "github":
		return CrawlGithub
	default:
		return CrawlUnknown
	}
}

func main() {
	indexNamePtr := flag.String(
		"index", "kustomize", "The name of the ElasticSearch index.")
	modePtr := flag.String("mode", "index+github",
		`The crawling mode, which can be one of [github-user, github-repo, index, github, index+github].
  * github-user: crawl all the kustomization files in all the repositories of a Github user (--github-user must be specified for this mode).
  * github-repo: crawl all the kustomization files in a Github repository (--github-repo must be specified for this mode).
  * index: crawl all the documents in the index.
  * gihub: crawl all the kustomization files on Github.
  * index+github: crawl all the documents in the index and crawling all the kustomization files on Github.`)
	githubUserPtr := flag.String("github-user", "",
		"A github user name (e.g., kubernetes-sigs). This flag is required for the `github-user` mode.")
	githubRepoPtr := flag.String("github-repo", "",
		"A github repository name (e.g., kubernetes-sigs/kustomize). This flag is required for the `github-repo` mode.")
	flag.Parse()

	githubToken := os.Getenv(githubAccessTokenVar)
	if githubToken == "" {
		log.Printf("Must set the variable '%s' to make github requests.\n",
			githubAccessTokenVar)
		return
	}

	ctx := context.Background()
	idx, err := index.NewKustomizeIndex(ctx, *indexNamePtr)
	if err != nil {
		log.Printf("Could not create an index: %v\n", err)
		return
	}

	cacheURL := os.Getenv(redisCacheURL)
	cache, err := redis.DialURL(cacheURL)
	clientCache := &http.Client{}
	if err != nil {
		log.Printf("Error: redis could not make a connection: %v\n", err)
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
				log.Printf("Deleting: %v", d)
				return idx.Delete(d.ID())
			default:
				log.Printf("Inserting: %v", d)
				return idx.Put(d.ID(), d)
			}
		default:
			return fmt.Errorf("type %T not supported", d)
		}
	}

	// seen tracks the IDs of all the documents in the index.
	// This helps avoid indexing a given document multiple times.
	seen := utils.NewSeenMap()

	mode := NewCrawlMode(*modePtr)

	ghCrawlerConstructor := func(user, repo string) crawler.Crawler {
		if user != "" {
			return github.NewCrawler(githubToken, retryCount, clientCache,
				github.QueryWith(
					github.Filename("kustomization.yaml"),
					github.Filename("kustomization.yml"),
					github.Filename("kustomization"),
					github.User(user)),
			)
		} else if repo != "" {
			return github.NewCrawler(githubToken, retryCount, clientCache,
				github.QueryWith(
					github.Filename("kustomization.yaml"),
					github.Filename("kustomization.yml"),
					github.Filename("kustomization"),
					github.Repo(repo)),
			)
		} else {
			return github.NewCrawler(githubToken, retryCount, clientCache,
				github.QueryWith(
					github.Filename("kustomization.yaml"),
					github.Filename("kustomization.yml"),
					github.Filename("kustomization")),
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
			log.Fatalf("getSeedDocsFunc Error iterating: %v\n", err)
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
		if *githubUserPtr == "" {
			flag.Usage()
			log.Fatalf("Please specify a github user with the github-user flag!")
		}
		crawlers := []crawler.Crawler{ghCrawlerConstructor(*githubUserPtr, "")}
		crawler.CrawlGithub(ctx, crawlers, docConverter, indexFunc, seen)
	case CrawlRepo:
		if *githubRepoPtr == "" {
			flag.Usage()
			log.Fatalf("Please specify a github repository with the github-repo flag!")
		}
		crawlers := []crawler.Crawler{ghCrawlerConstructor("", *githubRepoPtr)}
		crawler.CrawlGithub(ctx, crawlers, docConverter, indexFunc, seen)
	case CrawlUnknown:
		flag.Usage()
		log.Fatalf("The --mode flag must be one of [github-user, github-repo, index, github, index+github].")
	}
}
