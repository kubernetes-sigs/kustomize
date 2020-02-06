package main

import (
	"context"
	"crypto/sha256"
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

// SortMapKeyByValueInt takes a map as its input, sorts its keys according to their values
// in the map, and outputs the sorted keys as a slice.
func SortMapKeyByValueInt(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	// sort keys according to their values in the map m
	sort.Slice(keys, func(i, j int) bool { return m[keys[i]] > m[keys[j]] })
	return keys
}

// SortMapKeyByValue takes a map as its input, sorts its keys according to their values
// in the map, and outputs the sorted keys as a slice.
func SortMapKeyByValueLen(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	// sort keys according to their values in the map m
	sort.Slice(keys, func(i, j int) bool { return len(m[keys[i]]) > len(m[keys[j]]) })
	return keys
}

func GeneratorOrTransformerStats(docs []*doc.KustomizationDocument) {
	n := len(docs)
	if n == 0 {
		return
	}

	fileType := docs[0].FileType
	fmt.Printf("There are totally %d %s files.\n", n, fileType)

	GitRepositorySummary(docs, fileType)

	// key of kindToUrls: a string in the KustomizationDocument.Kinds field
	// value of kindToUrls: a slice of string urls defining a given kind.
	kindToUrls := make(map[string][]string)

	for _, d := range docs {
		url := fmt.Sprintf("%s/blob/%s/%s", d.RepositoryURL, d.DefaultBranch, d.FilePath)
		for _, kind := range d.Kinds {
			if _, ok := kindToUrls[kind]; !ok {
				kindToUrls[kind] = []string{url}
			} else {
				kindToUrls[kind] = append(kindToUrls[kind], url)
			}
		}
	}
	fmt.Printf("There are totally %d kinds of %s\n", len(kindToUrls), fileType)
	sortedKeys := SortMapKeyByValueLen(kindToUrls)
	for _, k := range sortedKeys {
		sort.Strings(kindToUrls[k])
		fmt.Printf("%s kind %s appears %d times\n", fileType, k, len(kindToUrls[k]))
		for _, url := range kindToUrls[k] {
			fmt.Printf("%s\n", url)
		}
	}
}

// GitRepositorySummary counts the distribution of docs:
// 1) how many git repositories are these docs from?
// 2) how many docs are from each git repository?
func GitRepositorySummary(docs []*doc.KustomizationDocument, fileType string) {
	m := make(map[string]int)
	for _, d := range docs {
		if _, ok := m[d.RepositoryURL]; ok {
			m[d.RepositoryURL]++
		} else {
			m[d.RepositoryURL] = 1
		}
	}
	sortedKeys := SortMapKeyByValueInt(m)
	topN := 10
	i := 0
	for _, k := range sortedKeys {
		if i >= topN {
			break
		}
		fmt.Printf("%d %s are from %s\n", m[k], fileType, k)
		i++
	}
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

	// generatorFiles include all the non-kustomization files whose FileType is generator
	generatorFiles := make([]*doc.KustomizationDocument, 0)

	// transformersFiles include all the non-kustomization files whose FileType is transformer
	transformersFiles := make([]*doc.KustomizationDocument, 0)

	checksums := make(map[string]int)

	// get all the documents in the index
	query := []byte(`{ "query":{ "match_all":{} } }`)
	it := idx.IterateQuery(query, 10000, 60*time.Second)
	for it.Next() {
		for _, hit := range it.Value().Hits.Hits {
			sum := fmt.Sprintf("%x", sha256.Sum256([]byte(hit.Document.DocumentData)))
			if _, ok := checksums[sum]; ok {
				checksums[sum]++
			} else {
				checksums[sum] = 1
			}

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

			} else {
				switch hit.Document.FileType {
				case "generator":
					generatorFiles = append(generatorFiles, hit.Document.Copy())
				case "transformer":
					transformersFiles = append(transformersFiles, hit.Document.Copy())
				}
			}
		}
	}

	if err := it.Err(); err != nil {
		log.Fatalf("Error iterating: %v\n", err)
	}

	sortedKindsMapKeys := SortMapKeyByValueInt(kindsMap)
	sortedIdentifiersMapKeys := SortMapKeyByValueInt(identifiersMap)
	sortedKustomizeIdentifiersMapKeys := SortMapKeyByValueInt(kustomizeIdentifiersMap)

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

	GeneratorOrTransformerStats(generatorFiles)
	GeneratorOrTransformerStats(transformersFiles)

	fmt.Printf("There are total %d checksums of document contents\n", len(checksums))
	sortedChecksums := SortMapKeyByValueInt(checksums)
	sortedChecksums = sortedChecksums[:20]
	fmt.Printf("The top 20 checksums are:\n")
	for _, key := range sortedChecksums {
		fmt.Printf("checksum %s apprears %d\n", key, checksums[key])
	}
}
