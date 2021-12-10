package deltasearcher

import (
	"reflect"
	"strings"
	"testing"
)

func TestEnUpdate(t *testing.T) {
	collection := []string{
		"Do you quarrel, sir?",
		"Quarrel sir! no, sir!",
		"No better.",
		"Well, sir",
	}
	indexer := NewIndexer(NewEnTokenizer())

	for i, doc := range collection {
		indexer.update(DocumentID(i), strings.NewReader(doc))

	}

	actual := indexer.index
	expected := &Index{
		Dictionary: map[string]PostingList{
			"better": NewPostingList(
				NewPosting(2, 1)),
			"do": NewPostingList(
				NewPosting(0, 0)),
			"no": NewPostingList(
				NewPosting(1, 2),
				NewPosting(2, 0)),
			"quarrel": NewPostingList(
				NewPosting(0, 2),
				NewPosting(1, 0)),
			"sir": NewPostingList(NewPosting(0, 3),
				NewPosting(1, 1, 3),
				NewPosting(3, 1)),
			"well": NewPostingList(
				NewPosting(3, 0)),
			"you": NewPostingList(
				NewPosting(0, 1)),
		},
		ToTalDocsCount: 4,
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("diff is :\nactual: %v\nexpected: %v\n", actual, expected)
	}
}
