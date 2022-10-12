package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	p "github.com/arunsworld/pipeline"
)

func main() {
	bookToRead := flag.String("b", "pg2600.txt", "book to analyze")
	concurrency := flag.Int("c", 10, "concurrency")
	flag.Parse()

	if err := run(*bookToRead, *concurrency); err != nil {
		log.Fatal(err)
	}

}

func run(name string, concurrency int) error {
	b, err := readBook(name)
	if err != nil {
		return err
	}

	result, _ := p.New(
		p.Concurrency[*paragraph](concurrency),
		p.PreFilter(filterWarAndPeaceSuperfluousLines),
		p.MustTransform(splitParaIntoWords()),
		p.MustTransform(removeStopWords()),
	).ApplyAndFold(b, p.NewMustFoldOperation(summarize))

	fmt.Println(result.String())

	return nil
}

func filterWarAndPeaceSuperfluousLines(v *paragraph) bool {
	if v.count < 395 {
		return false
	}
	if v.count > 12116 {
		return false
	}
	return true
}

func splitParaIntoWords() func(*paragraph) *paragraph {
	replacer := strings.NewReplacer(`"`, "", "?", "", "'", "", ",", "", ".", "", `“`, "", "(", "", ")", "",
		"”", "", "‘", "", "’", "", "|", "", ";", "")
	return func(v *paragraph) *paragraph {
		content := strings.ToLower(v.content)
		content = replacer.Replace(content)
		words := strings.Split(content, " ")
		for _, w := range words {
			w = strings.TrimSpace(w)
			if w == "" {
				continue
			}
			v.words[w] += 1
			v.totalWords++
		}
		return v
	}
}

func removeStopWords() func(*paragraph) *paragraph {
	stopWords := []string{"the", "and", "to", "of", "a", "he", "in", "his", "that", "was", "an",
		"with", "had", "at", "not", "her", "as", "it", "on", "but", "for", "i", "she", "is", "him", "you", "from", "were", "all",
		"said", "by", "were", "be", "they", "who", "what", "have", "which", "one", "this", "so", "or", "been", "their", "did", "when",
		"would", "up", "only", "are", "if", "my", "could", "there", "no", "out", "will", "them", "how", "we", "do", "has",
		"himself", "without", "before", "because"}
	return func(v *paragraph) *paragraph {
		wrds := v.words
		for _, w := range stopWords {
			delete(wrds, w)
		}
		v.words = wrds
		return v
	}
}

func summarize(a, b *paragraph) *paragraph {
	a.totalWords += b.totalWords
	for w, c := range b.words {
		a.words[w] += c
	}
	return a
}

type book = []*paragraph

type paragraph struct {
	count      int
	content    string
	words      map[string]int
	totalWords int
}

type wordCount struct {
	word  string
	count int
}

func (p *paragraph) topNWords(n int) []wordCount {
	result := make([]wordCount, 0, len(p.words))
	for w, c := range p.words {
		if len(w) < 5 {
			continue
		}
		result = append(result, wordCount{word: w, count: c})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].count > result[j].count
	})
	return result[:n]
}

func (p *paragraph) String() string {
	result := strings.Builder{}
	result.WriteString("--------------------\n")
	result.WriteString(fmt.Sprintf("Total words: %d\n\n", p.totalWords))
	top := 20
	result.WriteString(fmt.Sprintf("Top %d words: \n", top))
	for _, v := range p.topNWords(top) {
		result.WriteString(fmt.Sprintf("\t%s: %d\n", v.word, v.count))
	}
	result.WriteString("--------------------")
	return result.String()
}

func readBook(name string) (book, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return newBookFrom(f), nil
}

func newBookFrom(f io.Reader) book {
	result := make([]*paragraph, 0, 100)
	scanner := bufio.NewScanner(f)
	prevLine := make([]string, 0, 30)
	for scanner.Scan() {
		txt := strings.TrimSpace(scanner.Text())
		switch {
		case txt == "" && len(prevLine) > 0:
			para := strings.Join(prevLine, " ")
			result = append(result, &paragraph{count: len(result), content: para, words: make(map[string]int)})
			prevLine = make([]string, 0, 30)
		case txt == "":
		default:
			prevLine = append(prevLine, txt)
		}
	}
	if len(prevLine) > 0 {
		para := strings.Join(prevLine, " ")
		result = append(result, &paragraph{content: para, words: make(map[string]int)})
	}
	return result
}
