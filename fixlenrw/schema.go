package fixlenrw

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
)

type Column struct {
	Name   string
	Length int
}

type Record struct {
	Columns []Column
}

func readSchema(schemaReader io.Reader) *Record {
	content, err := ioutil.ReadAll(schemaReader)
	if err != nil {
		log.Fatal(err)
	}
	var record Record
	err = json.Unmarshal(content, &record)
	if err != nil {
		log.Fatal(err)
	}
	return &record
}
