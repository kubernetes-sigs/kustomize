package index

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	es "github.com/elastic/go-elasticsearch/v6"
	"github.com/elastic/go-elasticsearch/v6/esapi"
)

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
		return fmt.Errorf("%s: %s", messageStart, res.String())
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

// Create an index providing both the mappings and the settings.
func (idx *index) CreateIndex(mappings []byte, settings []byte) error {
	request := byteJoin(`{ "mappings":`, mappings, `, "settings":`, settings, `}`)
	op := idx.client.Indices.Create
	res, err := op(
		idx.name,
		op.WithBody(bytes.NewReader(request)),
		op.WithContext(idx.ctx),
		op.WithHuman(),
		op.WithPretty(),
		op.WithIncludeTypeName(true),
	)

	return idx.responseErrorOrNil(
		fmt.Sprintf("could not create index with config '%s'", request),
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
func (idx *index) Put(uniqueID string, doc interface{}) (string, error) {
	body, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}

	req := esapi.IndexRequest{
		Index:      idx.name,
		Body:       bytes.NewReader(body),
		DocumentID: uniqueID,
	}
	res, err := req.Do(idx.ctx, idx.client)

	var id string
	readId := func(reader io.Reader) error {
		type InsertResult struct {
			ID string `json:"_id,omitempty"`
		}
		var ir InsertResult
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &ir)
		if err != nil {
			return err
		}
		id = ir.ID

		return nil
	}

	// populates the id field.
	err = idx.responseErrorOrNil("could not insert document",
		res, err, readId)

	return id, err
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
