package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"sigs.k8s.io/kustomize/api/internal/crawl/crawler/github"

	"sigs.k8s.io/kustomize/api/internal/crawl/doc"

	"sigs.k8s.io/kustomize/api/internal/crawl/index"
)

const (
	githubAccessTokenVar = "GITHUB_ACCESS_TOKEN"
	retryCount           = 3
)

// iterateArr adds each item in arr into countMap.
func iterateArr(arr []string, countMap map[string]int) {
	for _, item := range arr {
		if _, ok := countMap[item]; !ok {
			countMap[item] = 1
		} else {
			countMap[item]++
		}
	}

}

// SortMapKeyByValue takes a map as its input, sorts its keys according to their values
// in the map, and outputs the sorted keys as a slice.
func SortMapKeyByValue(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	// sort keys according to their values in the map m
	sort.Slice(keys, func(i, j int) bool { return m[keys[i]] > m[keys[j]] })
	return keys
}

func GeneratorOrTransformerStats(ctx context.Context,
	docs []*doc.Document, isGenerator bool, idx *index.KustomizeIndex) {

	fieldName := "generators"
	if !isGenerator {
		fieldName = "transformers"
	}

	// allReferredDocs includes all the documents referred in the field
	allReferredDocs := doc.NewUniqueDocuments()

	// docUsingGeneratorCount counts the number of the kustomization files using generators or transformers
	docCount := 0

	// collect all the documents referred in the field
	for _, d := range docs {
		kdoc := doc.KustomizationDocument{
			Document: *d,
		}
		referredDocs, err := kdoc.GetResources(false, !isGenerator, isGenerator)
		if err != nil {
			log.Printf("failed to parse the %s field of the Document (%s): %v",
				fieldName, d.Path(), err)
		}
		if len(referredDocs) > 0 {
			docCount++
			allReferredDocs.AddDocuments(referredDocs)
		}
	}

	fileCount, dirCount, fileTypeDocs, dirTypeDocs := DocumentTypeSummary(ctx, allReferredDocs.Documents())

	// check whether any of the files are not in the index
	nonExistFileCount := ExistInIndex(idx, fileTypeDocs, fieldName + " file ")
	// check whether any of the dirs are not in the index
	nonExistDirCount := ExistInIndex(idx, dirTypeDocs, fieldName + " dir ")

	GitRepositorySummary(fileTypeDocs, fieldName + " files")
	GitRepositorySummary(dirTypeDocs, fieldName + " dirs")

	fmt.Printf("%d kustomization files use %s: %d %s are files and %d %s are dirs.\n",
		docCount, fieldName, fileCount, fieldName, dirCount, fieldName)
	fmt.Printf("%d %s files do not exist in the index\n", nonExistFileCount, fieldName)
	fmt.Printf("%d %s dirs do not exist in the index\n", nonExistDirCount, fieldName)
}

// GitRepositorySummary counts the distribution of docs:
// 1) how many git repositories are these docs from?
// 2) how many docs are from each git repository?
func GitRepositorySummary(docs []*doc.Document, msgPrefix string) {
	m := make(map[string]int)
	for _, d := range docs {
		if _, ok := m[d.RepositoryURL]; ok {
			m[d.RepositoryURL]++
		} else {
			m[d.RepositoryURL] = 1
		}
	}
	sortedKeys := SortMapKeyByValue(m)
	for _, k := range sortedKeys {
		fmt.Printf("%d %s are from %s\n", m[k], msgPrefix, k)
	}
}

// ExistInIndex goes through each Document in docs, and check whether it is in the index or not.
// It returns the number of documents which does not exist in the index.
func ExistInIndex(idx *index.KustomizeIndex, docs []*doc.Document, msgPrefix string) int {
	nonExistCount := 0
	for _, d := range docs {
		exists, err := idx.Exists(d.ID())
		if err != nil {
			log.Println(err)
		}
		if !exists {
			log.Printf("%s (%s) does not exist in the index", msgPrefix, d.Path())
			nonExistCount++
		}
	}
	return nonExistCount
}

// DocumentTypeSummary goes through each doc in docs, and determines whether it is a file or dir.
func DocumentTypeSummary(ctx context.Context, docs []*doc.Document) (
	fileCount, dirCount int, files, dirs []*doc.Document) {
	githubToken := os.Getenv(githubAccessTokenVar)
	if githubToken == "" {
		log.Fatalf("Must set the variable '%s' to make github requests.\n",
			githubAccessTokenVar)
	}
	ghCrawler := github.NewCrawler(githubToken, retryCount, &http.Client{}, github.QueryWith())

	for _, d := range docs {
		oldFilePath := d.FilePath
		if err := ghCrawler.FetchDocument(ctx, d); err != nil {
			log.Printf("FetchDocument failed on %s: %v", d.Path(), err)
			continue
		}

		if d.FilePath == oldFilePath {
			fileCount++
			files = append(files, d)
		} else {
			dirCount++
			dirs = append(dirs, d)
		}
	}
	return fileCount, dirCount, files, dirs
}

