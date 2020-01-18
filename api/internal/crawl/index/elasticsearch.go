package index

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	es "github.com/elastic/go-elasticsearch/v6"
	"github.com/elastic/go-elasticsearch/v6/esapi"
)

const IndexConfig = `
{
	"mappings": {
		"_doc": {
			"properties": {
				"repositoryUrl": {
					"type": "keyword"
				},
				"user": {
					"type": "keyword"
				},
				"filePath": {
					"type": "keyword"
				},
				"defaultBranch": {
					"type": "keyword"
				},
				"fileType": {
					"type": "keyword"
				},
				"document": {
					"type": "text"
				},
				"creationTime": {
					"type": "date"
				},
				"kinds": {
					"type": "text"
				},
				"identifiers": {
					"type": "text"
				},
				"values": {
					"type": "text"
				}
			}
		}
	}
}`

// TODO(damienr74) Split index into reader and writer?
type index struct {
	ctx    context.Context
	client *es.Client
	name   string
}

func newIndex(ctx context.Context, name string) (*index, error) {
	client, err := es.NewDefaultClient()
	if err != nil {
		return nil, err
	}

	return &index{
		ctx:    ctx,
		client: client,
		name:   name,
	}, nil
}

type readerFunc func(io.Reader) error

func ignoreResponseBody(_ io.Reader) error {
	return nil
}

// checks that elastic returned successfully. If it has not, it will read the
// body and return it in an error message.
//
// Otherwise, it will use the readerFunc to read the body. This function is a
// mechanism for getting relevant data from the response only if it was successful.
func (idx *index) responseErrorOrNil(info string, res *esapi.Response,
	err error, reader readerFunc) error {

	messageStart := fmt.Sprintf("index %s error: %s", idx.name, info)
	if err != nil || res == nil {
		return fmt.Errorf("%s: %v", messageStart, err)
	}

	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("%s: %s [%d]", messageStart, res.String(), res.StatusCode)
	}

	if reader != nil {
		err = reader(res.Body)
		if err != nil {
			return fmt.Errorf("%s: %v", messageStart, err)
		}
	}

	return nil
}

func byteJoin(bts ...interface{}) []byte {
	ret := make([][]byte, len(bts))
	for i, v := range bts {
		switch bt := v.(type) {
		case []byte:
			ret[i] = bt
		case string:
			ret[i] = []byte(bt)
		default:
			ret[i] = []byte(fmt.Sprintf("%v", bt))
		}
	}

	return bytes.Join(ret, []byte(` `))
}

// Update the elasticsearch index mappings. (describes how to index/search for the documents).
func (idx *index) UpdateMapping(mappings []byte) error {
	request := byteJoin(`{ "mappings":`, mappings, `}`)

	op := idx.client.Indices.PutMapping
	res, err := op(
		bytes.NewReader(request),
		op.WithContext(idx.ctx),
		op.WithIndex(idx.name),
		op.WithIncludeTypeName(true),
		op.WithPretty(),
	)

	return idx.responseErrorOrNil(
		fmt.Sprintf("could not update index mappings '%s'", request),
		res, err, ignoreResponseBody)
}

// Update the elasticsearch index settings. (describes default parameters and
// some analyzer definitions, etc.)
func (idx *index) UpdateSetting(settings []byte) error {
	request := byteJoin(`{ "settings": `, settings, `}`)
	op := idx.client.Indices.PutSettings
	res, err := op(
		bytes.NewReader(request),
		op.WithContext(idx.ctx),
		op.WithIndex(idx.name),
		op.WithPretty(),
	)

	return idx.responseErrorOrNil(
		fmt.Sprintf("could not update index settings '%s'", request),
		res, err, ignoreResponseBody)
}

// Create an index providing the config for both the mappings and the settings.
func (idx *index) CreateIndex(config []byte) error {
	op := idx.client.Indices.Create
	res, err := op(
		idx.name,
		op.WithBody(bytes.NewReader(config)),
		op.WithContext(idx.ctx),
		op.WithHuman(),
		op.WithPretty(),
	)

	return idx.responseErrorOrNil(
		fmt.Sprintf("could not create index with config '%s'", config),
		res, err, ignoreResponseBody)
}

