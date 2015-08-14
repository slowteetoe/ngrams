package main

import (
    "database/sql"
    "log"
    "encoding/csv"
    "io"
    "os"
    "strings"
    "strconv"
    _ "gopkg.in/cq.v1"
)

func prepareInsert(tx *sql.Tx) *sql.Stmt {
    combined, err := tx.Prepare(`merge(n:Ngram { phrase: {0} } ) create unique (n)-[:PRECEDED {p: {2} }]->(nw:NextWord { word: {1} } )`)
    if err != nil {
        log.Fatal(err)
    }
    return combined
}

type ngram struct {
    phrase string
    nextWord string
    prob float64
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

    prob, err := strconv.ParseFloat(record[2], 64)
    if err != nil {
        log.Fatal(err)
    }

    return ngram{ phrase: phrase, nextWord: nextWord, prob: prob}
}


// Don't forget to run
// MATCH (n) OPTIONAL MATCH (n)-[r]-() DELETE n,r
// create index on :Ngram(phrase)
func main() {

    db, err := sql.Open("neo4j-cypher", "http://neo4j:n304j@localhost:7474")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    files := []string{"bigrams.csv","trigrams.csv","fourgrams.csv"}

    for _, f := range files {

        file, err := os.Open(f)

        if err != nil {
            log.Printf("Skipping %v due to %v\n", f, err)
            continue
        }
        defer file.Close()

        tx, err := db.Begin()
        if err != nil {
            log.Fatal(err)
        }

        combined := prepareInsert(tx)
        defer combined.Close()

        reader := csv.NewReader(file)
        lineCount := 0
        batchSize := 5000
        for {
            record, err := reader.Read()
            if err == io.EOF {
                break
            } else if err != nil {
                log.Fatal(err)
            }

            ngram := parseCSVRecord(record)
            _, err = combined.Exec(ngram.phrase, ngram.nextWord, ngram.prob)
            
            if err != nil {
                log.Fatal(err)
            }
            
            lineCount += 1
            if lineCount % batchSize == 0 {
                log.Printf("Committing %v entries and starting a new transaction, current total is: %v\n", batchSize, lineCount)
                err = tx.Commit()
                if err != nil {
                    log.Fatal(err)
                }
                tx, err = db.Begin()
                if err != nil {
                    log.Fatal(err)
                }
                combined = prepareInsert(tx)
            }
        }
        err = tx.Commit()
        if err != nil {
            log.Fatal(err)
        }
        log.Printf("Recorded %v records\n", lineCount)
    }
}