// ExistInSlice checks where target exits in items.
func ExistInSlice(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func main() {
	topKindsPtr := flag.Int(
		"kinds", -1,
		`the number of kubernetes object kinds to be listed according to their popularities.
By default, all the kinds will be listed.
If you only want to list the 10 most popular kinds, set the flag to 10.`)
	topIdentifiersPtr := flag.Int(
		"identifiers", -1,
		`the number of identifiers to be listed according to their popularities.
By default, all the identifiers will be listed.
If you only want to list the 10 most popular identifiers, set the flag to 10.`)
	topKustomizeFeaturesPtr := flag.Int(
		"kustomize-features", -1,
		`the number of kustomize features to be listed according to their popularities.
By default, all the features will be listed.
If you only want to list the 10 most popular features, set the flag to 10.`)
	indexNamePtr := flag.String(
		"index", "kustomize", "The name of the ElasticSearch index.")
	flag.Parse()

	ctx := context.Background()
	idx, err := index.NewKustomizeIndex(ctx, *indexNamePtr)
	if err != nil {
		log.Fatalf("Could not create an index: %v\n", err)
	}

	// count tracks the number of documents in the index
	count := 0

	// kustomizationFilecount tracks the number of kustomization files in the index
	kustomizationFilecount := 0

	kindsMap := make(map[string]int)
	identifiersMap := make(map[string]int)
	kustomizeIdentifiersMap := make(map[string]int)

	// ids tracks the unique IDs of the documents in the index
	ids := make(map[string]struct{})

	// generatorDocs includes all the docs using generators
	generatorDocs := make([]*doc.Document, 0)

	// transformersDocs includes all the docs using transformers
	transformersDocs := make([]*doc.Document, 0)

	// get all the documents in the index
	query := []byte(`{ "query":{ "match_all":{} } }`)
	it := idx.IterateQuery(query, 10000, 60*time.Second)
	for it.Next() {
		for _, hit := range it.Value().Hits.Hits {
			// check whether there is any duplicate IDs in the index
			if _, ok := ids[hit.ID]; !ok {
				ids[hit.ID] = struct{}{}
			} else {
				log.Printf("Found duplicate ID (%s)\n", hit.ID)
			}

			count++
			iterateArr(hit.Document.Kinds, kindsMap)
			iterateArr(hit.Document.Identifiers, identifiersMap)

			if doc.IsKustomizationFile(hit.Document.FilePath) {
				kustomizationFilecount++
				iterateArr(hit.Document.Identifiers, kustomizeIdentifiersMap)
				if ExistInSlice(hit.Document.Identifiers, "generators") {
					generatorDocs = append(generatorDocs, hit.Document.Copy())
				}
				if ExistInSlice(hit.Document.Identifiers, "transformers") {
					transformersDocs = append(transformersDocs, hit.Document.Copy())
				}
			}
		}
	}

	if err := it.Err(); err != nil {
		log.Fatalf("Error iterating: %v\n", err)
	}

	sortedKindsMapKeys := SortMapKeyByValue(kindsMap)
	sortedIdentifiersMapKeys := SortMapKeyByValue(identifiersMap)
	sortedKustomizeIdentifiersMapKeys := SortMapKeyByValue(kustomizeIdentifiersMap)

	fmt.Printf(`The count of unique document IDs in the kustomize index: %d
There are %d documents in the kustomize index.
%d kinds of kubernetes objects are customized:`, len(ids), count, len(kindsMap))
	fmt.Printf("\n")

	kindCount := 0
	for _, key := range sortedKindsMapKeys {
		if *topKindsPtr < 0 || (*topKindsPtr >= 0 && kindCount < *topKindsPtr) {
			fmt.Printf("\tkind `%s` is customimzed in %d documents\n", key, kindsMap[key])
			kindCount++
		}
	}

	fmt.Printf("%d kinds of identifiers are found:\n", len(identifiersMap))
	identifierCount := 0
	for _, key := range sortedIdentifiersMapKeys {
		if *topIdentifiersPtr < 0 || (*topIdentifiersPtr >= 0 && identifierCount < *topIdentifiersPtr) {
			fmt.Printf("\tidentifier `%s` appears in %d documents\n", key, identifiersMap[key])
			identifierCount++
		}
	}

	fmt.Printf(`There are %d kustomization files in the kustomize index.
%d kinds of kustomize features are found:`, kustomizationFilecount, len(kustomizeIdentifiersMap))
	fmt.Printf("\n")
	kustomizeFeatureCount := 0
	for _, key := range sortedKustomizeIdentifiersMapKeys {
		if *topKustomizeFeaturesPtr < 0 || (*topKustomizeFeaturesPtr >= 0 && kustomizeFeatureCount < *topKustomizeFeaturesPtr) {
			fmt.Printf("\tfeature `%s` is used in %d documents\n", key, kustomizeIdentifiersMap[key])
			kustomizeFeatureCount++
		}
	}

	GeneratorOrTransformerStats(ctx, generatorDocs, true, idx)
	GeneratorOrTransformerStats(ctx, transformersDocs, false, idx)
}
