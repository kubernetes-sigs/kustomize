package index

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"sigs.k8s.io/kustomize/api/internal/crawl/doc"
)

const (
	AggregationKeyword = "aggs"
)

type Mode int

const (
	InsertOrUpdate = iota
	Delete
)

// Redefinition of Hits structure. Must match the json string of
// KustomizeResult.Hits.Hits. Declared as a convenience for iteration.
type KustomizeHits []struct {
	ID       string                    `json:"id"`
	Document doc.KustomizationDocument `json:"result"`
}

type KustomizeResult struct {
	ScrollID *string `json:"-"`

	Hits *struct {
		Total int `json:"total"`
		Hits  []struct {
			ID       string                    `json:"id"`
			Document doc.KustomizationDocument `json:"result"`
		} `json:"hits"`
	} `json:"hits,omitempty"`

	Aggregations *struct {
		Timeseries *struct {
			Buckets []struct {
				Key   string `json:"key"`
				Count int    `json:"count"`
			} `json:"buckets"`
		} `json:"timeseries,omitempty"`

		Kinds *struct {
			OtherCount int `json:"otherResults"`
			Buckets    []struct {
				Key   string `json:"key"`
				Count int    `json:"count"`
			} `json:"buckets"`
		} `json:"kinds,omitempty"`
	} `json:"aggregations,omitempty"`
}

// Elasticsearch has some sometimes inconsistent labels, and some pretty ugly label choices.
// However, the structure seems reasonable, so I wanted to use it if possible. This method
// needs two copies of the types to make the json strings different. The Copies must be the
// exact same type/structure, so the types must be declared inline. Go will check that these
// are convertible at compile time, and converting at runtime is a noop.
type ElasticKustomizeResult struct {
	ScrollID *string `json:"_scroll_id,omitempty"`

	Hits *struct {
		Total int `json:"total"`
		Hits  []struct {
			ID       string                    `json:"_id"`
			Document doc.KustomizationDocument `json:"_source"`
		} `json:"hits"`
	} `json:"hits,omitempty"`

	Aggregations *struct {
		Timeseries *struct {
			Buckets []struct {
				Key   string `json:"key_as_string"`
				Count int    `json:"doc_count"`
			}
		} `json:"timeseries,omitempty"`

		Kinds *struct {
			OtherCount int `json:"sum_other_doc_count"`
			Buckets    []struct {
				Key   string `json:"key"`
				Count int    `json:"doc_count"`
			}
		} `json:"kinds,omitempty"`
	} `json:"aggregations,omitempty"`
}

type KustomizeIndex struct {
	*index
}

// Create index reference to the index containing the kustomize documents.
func NewKustomizeIndex(ctx context.Context, indexName string) (*KustomizeIndex, error) {
	idx, err := newIndex(ctx, indexName)
	if err != nil {
		return nil, err
	}

	indicesExistsOp := idx.client.Indices.Exists
	resp, err := indicesExistsOp([]string{indexName},
		indicesExistsOp.WithContext(idx.ctx),
		indicesExistsOp.WithPretty())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 200 {
		log.Printf("The %s index already exists", indexName)
	} else {
		log.Printf("Creating the %s index\n", indexName)
		if err := idx.CreateIndex([]byte(IndexConfig)); err != nil {
			return nil, err
		}
	}

	return &KustomizeIndex{idx}, nil
}

// Return a timeseries of kustomization file counts.
func TimeseriesAggregation() (string, map[string]interface{}) {
	return "timeseries", map[string]interface{}{
		"date_histogram": map[string]interface{}{
			"field":    "creationTime",
			"interval": "day",
			/// XXX Only return values with counts, otherwise
			// every day is added to the output...
			// This matters if ever a zero valued time would
			// be stored in the creationTime field... it would
			// return >600k entries (for every day since year 0).
			// IDK why this is default, but I would not want this
			// to happen...
			"min_doc_count": 1,
		},
	}
}

// Return aggregation of results based off of their kinds.
func KindAggregation(maxBuckets int) (string, map[string]interface{}) {
	if maxBuckets < 1 {
		maxBuckets = 1
	}
	return "kinds", map[string]interface{}{
		"terms": map[string]interface{}{
			"field": "kinds.keyword",
			"size":  maxBuckets,
		},
	}
}

// The multi_match search type in elasticsearch will check each field according
// to their respective analyzers for the identifier.
func multiMatch(query string) map[string]interface{} {
	return map[string]interface{}{
		"multi_match": map[string]interface{}{
			"type": "cross_fields",
			"fields": []string{
				"values.keyword^3",
				"identifiers.keyword^3",
				"values.ngram",
				"identifiers.ngram",
				// TODO(damienr74) remove document with default
				// analyzer. It does not handle special (=,: etc)
				// characters properly, and matches with false
				// positives. document.whitespace does not exist
				// yet, but should use the whitespace analyzer.
				"document",
				"document.whitespace",
			},
			"query": query,
		},
	}
}

