package main

import (
    "database/sql"
    "log"

    _ "gopkg.in/cq.v1"
)

func main() {
    db, err := sql.Open("neo4j-cypher", "http://neo4j:n304j@localhost:7474")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    stmt, err := db.Prepare(`
        match (a {phrase: {0} })-[p:PRECEDED]->(n) return n.word,p.p order by p.p desc limit 5
    `)
    if err != nil {
        log.Fatal(err)
    }
    defer stmt.Close()

    q := "^thanks"

    rows, err := stmt.Query(q)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    var nextWord string
    var prob string
    for rows.Next() {
        err := rows.Scan(&nextWord, &prob)
        if err != nil {
            log.Fatal(err)
        }
        log.Printf("%v -> %v (with probability %v)",q, nextWord, prob)
    }
}

// match (a {phrase:"^this"})-[p:PRECEDED]->(n) return a,n,p.p order by p.p desc
// create (startingThis:Ngram {phrase: "^this", length:1})
// match (a {phrase:"^this"}) create a-[:PRECEDED {p: -3}]->(rocks:NextWord {word:"rocks"})

// insert all the unigrams