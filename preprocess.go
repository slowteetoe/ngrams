package main

import (
	"encoding/csv"
	_ "gopkg.in/cq.v1"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type ngram struct {
	phrase   string
	nextWord string
	freq 	 int
	prob     float64
}

type ngramHolder struct {
	phrase string
	topN   Ngrams
	n      int
	min    float64
}

type Ngrams []ngram

func (slice Ngrams) Len() int {
	return len(slice)
}

func (slice Ngrams) Less(i, j int) bool {
	return slice[i].prob <= slice[j].prob
}

func (slice Ngrams) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// phrase, freq, probablility
func parseCSVRecord(record []string) ngram {
	var phrase string
	var nextWord string
	i := strings.LastIndex(record[0], " ")
	if i == -1 {
		log.Fatalf("Could not find a space in %v, this seems likely to be a fatal error", record[0])
	}

	phrase = record[0][0:i]
	nextWord = record[0][i+1:]

	freq, err := strconv.Atoi(record[1])
	if err != nil {
		log.Fatal(err)
	}

	prob, err := strconv.ParseFloat(record[2], 64)
	if err != nil {
		log.Fatal(err)
	}

	return ngram{phrase: phrase, nextWord: nextWord, freq: freq, prob: prob}
}

func (h *ngramHolder) Add(n ngram) {

	// decide whether to add it or not
	if n.prob > h.min {
		// log.Printf("We want to add %v at %v\n", n.nextWord, h.n)
		// add it
		h.topN[h.n] = n
		h.n = h.n + 1
		if h.min == -9999 {
			h.min = n.prob
		}
		sort.Sort(h.topN)
		t := make([]ngram, 10)
		copy(t, h.topN[:5])
		h.topN = t
		if h.n > 6 {
			h.n = 5
		}
	}
}

func main() {

	files := []string{"bigrams.csv", "trigrams.csv", "fourgrams.csv"}

	for _, f := range files {

		file, err := os.Open(f)

		if err != nil {
			log.Printf("Skipping %v due to %v\n", f, err)
			continue
		}
		defer file.Close()

		reader := csv.NewReader(file)
		lineCount := 0

		outfile, err := os.OpenFile("output.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)

	    if err != nil {panic(err)}
	    defer outfile.Close()

	    writer := csv.NewWriter(outfile)
	    defer writer.Flush()

		thisHolderList := make([]ngram, 10)
		holder := ngramHolder{min: -9999, n: 0, topN: thisHolderList}
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}

			thisngram := parseCSVRecord(record)

			if holder.phrase != thisngram.phrase {
				if holder.phrase != "" {
					for _, data := range holder.topN {
						if data.phrase == "" {
							continue
						}
						// loop over the top ngrams for this phrase and write them to the output file
						var record []string
						record = append(record, data.phrase + " " + data.nextWord)
	        			record = append(record, strconv.Itoa(data.freq))
	        			record = append(record, strconv.FormatFloat(data.prob, 'f', 6, 64))
						writer.Write(record)
					}
				}
				thisHolderList := make([]ngram, 10)
				holder = ngramHolder{min: -9999, n: 0, phrase: thisngram.phrase, topN: thisHolderList}
			}

			holder.Add(thisngram)

			// insert logic here
			lineCount += 1
		}
	}
}
