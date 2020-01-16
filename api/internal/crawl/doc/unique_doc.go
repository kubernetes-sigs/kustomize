package doc

import (
	"sigs.k8s.io/kustomize/api/internal/crawl/utils"
)

// UniqueDocuments make sure a Document with a given ID appears only once
type UniqueDocuments struct {
	docs   []*Document
	docIDs utils.SeenMap
}

func NewUniqueDocuments() UniqueDocuments {
	return UniqueDocuments{
		docs:   []*Document{},
		docIDs: utils.NewSeenMap(),
	}
}

func (uds *UniqueDocuments) Add(d *Document) {
	if uds.docIDs.Seen(d.ID()) {
		return
	}
	uds.docs = append(uds.docs, d)
	uds.docIDs.Add(d.ID())
}

func (uds *UniqueDocuments) AddDocuments(docs []*Document) {
	for _, d := range docs {
		uds.Add(d)
	}
}

func (uds *UniqueDocuments) Documents() []*Document {
	return uds.docs
}
