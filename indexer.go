package deltasearcher

import (
	"io"
)

type Indexer struct {
	index     *Index
	tokenizer Tokenizer
}

func NewIndexer(tokenizer Tokenizer) *Indexer {
	return &Indexer{
		index:     NewIndex(),
		tokenizer: tokenizer,
	}
}

func (i *Indexer) splitTerm(reader io.Reader) []string {
	return i.tokenizer.SplitTerm(reader)
}

func (i *Indexer) update(docID DocumentID, reader io.Reader) {

	terms := i.splitTerm(reader)

	for pos, term := range terms {
		if postingList, ok := i.index.Dictionary[term]; !ok {
			i.index.Dictionary[term] = NewPostingList(NewPosting(docID, pos))
		} else {
			postingList.Add(NewPosting(docID, pos))
		}
	}

	i.index.ToTalDocsCount++
}