// Delete an index.
func (idx *index) DeleteIndex() error {
	res, err := idx.client.Indices.Delete(
		[]string{idx.name},
	)

	return idx.responseErrorOrNil("could not delete index",
		res, err, ignoreResponseBody)
}

// Insert or update the document by ID.
func (idx *index) Put(uniqueID string, doc interface{}) error {
	exists, err := idx.Exists(uniqueID)
	if err != nil {
		return err
	}

	if exists {
		docBytes, err := json.Marshal(doc)
		if err != nil {
			return err
		}
		body := byteJoin(`{"doc":`, docBytes, `}`)

		// For a document with a given id, every call of IndexRequest.Do will increase the version of a document.
		// To avoid increasing the document version unnecessarily, use UpdateRequest here.
		req := esapi.UpdateRequest{
			Index:      idx.name,
			Body:       bytes.NewReader(body),
			DocumentID: uniqueID,
		}
		res, err := req.Do(idx.ctx, idx.client)

		err = idx.responseErrorOrNil("could not update document",
			res, err, ignoreResponseBody)
	} else {
		body, err := json.Marshal(doc)
		if err != nil {
			return err
		}

		req := esapi.IndexRequest{
			Index:      idx.name,
			Body:       bytes.NewReader(body),
			DocumentID: uniqueID,
		}
		res, err := req.Do(idx.ctx, idx.client)

		err = idx.responseErrorOrNil("could not insert document",
			res, err, ignoreResponseBody)
	}
	return err
}

type scrollUpdater func(string, readerFunc) error

// Update the scroll for iteration. If no scroll exists, create one.
func (idx *index) scrollUpdater(query []byte, batchSize int,
	timeout time.Duration) scrollUpdater {

	return func(scrollID string, reader readerFunc) error {
		var res *esapi.Response
		var err error

		if scrollID == "" {
			search := idx.client.Search
			res, err = search(
				search.WithContext(idx.ctx),
				search.WithIndex(idx.name),
				search.WithBody(bytes.NewBuffer(query)),
				search.WithScroll(timeout),
				search.WithSize(batchSize),
			)
		} else {
			scroll := idx.client.Scroll
			res, err = scroll(
				scroll.WithContext(idx.ctx),
				scroll.WithScroll(timeout),
				scroll.WithScrollID(scrollID),
			)
		}

		return idx.responseErrorOrNil(
			fmt.Sprintf("could not scroll for query %s", query),
			res, err, reader)
	}
}

// Simple search options. Size is the number of elements to return, From is the
// rank of the results according to the query. Used as a simple (stateless)
// pagination technique.
type SearchOptions struct {
	Size int
	From int
}

// Search for a query (json query dsl) with some options, and use the reader func
// to extract the response.
func (idx *index) Search(query []byte, opts SearchOptions,
	responseReader readerFunc) error {

	op := idx.client.Search
	res, err := op(
		op.WithContext(idx.ctx),
		op.WithIndex(idx.name),
		op.WithBody(bytes.NewBuffer(query)),
		op.WithTrackTotalHits(true),
		op.WithSize(opts.Size),
		op.WithFrom(opts.From),
		op.WithPretty(),
	)

	return idx.responseErrorOrNil(
		fmt.Sprintf("could not complete search query %v", query),
		res, err, responseReader)
}

// Delete an element from elasticsearch by Id.
func (idx *index) Delete(id string) error {
	op := idx.client.Delete
	res, err := op(
		idx.name,
		id,
		op.WithContext(idx.ctx),
		op.WithPretty(),
	)

	return idx.responseErrorOrNil(
		fmt.Sprintf("could not delete id(%s) from index(%s)", id, idx.name),
		res, err, ignoreResponseBody)
}

// Check whether a given document id is in the index
func (idx *index) Exists(id string) (bool, error) {
	op := idx.client.Exists
	res, err := op(
		idx.name,
		id,
		op.WithContext(idx.ctx),
		op.WithPretty(),
	)

	if res != nil && !res.IsError() {
		return true, nil
	} else if res != nil && res.StatusCode == 404 {
		return false, nil
	} else {
		return false, idx.responseErrorOrNil(
			fmt.Sprintf("could not check the existence of id(%s) from index(%s)", id, idx.name),
			res, err, ignoreResponseBody)
	}
}
