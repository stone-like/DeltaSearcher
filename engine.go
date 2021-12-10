package deltasearcher

import (
	"io"
	"os"
	"path/filepath"
)

type Engine struct {
	tokenizer     Tokenizer
	indexer       *Indexer
	documentStore *DocumentStore
	indexDir      string
}

//ここのDBは抽象化してStoreにして良さそう、Tokenizerも多分抽象化して良さそう
func NewEngine(indexer *Indexer, documentStore *DocumentStore) *Engine {
	path, ok := os.LookupEnv("INDEX_DIR_PATH")
	if !ok {
		curDir, _ := os.Getwd()
		path = filepath.Join(curDir, "_index_data")
	}
	return &Engine{
		tokenizer:     indexer.tokenizer,
		indexer:       indexer,
		documentStore: documentStore,
		indexDir:      path,
	}
}

func (e *Engine) AddDocument(title string, r io.Reader) error {
	id, err := e.documentStore.save(title)
	if err != nil {
		return err
	}

	e.indexer.update(id, r)
	return nil
}

func (e *Engine) Flush() error {
	w := NewIndexWriter(e.indexDir)
	return w.Flush(e.indexer.index)
}

func (e *Engine) Search(query string, k int, indexReader *IndexReader) ([]*SearchResult, error) {
	terms := e.tokenizer.TextToWordSequence(query)

	docs := NewSearcher(indexReader).SearchTopK(terms, k)

	results := make([]*SearchResult, 0, k)
	for _, result := range docs.scoreDocs {
		title, err := e.documentStore.fetchTitle(result.docID)
		if err != nil {
			return nil, err
		}
		results = append(results, &SearchResult{
			result.docID, result.score, title,
		})
	}
	return results, nil
}

type SearchResult struct {
	DocID DocumentID
	Score float64
	Title string
}
