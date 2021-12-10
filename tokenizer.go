package deltasearcher

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"unicode"
)

type Tokenizer interface {
	SplitTerm(r io.Reader) []string
	TextToWordSequence(text string) []string
}

func TextToWordSequence(text string, t Tokenizer) []string {
	return t.SplitTerm(strings.NewReader(text))
}

type JpTokenizer struct{}

func NewJpTokenizer() *JpTokenizer {
	return &JpTokenizer{}
}

func (t *JpTokenizer) SplitTerm(r io.Reader) []string {
	return []string{}
}

func (t *JpTokenizer) TextToWordSequence(text string) []string {
	return TextToWordSequence(text, t)
}

type EnTokenizer struct{}

func NewEnTokenizer() *EnTokenizer {
	return &EnTokenizer{}
}

func isAlpha(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

func replace(r rune) rune {
	//英数字以外なら捨てる
	if !isAlpha(r) && !unicode.IsNumber(r) {
		return -1
	}
	//大文字から小文字へ
	return unicode.ToLower(r)
}

func (t *EnTokenizer) SplitTerm(reader io.Reader) []string {
	scanner := bufio.NewScanner(reader)
	scanner.Split(t.SplitFunc)

	var terms []string

	for scanner.Scan() {
		term := scanner.Text()
		terms = append(terms, term)
	}
	return terms
}

func (t *EnTokenizer) SplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advance, token, err = bufio.ScanWords(data, atEOF) //空白区切り

	//空白で区切ってquarrel,みたいなやつはbytes.mapの引数に入っているreplace関数で変換されるときに,英数字以外は捨てられ、
	//quarrel, -> quarrelとなる
	if err == nil && token != nil {
		token = bytes.Map(replace, token)
		if len(token) == 0 {
			token = nil
		}
	}

	return
}

func (t *EnTokenizer) TextToWordSequence(text string) []string {
	return TextToWordSequence(text, t)
}
