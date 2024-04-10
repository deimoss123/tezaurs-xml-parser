package main

import (
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

type Entry struct {
	// XMLName xml.Name `xml:"entry"`
	SortKey string  `xml:"sortKey,attr"`
	Id      string  `xml:"id,attr"`
	N       int     `xml:"n,attr"`
	Type    string  `xml:"type,attr"` // abbr, affix, foreign, main, mwe
	Sense   []Sense `xml:"sense"`
}

type Sense struct {
	// XMLName xml.Name `xml:"sense"`
	Id    string  `xml:"id,attr"`
	N     int     `xml:"n,attr"`
	Def   string  `xml:"def"`
	Sense []Sense `xml:"sense"`
}

type Measure struct {
	Unit     string `xml:"unit,attr"`
	Quantity string `xml:"quantity,attr"`
}

type Extent struct {
	Measure []Measure `xml:"measure"`
}

var (
	entryQty  = 0
	senseQty  = 0
	lexemeQty = 0
)

var defCount = 0

const delimiter = "    "

// fzf --delimiter "    " --nth 2 --with-nth 2,3 --literal --no-hscroll < def.txt
func writeDefinitions(defFile *os.File, id string, sortKey string, sense []Sense) {
	for _, s := range sense {
		// fmt.Printf("%s: %s: %s\n", id, sortKey, s.Def)
		_, err := defFile.WriteString(fmt.Sprintf("%s%s%s%s%s\n", id, delimiter, sortKey, delimiter, s.Def))
		if err != nil {
			fmt.Println("Error writing to file: ", err)
		}
		defCount++
		if len(s.Sense) > 0 {
			writeDefinitions(defFile, id, sortKey, s.Sense)
		}
	}
}

var (
	entryQuery = `INSERT INTO entries (id, n, type, sort_key) VALUES ($1, $2, $3, $4)`
	senseQuery = `INSERT INTO senses (id, n, def, parent_id, entry_id) VALUES ($1, $2, $3, $4, $5)`
)

func addEntryToBatch(batch *pgx.Batch, entry Entry) {
	batch.Queue(entryQuery, entry.Id, entry.N, entry.Type, entry.SortKey)
	addSensesToBatch(batch, entry.Sense, nil, entry.Id)
}

func addSensesToBatch(batch *pgx.Batch, senses []Sense, parentId any, entryId string) {
	for _, s := range senses {
		batch.Queue(senseQuery, s.Id, s.N, s.Def, parentId, entryId)
		if len(s.Sense) > 0 {
			addSensesToBatch(batch, s.Sense, s.Id, entryId)
		}
	}
}

func createTables(conn *pgx.Conn) {
	_, err := conn.Exec(context.Background(), `
		DROP TABLE IF EXISTS entries CASCADE;
		DROP TABLE IF EXISTS senses;

		CREATE EXTENSION IF NOT EXISTS pg_trgm;

		CREATE TABLE entries (
			id TEXT PRIMARY KEY,
			n INTEGER NOT NULL,
			type TEXT NOT NULL,
			sort_key TEXT NOT NULL
		);

		CREATE TABLE senses (
			id TEXT PRIMARY KEY,
			n INTEGER NOT NULL,
			def TEXT NOT NULL,
			parent_id TEXT REFERENCES senses(id),
			entry_id TEXT REFERENCES entries(id)
		);
		
		CREATE INDEX def_gin_trgm_idx ON senses USING gin(def gin_trgm_ops);
		CREATE INDEX sort_key_gin_trgm_idx ON entries USING gin(sort_key gin_trgm_ops);
	`)
	// ALTER TABLE Sense ADD COLUMN entry_id TEXT REFERENCES Entry(id);
	if err != nil {
		log.Fatal("Error creating tables: ", err)
	}
}

func main() {
	filename := flag.String("f", "", "Tezaurs file name")
	writeToFile := flag.Bool("w", false, "Write to file")
	pgString := flag.String("pg", "", "Postgres connection string")
	pgCreateTable := flag.Bool("table", false, "Create tables in db")

	flag.Parse()

	tezFile, err := os.Open(*filename)

	if err != nil {
		log.Fatal("Error opening file: ", err)
	}
	fmt.Printf("Opened file %s\n", *filename)

	defer tezFile.Close()

	var defFile *os.File

	if *writeToFile {
		defFile, err = os.Create("def.txt")
		if err != nil {
			log.Fatal("Error opening file: ", err)
		}

		fmt.Printf("Opened file %s\n", "def.txt")

		defer defFile.Close()
	}

	pgConn, err := pgx.Connect(context.Background(), *pgString)
	if err != nil {
		log.Fatal("Error connecting to Postgres: ", err)
	}

	fmt.Println("Connected to Postgres")
	defer pgConn.Close(context.Background())

	if *pgCreateTable {
		createTables(pgConn)
		fmt.Println("Created 'entries' and 'senses' tables")
	}

	start := time.Now()

	decoder := xml.NewDecoder(tezFile)

	entryCount := 0

	batch := pgx.Batch{}

	dbWriteDone := make(chan struct{}, 1)
	dbWriteDone <- struct{}{}

	for {
		t, _ := decoder.Token()

		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "extent":
				var extent Extent
				decoder.DecodeElement(&extent, &se)
				for _, m := range extent.Measure {
					switch m.Unit {
					case "entry":
						entryQty, _ = strconv.Atoi(m.Quantity)
					case "sense":
						senseQty, _ = strconv.Atoi(m.Quantity)
					case "lexeme":
						lexemeQty, _ = strconv.Atoi(m.Quantity)
					}
				}

			case "entry":
				var currentEntry Entry
				decoder.DecodeElement(&currentEntry, &se)
				entryCount++

				if *writeToFile {
					writeDefinitions(defFile, currentEntry.Id, currentEntry.SortKey, currentEntry.Sense)
				}

				if *pgString != "" {
					addEntryToBatch(&batch, currentEntry)
				}

				if entryCount%50000 == 0 {
					fmt.Printf("Parsed %d entries of %d (%d%%)\n", entryCount, entryQty, entryCount*100/entryQty)

					// wait for previous batch to finish writing to db
					<-dbWriteDone

					// send the current batch in goroutine and continue parsing
					go func(b pgx.Batch) {
						res := pgConn.SendBatch(context.Background(), &b)
						err := res.Close()
						if err != nil {
							log.Fatal("Error sending batch: ", err)
						}
						dbWriteDone <- struct{}{}
					}(batch)

					batch = pgx.Batch{}
				}
			}
		}
	}

	// send the last batch
	<-dbWriteDone
	res := pgConn.SendBatch(context.Background(), &batch)
	err = res.Close()
	if err != nil {
		log.Fatal("Error sending batch: ", err)
	}

	fmt.Printf("Finished parsing %d entries\n", entryCount)

	if *pgString != "" {
		fmt.Println("Wrote all entries to DB")
	}

	if *writeToFile {
		fmt.Printf("Wrote %d definitions to %s\n", defCount, "def.txt")
	}

	fmt.Printf("Total time: %.2fs\n", time.Since(start).Seconds())
}
