package fixlenrw

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"testing"
)

type Address struct {
	City   string
	Street string
	Number int
	Main   bool
}

func TestWriter_Write(t *testing.T) {
	type fields struct {
		record   *Record
		buffer   *bufio.Writer
		numCols  int
		lineSize int
		padChar  rune
		userCrlf bool
	}
	tests := []struct {
		name    string
		fields  fields
		input   []Address
		want    string
		wantErr bool
	}{
		{
			name: "Regular",
			input: []Address{
				{City: "city1", Street: "str1", Number: 5, Main: true},
				{City: "city2", Street: "str2", Number: 4537, Main: true},
				{City: "city3", Street: "str3", Number: 24, Main: false},
				{City: "city4", Street: "str4", Number: 125, Main: true},
				{City: "city5", Street: "str5", Number: 2, Main: true},
				{City: "city6", Street: "", Number: 5, Main: true},
			},
			want: `city1str15   t
city2str24537t
city3str324  f
city4str4125 t
city5str52   t
city6    5   t
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputWriter := bytes.NewBufferString("")
			schemaFile, err := os.Open("../schema.json")
			if err != nil {
				t.Fatal("error reading schema file")
			}
			defer schemaFile.Close()
			writer := NewWriter(outputWriter, schemaFile)
			for _, record := range tt.input {
				fmt.Println("writing ", record)
				if err := writer.Write(&record); (err != nil) != tt.wantErr {
					t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			writer.Flush()
			if gotOutputWriter := outputWriter.String(); gotOutputWriter != tt.want {
				t.Errorf("NewWriter() gotOutputWriter = %v, want %v", gotOutputWriter, tt.want)
			}
		})
	}
}
