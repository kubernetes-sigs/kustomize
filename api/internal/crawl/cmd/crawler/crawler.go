// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

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
)

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

	query := []byte(`{ "query":{ "match_all":{} } }`)
	it := idx.IterateQuery(query, 10000, 60*time.Second)
	docs := make(crawler.CrawlSeed, 0)
	for it.Next() {
		for _, hit := range it.Value().Hits.Hits {
			docs = append(docs, hit.Document.Copy())
		}
	}

	if err := it.Err(); err != nil {
		fmt.Printf("Error iterating: %v\n", err)
	}

	cache, err := redis.DialURL(cacheURL)
	clientCache := &http.Client{}
	if err != nil {
		fmt.Printf("Error: redis could not make a connection: %v\n", err)
	} else {
		clientCache = httpclient.NewClient(cache)
	}

	ghCrawler := github.NewCrawler(githubToken, retryCount, clientCache,
		github.QueryWith(
			github.Filename("kustomization.yaml"),
			github.Filename("kustomization.yml")),
	)

	crawler.CrawlFromSeed(ctx, docs, []crawler.Crawler{ghCrawler},
		// Converter takes in a plain document and processes it for the
		// index.
		func(d *doc.Document) (crawler.CrawledDocument, error) {
			kdoc := doc.KustomizationDocument{
				Document: *d,
			}

			err := kdoc.ParseYAML()
			return &kdoc, err
		},
		// IndexFunc updates the value in the index.
		func(cdoc crawler.CrawledDocument, crwlr crawler.Crawler) error {
			switch d := cdoc.(type) {
			case *doc.KustomizationDocument:
				fmt.Println("Inserting: ", d.ID(), d)
				_, err := idx.Put(d.ID(), d)
				return err
			default:
				return fmt.Errorf("type %T not supported", d)
			}
		},
	)
}
