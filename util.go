package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/sajari/fuzzy"
)

// SplitOnNonLetters splits a string on non-letter runes
func SplitOnNonLetters(s string) []string {
	notALetter := func(char rune) bool { return !unicode.IsLetter(char) }
	return strings.FieldsFunc(s, notALetter)
}

func getWordTokens(s string) []string {
	str := strings.ToLower(s)
	parts := SplitOnNonLetters(str)
	return parts
}

func ngrams(words []string, size int) (count map[string]uint64) {

	count = make(map[string]uint64, 0)
	offset := int(math.Floor(float64(size / 2)))

	max := len(words)
	for i := range words {
		if i < offset || i+size-offset > max {
			continue
		}
		gram := strings.Join(words[i-offset:i+size-offset], " ")
		count[gram]++
	}

	return count
}

func loadDirAsStr(path string) string {
	buf := bytes.NewBuffer(nil)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fb, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", path, file.Name()))
		if err != nil {
			log.Fatal(err)
		}
		buf.Write(fb)
		buf.WriteString(" \n")
	}

	s := string(buf.Bytes())

	return s
}

func generateCsv(filepath string, data map[string]uint64) {
	file, err := os.Create(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for k, v := range data {
		err := writer.Write([]string{k, fmt.Sprintf("%v", v)})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func loadCsv(filepath string) (map[string]uint64, []string) {
	var res map[string]uint64
	var words []string

	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}

	r := csv.NewReader(bufio.NewReader(f))

	res = make(map[string]uint64)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if len(record) > 1 {
			i, err := strconv.ParseUint(record[1], 10, 64)
			if err == nil {
				res[record[0]] = i
			}
			words = append(words, record[0])
		}
	}

	return res, words
}

// possibilities only generate deletion
func loadFuzzy(words []string) *fuzzy.Model {
	model := fuzzy.NewModel()

	// For testing only, this is not advisable on production
	model.SetThreshold(4)

	// This expands the distance searched, but costs more resources (memory and time).
	// For spell checking, "2" is typically enough, for query suggestions this can be higher
	model.SetDepth(2)

	model.Train(words)

	// Train word by word (typically triggered in your application once a given word is popular enough)
	model.TrainWord("single")

	return model
}

// Tripple represents observed and intended word occurences
type Tripple struct {
	ObservedWord string
	IntendedWord string
	Count        uint64
}

// Tripples represents slice of Tripple
type Tripples []Tripple

func (c Tripples) Len() int           { return len(c) }
func (c Tripples) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c Tripples) Less(i, j int) bool { return c[i].Count > c[j].Count }

// Split represents word split
type Split struct {
	L string
	R string
}

const letters = "abcdefghijklmnopqrstuvwxyz"

func generateTripples(tl map[string]uint64, word string) Tripples {
	var res Tripples
	var trp Tripple
	words := edits1(word)

	for _, w := range words {
		v, ok := tl[w]
		if ok {
			trp = Tripple{
				ObservedWord: word,
				IntendedWord: w,
				Count:        v,
			}

			res = append(res, trp)
		}
	}

	sort.Sort(res)

	return res
}

func generateTripples2(tl map[string]uint64, word string) Tripples {
	var res Tripples
	var trp Tripple
	words := edits2(word)

	for _, w := range words {
		v, ok := tl[w]
		if ok {
			trp = Tripple{
				ObservedWord: word,
				IntendedWord: w,
				Count:        v,
			}

			res = append(res, trp)
		}
	}

	sort.Sort(res)

	return res
}

func edits1(word string) []string {
	deletes := make(chan []string)
	transposes := make(chan []string)
	replaces := make(chan []string)
	inserts := make(chan []string)

	splits := getSplits(word)

	go getDeletes(word, splits, deletes)
	go getTransposes(word, splits, transposes)
	go getReplaces(word, splits, letters, replaces)
	go getInserts(word, splits, letters, inserts)

	words := mergeResult(deletes, transposes, replaces, inserts)

	return distinct(words)
}

func edits2(word string) []string {
	words := edits1(word)

	var words2 []string
	for _, v := range words {
		if v != word {
			words2 = edits1(v)
			words = append(words, words2...)
		}
	}

	return distinct(words)
}

func distinct(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}

func getSplits(word string) []Split {
	var res []Split
	for i := 0; i < len(word)+1; i++ {
		spl := Split{
			L: word[:i],
			R: word[i:],
		}
		res = append(res, spl)
	}

	return res
}

func getTransposes(word string, splits []Split, c chan []string) {
	defer close(c)

	var intended string
	var res []string
	// transposes = [L + R[1] + R[0] + R[2:] for L, R in splits if len(R)>1]
	for _, spl := range splits {
		if len(spl.R) > 1 {
			intended = spl.L + string(spl.R[1]) + string(spl.R[0]) + spl.R[2:]
			res = append(res, intended)
		}
	}

	if c != nil && len(res) > 0 {
		c <- res
	}
}

func getReplaces(word string, splits []Split, letters string, c chan []string) {
	defer close(c)

	var intended string
	var res []string
	// replaces   = [L + c + R[1:]           for L, R in splits if R for c in letters]
	for _, spl := range splits {
		for _, c := range letters {
			if len(spl.R) > 0 {
				intended = spl.L + string(c) + spl.R[1:]
				res = append(res, intended)
			}
		}
	}

	if c != nil && len(res) > 0 {
		c <- res
	}
}

func getInserts(word string, splits []Split, letters string, c chan []string) {
	defer close(c)

	var intended string
	var res []string
	// inserts    = [L + c + R               for L, R in splits for c in letters]
	for _, spl := range splits {
		for _, c := range letters {
			intended = spl.L + string(c) + spl.R
			res = append(res, intended)
		}
	}

	if c != nil && len(res) > 0 {
		c <- res
	}
}

func getDeletes(word string, splits []Split, c chan []string) {
	defer close(c)

	var intended string
	var res []string
	// deletes    = [L + R[1:]               for L, R in splits if R]
	for _, spl := range splits {
		if len(spl.R) > 0 {
			intended = spl.L + spl.R[1:]
			res = append(res, intended)
		}
	}

	if c != nil && len(res) > 0 {
		c <- res
	}
}

func mergeResult(deletes, transposes, replaces, inserts chan []string) []string {
	var res []string

	res = append(res, <-deletes...)
	res = append(res, <-transposes...)
	res = append(res, <-replaces...)
	res = append(res, <-inserts...)

	return res
}