// Build an elasticsearch query from a user query.
func BuildQuery(query string) map[string]interface{} {
	queryTokens := strings.Fields(query)
	if len(queryTokens) == 0 {
		return map[string]interface{}{
			"size": 0,
		}
	}

	mustMatch := make([]map[string]interface{}, len(queryTokens))

	for i, tok := range queryTokens {
		if strings.HasPrefix(strings.ToLower(tok), "kind=") {
			mustMatch[i] = map[string]interface{}{
				"term": map[string]interface{}{
					"kinds.keyword": tok[5:],
				},
			}
			continue
		}
		mustMatch[i] = multiMatch(tok)
	}

	structuredQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustMatch,
			},
		},
	}

	return structuredQuery
}

// Iterator based off of the way bufio.Scanner works.
//
// Example:
//	for it.Next() {
//		for _, doc := range it.Value().Hits {
//			// Handle KustomizationDocument.
//		}
//	}
//
//	if err := it.Err(); err != nil {
//		// Handle err.
//	}
type KustomizeIterator struct {
	update scrollUpdater
	err    error
	// Matches the return definition of elasticsearch search results. The
	// scroll ID is practically a database cursor.
	scrollImpl KustomizeResult
}

// Get the next batch of results. Note that this returns multiple results that
// can be iterated.
func (it *KustomizeIterator) Next() bool {
	reader := func(reader io.Reader) error {
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("could not read from body: %v", err)
		}
		var scrollInput ElasticKustomizeResult
		err = json.Unmarshal(data, &scrollInput)
		if err != nil {
			return fmt.Errorf("cloud not marshal %s into %T: %v",
				data, scrollInput, err)
		}
		it.scrollImpl = KustomizeResult(scrollInput)

		return nil
	}

	if it.err == nil {
		log.Printf("updating scroll: %s\n", *it.scrollImpl.ScrollID)
		it.err = it.update(*it.scrollImpl.ScrollID, reader)
	}

	// if there is no error and the array is not empty, then Value is
	// obligated to return a valid result.
	return it.err == nil &&
		it.scrollImpl.Hits != nil &&
		len(it.scrollImpl.Hits.Hits) > 0
}

// Get the value from this batch of iterations.
func (it *KustomizeIterator) Value() KustomizeResult {
	return it.scrollImpl
}

// Check if any errors have occurred.
func (it *KustomizeIterator) Err() error {
	return it.err
}

// Create an iterator over query. Iterate in chunks of batchSize, each batch
// should take no longer than timeout to read (otherwise, elasticsearch will
// delete the context).
//
// XXX Important to set a reasonable amount of time to read the documents. If
// a lot of processing must be done, consider loading everything in memory
// before doing it so that, a short timeout period can be set. Scrolling creates
// a consistent DB context, so this can be costly.
//
// Scrolling is also not meant to be used for real time purposes. If you need
// results quickly, consider using the From: field in SearchOptions and a normal
// search. This will not guarantee that the values will not change but is more
// suitable for lower latencies/long execution timeouts.
func (ki *KustomizeIndex) IterateQuery(query []byte, batchSize int,
	timeout time.Duration) *KustomizeIterator {

	emptyScroll := ""
	return &KustomizeIterator{
		update: ki.scrollUpdater(query, batchSize, timeout),
		scrollImpl: KustomizeResult{
			ScrollID: &emptyScroll,
		},
	}
}

// type specific Put for inserting structured kustomization documents.
func (ki *KustomizeIndex) Put(id string, doc *doc.KustomizationDocument) error {
	return ki.index.Put(id, doc)
}

// Delete a document with a given id from the kustomize index.
func (ki *KustomizeIndex) Delete(id string) error {
	return ki.index.Delete(id)
}

// Kustomize search options: What metrics should be returned? Kind Aggregation,
// TimeseriesAggregation, etc. Also embedds the SearchOptions field to specify
// the position in the sorted list of results and the number of results to return.
type KustomizeSearchOptions struct {
	SearchOptions
	KindAggregation       bool
	TimeseriesAggregation bool
}

// Search the index with the given query string. Returns a structured result and possible
// aggregates.
func (ki *KustomizeIndex) Search(query string,
	opts KustomizeSearchOptions) (*KustomizeResult, error) {

	aggMap := make(map[string]interface{})
	if opts.KindAggregation {
		k, kAgg := KindAggregation(15)
		aggMap[k] = kAgg
	}
	if opts.TimeseriesAggregation {
		t, tAgg := TimeseriesAggregation()
		aggMap[t] = tAgg
	}

	esQuery := BuildQuery(query)
	if len(aggMap) > 0 {
		esQuery[AggregationKeyword] = aggMap
	}

	data, err := json.Marshal(&esQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to format query %s", query)
	}
	log.Printf("formated query: %s\n", data)

	var kr ElasticKustomizeResult
	err = ki.index.Search(data, opts.SearchOptions, func(results io.Reader) error {
		data, err = ioutil.ReadAll(results)
		if err != nil {
			return fmt.Errorf("could not read results from search: %v", err)
		}

		if err = json.Unmarshal(data, &kr); err != nil {
			return fmt.Errorf("could not parse results from search: %v", err)
		}

		return nil
	})
	res := KustomizeResult(kr)

	return &res, err
}
