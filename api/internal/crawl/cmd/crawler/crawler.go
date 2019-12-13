package main

import (
	"context"
	"fmt"
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
	githubUserEnv        = "GITHUB_USER"
	githubRepoEnv        = "GITHUB_REPO"
)

func main() {
	githubUser := os.Getenv(githubUserEnv)
	githubRepo := os.Getenv(githubRepoEnv)

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

	seedDocs := make(crawler.CrawlSeed, 0)

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
	index := func(cdoc crawler.CrawledDocument, crwlr crawler.Crawler) error {
		switch d := cdoc.(type) {
		case *doc.KustomizationDocument:
			fmt.Println("Inserting: ", d)
			_, err := idx.Put(d.ID(), d)
			return err
		default:
			return fmt.Errorf("type %T not supported", d)
		}
	}

	// seen tracks the IDs of all the documents in the index.
	// This helps avoid indexing a given document multiple times.
	seen := make(map[string]struct{})

	var ghCrawler crawler.Crawler

	if githubRepo != "" {
		ghCrawler = github.NewCrawler(githubToken, retryCount, clientCache,
			github.QueryWith(
				github.Filename("kustomization.yaml"),
				github.Filename("kustomization.yml"),
				github.Repo(githubRepo)),
		)
	} else if githubUser != "" {
		ghCrawler = github.NewCrawler(githubToken, retryCount, clientCache,
			github.QueryWith(
				github.Filename("kustomization.yaml"),
				github.Filename("kustomization.yml"),
				github.User(githubUser)),
		)
	} else {
		ghCrawler = github.NewCrawler(githubToken, retryCount, clientCache,
			github.QueryWith(
				github.Filename("kustomization.yaml"),
				github.Filename("kustomization.yml")),
		)

		// get all the documents in the index
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

	crawlers := []crawler.Crawler{ghCrawler}
	crawler.CrawlFromSeed(ctx, seedDocs, crawlers, docConverter, index, seen)
	crawler.CrawlGithub(ctx, crawlers, docConverter, index, seen)
}
