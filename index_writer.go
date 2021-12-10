package deltasearcher

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type IndexWriter struct {
	indexDir string
}

func NewIndexWriter(path string) *IndexWriter {
	return &IndexWriter{
		indexDir: path,
	}
}

func (w *IndexWriter) Flush(index *Index) error {
	for term, positionList := range index.Dictionary {
		//ファイルへの書き込みを中止するなりなんなりしないとダメそう、
		//tempFile => renameとして全部書き込めたらrename,失敗したらrenameせずにtempFile削除でもいいかもしれない
		if err := w.PostingFlush(term, positionList); err != nil {
			fmt.Printf("failed to save %s postingList: %v", term, err)
			return err
		}
	}

	return w.docCount(index.ToTalDocsCount)
}

func (w *IndexWriter) docCount(count int) error {
	filename := filepath.Join(w.indexDir, "_0.dc")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write([]byte(strconv.Itoa(count)))
	return err
}

func (w *IndexWriter) PostingFlush(term string, list PostingList) error {
	bytes, err := json.Marshal(&list)
	if err != nil {
		return err
	}

	filename := filepath.Join(w.indexDir, term)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.Write(bytes)
	if err != nil {
		return err
	}

	return writer.Flush()
}
