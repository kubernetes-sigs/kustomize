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
	githubUserEnv        = "GITHUB_USER"
	githubRepoEnv        = "GITHUB_REPO"
	crawlIndexOnlyEnv    = "CRAWL_INDEX_ONLY"
	crawlGithubOnlyEnv   = "CRAWL_GITHUB_ONLY"
)

// countEnvs count the environment variables whose values are not empty.
func countEnvs(envs ...string) int {
	count := 0
	for _, env := range envs {
		if env != "" {
			count++
		}
	}
	return count
}

func main() {
	githubUser := os.Getenv(githubUserEnv)
	githubRepo := os.Getenv(githubRepoEnv)
	crawlIndexOnly := os.Getenv(crawlIndexOnlyEnv)
	crawlGithubOnly := os.Getenv(crawlGithubOnlyEnv)

	if countEnvs(githubUser, githubRepo, crawlIndexOnly, crawlGithubOnly) > 1 {
		log.Fatalf("only one of [%s, %s, %s, %s] should be set",
			githubUserEnv, githubRepoEnv, crawlIndexOnlyEnv, crawlGithubOnlyEnv)
	}

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
	indexFunc := func(cdoc crawler.CrawledDocument, crwlr crawler.Crawler, mode index.Mode) error {
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

	if crawlGithubOnly == "true" || githubRepo != "" || githubUser != "" {
		crawler.CrawlGithub(ctx, crawlers, docConverter, indexFunc, seen)
	} else if crawlIndexOnly == "true" {
		crawler.CrawlFromSeed(ctx, seedDocs, crawlers, docConverter, indexFunc, seen)
	} else {
		crawler.CrawlFromSeed(ctx, seedDocs, crawlers, docConverter, indexFunc, seen)
		crawler.CrawlGithub(ctx, crawlers, docConverter, indexFunc, seen)
	}
}
