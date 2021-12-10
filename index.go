package deltasearcher

import (
	"container/list"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Index struct {
	Dictionary     map[string]PostingList
	ToTalDocsCount int
}

func (ind Index) String() string {
	var padding int
	keys := make([]string, 0, len(ind.Dictionary))
	for k := range ind.Dictionary {
		l := utf8.RuneCountInString(k)
		if padding < l {
			padding = l
		}
		keys = append(keys, k)
	}

	sort.Strings(keys)
	strs := make([]string, len(keys))
	format := "  [%=" + strconv.Itoa(padding) + "s] -> %s"
	for i, k := range keys {
		if postingList, ok := ind.Dictionary[k]; ok {
			strs[i] = fmt.Sprintf(format, k, postingList.String())
		}
	}

	return fmt.Sprintf("total documents : %v\ndictonary:\n%v\n", ind.ToTalDocsCount, strings.Join(strs, "\n"))
}

func NewIndex() *Index {

	return &Index{
		Dictionary: make(map[string]PostingList),
	}
}

type DocumentID int64

//Postingは一つのDocumentのことを表す、PostingListで複数個のDocumentのことが表せる
type Posting struct {
	DocID         DocumentID
	Positions     []int
	TermFrequency int
}

func (p Posting) String() string {
	return fmt.Sprintf("(%v,%v,%v)", p.DocID, p.TermFrequency, p.Positions)
}

func NewPosting(docID DocumentID, postions ...int) *Posting {
	return &Posting{
		DocID:         docID,
		Positions:     postions,
		TermFrequency: len(postions),
	}
}

type PostingList struct {
	list *list.List
}

func NewPostingList(postings ...*Posting) PostingList {
	l := list.New()
	for _, posting := range postings {
		l.PushBack(posting)
	}

	return PostingList{
		list: l,
	}

}

func (pl PostingList) add(p *Posting) {
	pl.list.PushBack(p)
}

func (pl PostingList) last() *Posting {
	e := pl.list.Back()
	if e == nil {
		return nil
	}

	return e.Value.(*Posting)
}

func (pl PostingList) Add(newP *Posting) {
	last := pl.last()
	if last == nil || last.DocID != newP.DocID {
		pl.add(newP)
		return
	}

	last.Positions = append(last.Positions, newP.Positions...)
	last.TermFrequency++
}

func (pl PostingList) String() string {
	str := make([]string, 0, pl.list.Len())
	for e := pl.list.Front(); e != nil; e = e.Next() {
		str = append(str, e.Value.(*Posting).String())
	}

	return strings.Join(str, "=>")
}

func (pl *PostingList) MarshalJSON() ([]byte, error) {
	postingList := make([]*Posting, 0, pl.list.Len())
	for e := pl.list.Front(); e != nil; e = e.Next() {
		postingList = append(postingList, e.Value.(*Posting))
	}
	return json.Marshal(postingList)
}

func (pl *PostingList) UnmarshalJSON(b []byte) error {
	var postingList []*Posting
	if err := json.Unmarshal(b, &postingList); err != nil {
		return err
	}
	pl.list = list.New()
	for _, posting := range postingList {
		pl.add(posting)
	}
	return nil
}

func (pl PostingList) OpenCursor() *Cursor {
	return &Cursor{
		postingList: &pl,
		current:     pl.list.Front(),
	}
}

//cursorはpostingListを対象にとってそのpostingList内を辿って書くpostingを取り出す
type Cursor struct {
	postingList *PostingList  // cursorが辿っているポスティングリストへの参照
	current     *list.Element //現在の読み込み位置
}

func (c *Cursor) Next() {
	c.current = c.current.Next()
}

//引数のid以上のドキュメントIDになるまでcursorを進める
func (c *Cursor) NextDoc(id DocumentID) {
	for !c.Empty() && c.DocId() < id {
		c.Next()
	}
}

func (c *Cursor) Empty() bool {
	if c.current == nil {
		return true
	}

	return false
}

func (c *Cursor) Posting() *Posting {
	return c.current.Value.(*Posting)
}

func (c *Cursor) DocId() DocumentID {
	return c.Posting().DocID
}

func (c *Cursor) String() string {
	return fmt.Sprint(c.Posting())
}
