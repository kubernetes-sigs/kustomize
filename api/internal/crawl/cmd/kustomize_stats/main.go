package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"
	"time"

	"sigs.k8s.io/kustomize/api/internal/crawl/doc"

	"sigs.k8s.io/kustomize/api/internal/crawl/index"
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
	flag.Parse()

	ctx := context.Background()
	idx, err := index.NewKustomizeIndex(ctx)
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

	// get all the documents in the index
	query := []byte(`{ "query":{ "match_all":{} } }`)
	it := idx.IterateQuery(query, 10000, 60*time.Second)
	for it.Next() {
		for _, hit := range it.Value().Hits.Hits {
			// check whether there is any duplicate IDs in the index
			if _, ok := ids[hit.ID]; !ok {
				ids[hit.ID] = struct{}{}
			} else {
				fmt.Printf("Found duplicate ID (%s)\n", hit.ID)
			}

			count++
			iterateArr(hit.Document.Kinds, kindsMap)
			iterateArr(hit.Document.Identifiers, identifiersMap)

			if doc.IsKustomizationFile(hit.Document.FilePath) {
				kustomizationFilecount++
				iterateArr(hit.Document.Identifiers, kustomizeIdentifiersMap)
			}
		}
	}
	if err := it.Err(); err != nil {
		fmt.Printf("Error iterating: %v\n", err)
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
}
