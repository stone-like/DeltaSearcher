package deltasearcher

import (
	"fmt"
	"math"
	"sort"
)

type TopDocs struct {
	totalHits int
	scoreDocs []*ScoreDoc
}

func (t *TopDocs) String() string {
	return fmt.Sprintf("\ntotal hits: %v\nresults: %v\n", t.totalHits, t.scoreDocs)
}

type ScoreDoc struct {
	docID DocumentID
	score float64
}

func (d ScoreDoc) String() string {
	return fmt.Sprintf("docId: %v,Score: %v", d.docID, d.score)
}

type Searcher struct {
	indexReader *IndexReader
	cursors     []*Cursor //現在辿っているポスティングリストの位置
}

func NewSearcher(indexReader *IndexReader) *Searcher {
	return &Searcher{
		indexReader: indexReader,
	}
}

func (s *Searcher) SearchTopK(query []string, k int) *TopDocs {
	results := s.search(query)

	//スコアの降順でソート
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	total := len(results)

	if len(results) > k {
		results = results[:k]
	}

	return &TopDocs{
		totalHits: total,
		scoreDocs: results,
	}
}

func (s *Searcher) search(query []string) []*ScoreDoc {
	//まずqueryごとにpostingListをファイルからとってきて、postingListからそれぞれのcursorをつくる
	if s.openCursors(query) == 0 {
		return []*ScoreDoc{}
	}

	//openCursorの時にpostingListの長さで短い順にpostingListsをソートしているのでcursorsも短い順となる

	c := s.cursors[0]
	cursors := s.cursors[1:]

	docs := make([]*ScoreDoc, 0)

	//共通で登場するDoucmentgだけしかほしくない＋score順に並び変えたい

	//TFスコア等を計算する必要があるのですべてのcursorを上手く動かす必要がある
	//つまりA,B,Cを検索して
	//A {Doc1,Doc4}
	//B {Doc1,Doc3,Doc4,Doc5}
	//C {Doc1,Doc3,Doc5,Doc6,Doc7}
	//とすると、AのDoc1を読むときにその他のcursorはDoc1以上にしなくてはいけない

	for !c.Empty() {
		var nextDocId DocumentID

		for _, cursor := range cursors {

			cursor.NextDoc(c.DocId())

			if cursor.Empty() {
				return docs
			}

			if cursor.DocId() != c.DocId() {
				nextDocId = cursor.DocId()
				break
			}
		}

		//一個でもcursor.DocId() != c.DocId()なものがあるとnextDocId != 0となる
		if nextDocId > 0 {
			//最も短いpostingリストを読み進める
			c.NextDoc(nextDocId)
			if c.Empty() {
				return docs
			}
		} else {
			//ここまで来るとすべてのcursorが同じDocumentで一致している？
			docs = append(docs, &ScoreDoc{
				docID: c.DocId(),
				score: s.calcScore(),
			})
			c.Next()
		}

	}

	return docs
}

//postingListそれぞれからそれぞれのcursorをつくる
func (s *Searcher) openCursors(query []string) int {
	postingLists := s.indexReader.postingLists(query) //複数のpostingList
	if len(postingLists) == 0 {
		return 0
	}

	sort.Slice(postingLists, func(i, j int) bool {
		return postingLists[i].list.Len() < postingLists[j].list.Len()
	})

	cursors := make([]*Cursor, len(postingLists))
	for i, postingList := range postingLists {
		cursors[i] = postingList.OpenCursor()
	}
	s.cursors = cursors
	return len(cursors)

}

func (s *Searcher) calcScore() float64 {
	var score float64
	for i := 0; i < len(s.cursors); i++ {
		termFreq := s.cursors[i].Posting().TermFrequency
		docCount := s.cursors[i].postingList.list.Len()
		totalDocCount := s.indexReader.totalDocCount()
		score += calcTF(termFreq) * calcIDF(totalDocCount, docCount)
	}

	return score
}

func calcTF(termCount int) float64 {
	if termCount <= 0 {
		return 0
	}

	return math.Log2(float64(termCount)) + 1
}

func calcIDF(N, df int) float64 {
	return math.Log2(float64(N) / float64(df))
}
