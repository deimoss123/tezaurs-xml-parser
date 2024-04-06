package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"time"
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

const delimiter = "    "

var defCount = 0

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

func main() {
	filename := flag.String("f", "", "Tezaurs file name")
	_ = flag.String("c", "", "Postgres connection string")

	flag.Parse()

	file, err := os.Open(*filename)

	if err != nil {
		fmt.Println("Error opening file: ", err)
		return
	}
	fmt.Printf("Opened file %s\n", *filename)

	defFile, err := os.Create("def.txt")
	if err != nil {
		fmt.Println("Error opening file: ", err)
		return
	}

	fmt.Printf("Opened file %s\n", "def.txt")

	start := time.Now()

	defer file.Close()
	defer defFile.Close()

	decoder := xml.NewDecoder(file)

	entryCount := 0

	for {
		t, _ := decoder.Token()

		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "entry":
				var currentEntry Entry
				decoder.DecodeElement(&currentEntry, &se)
				entryCount++

				writeDefinitions(defFile, currentEntry.Id, currentEntry.SortKey, currentEntry.Sense)

				if entryCount%50000 == 0 {
					fmt.Printf("Parsed %d entries\n", entryCount)
				}

				// if currentEntry.SortKey == "braukt" {
				// if currentEntry.SortKey == "sieviete" {
				// 	jsonEntry, _ := json.MarshalIndent(currentEntry, "", "  ")
				// 	fmt.Println(string(jsonEntry))
				// 	fmt.Printf("%+v\n", currentEntry)
				// }
			}
		}
	}

	// fmt.Println(currentEntry)
	// fmt.Printf("%+v\n", currentEntry)
	// fmt.Printf("%+v\n", senseTypes)

	fmt.Printf("Finished parsing %d entries\n", entryCount)
	fmt.Printf("Wrote %d definitions to %s\n", defCount, "def.txt")
	fmt.Printf("Total time: %.2fs\n", time.Since(start).Seconds())
}
