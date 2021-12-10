package deltasearcher

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

type IndexReader struct {
	indexDir      string
	postingCache  map[string]*PostingList
	docCountCache int
}

func NewIndexReader(path string) *IndexReader {
	return &IndexReader{
		indexDir:      path,
		postingCache:  make(map[string]*PostingList),
		docCountCache: -1,
	}
}

func (r *IndexReader) postingLists(terms []string) []*PostingList {
	postingLists := make([]*PostingList, 0, len(terms))
	for _, term := range terms {
		if postingList := r.postingList(term); postingList != nil {
			postingLists = append(postingLists, postingList)
		}
	}

	return postingLists
}

func (r *IndexReader) postingList(term string) *PostingList {
	if postingList, ok := r.postingCache[term]; ok {
		return postingList
	}

	filename := filepath.Join(r.indexDir, term)
	file, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil
	}
	var postingList PostingList
	if err = json.Unmarshal(bytes, &postingList); err != nil {
		return nil
	}

	r.postingCache[term] = &postingList
	return &postingList
}

func (r *IndexReader) totalDocCount() int {
	//docCountCacheの初期値は-1としているので0以上ならもうCache済み
	if r.docCountCache > 0 {
		return r.docCountCache
	}

	filename := filepath.Join(r.indexDir, "_0.dc")
	file, err := os.Open(filename)
	if err != nil {
		return 0
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return 0
	}

	count, err := strconv.Atoi(string(bytes))
	if err != nil {
		return 0
	}
	r.docCountCache = count
	return count
}
