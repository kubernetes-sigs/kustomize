package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"sigs.k8s.io/kustomize/api/internal/crawl/index"
)

type kustomizeSearch struct {
	ctx context.Context
	// Eventually pIndex *index.PlugginIndex
	idx    *index.KustomizeIndex
	router *mux.Router
	log    *log.Logger
}

// New server. Creating a server does not launch it. To launch simply:
//	srv, _ := NewKustomizeSearch(context.Backgroud())
//	err := srv.Serve()
//	if err != nil {
//		// Handle server issues.
//	}
//
// The server has three enpoints, two of which are functional:
//
// /search: processes the ?q= parameter for a text query and
// returns a list of 10 resutls starting from the ?from= value provided,
// with the default being zero.
//
// /metrics: returns overall metrics about the files indexed. Returns
// timeseries data for kustomization files, and returns breakdown of file
// counts by their 'kind' fields
//
// /register: not implemented, but meant as an endpoint for adding new
// kustomization files to the corpus.
func NewKustomizeSearch(ctx context.Context) (*kustomizeSearch, error) {
	idx, err := index.NewKustomizeIndex(ctx)
	if err != nil {
		return nil, err
	}

	ks := &kustomizeSearch{
		ctx:    ctx,
		idx:    idx,
		router: mux.NewRouter(),
		log: log.New(os.Stdout, "Kustomize server: ",
			log.LstdFlags|log.Llongfile|log.LUTC),
	}

	return ks, nil
}

// Set up common middleware and the routes for the server.
func (ks *kustomizeSearch) routes() {

	// Setup middleware.
	ks.router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			handler.ServeHTTP(w, r)
		})
	})

	ks.router.HandleFunc("/liveness", ks.liveness()).Methods(http.MethodGet)
	ks.router.HandleFunc("/readiness", ks.readiness()).Methods(http.MethodGet)
	ks.router.HandleFunc("/search", ks.search()).Methods(http.MethodGet)
	ks.router.HandleFunc("/metrics", ks.metrics()).Methods(http.MethodGet)
	ks.router.HandleFunc("/register", ks.register()).Methods(http.MethodPost)
}

// Start listening and serving on the provided port.
func (ks *kustomizeSearch) Serve(port int) error {
	ks.routes()
	handler := cors.Default().Handler(ks.router)
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
		// Timeouts/Limits
	}

	return s.ListenAndServe()
}

// /liveness endpoint
func (ks *kustomizeSearch) liveness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

// /readyness endpoint
func (ks *kustomizeSearch) readiness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opt := index.KustomizeSearchOptions{}
		_, err := ks.idx.Search("", opt)
		if err != nil {
			http.Error(w,
				`{ "error": "could not connect to database" }`,
				http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// /register endpoint.
func (ks *kustomizeSearch) register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not implemented", http.StatusInternalServerError)
	}
}

// /search endpoint.
func (ks *kustomizeSearch) search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()

		queries := values["q"]
		ks.log.Println("Query: ", values)

		var from int
		fromParam := values["from"]
		if len(fromParam) > 0 {
			from, _ = strconv.Atoi(fromParam[0])
			if from < 0 {
				from = 0
			}
		}
		_, noKinds := values["nokinds"]

		opt := index.KustomizeSearchOptions{
			SearchOptions: index.SearchOptions{
				Size: 10,
				From: from,
			},
			KindAggregation: !noKinds,
		}

		results, err := ks.idx.Search(strings.Join(queries, " "), opt)
		if err != nil {
			ks.log.Println("Error: ", err)
			http.Error(w, fmt.Sprintf(
				`{ "error": "could not complete the query" }`),
				http.StatusInternalServerError)
			return
		}

		enc := json.NewEncoder(w)
		setIndent(enc)
		if err = enc.Encode(results); err != nil {
			http.Error(w, `{ "error": "failed to send back results" }`,
				http.StatusInternalServerError)
			return
		}
		return
	}
}

// metrics endpoint.
func (ks *kustomizeSearch) metrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := ks.idx.Search("", index.KustomizeSearchOptions{
			KindAggregation:       true,
			TimeseriesAggregation: true,
		})
		if err != nil {
			http.Error(w, `{ "error": "could not perform the search."}`,
				http.StatusInternalServerError)
			return
		}

		enc := json.NewEncoder(w)
		setIndent(enc)
		if err := enc.Encode(res); err != nil {
			http.Error(w, `{ "error": "could not format return value" }`,
				http.StatusInternalServerError)
			return
		}
	}
}

// make json response human readable.
func setIndent(e *json.Encoder) {
	e.SetIndent("", "  ")
}
