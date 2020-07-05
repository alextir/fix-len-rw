package fixlenrw

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestParser_All(t *testing.T) {
	type fields struct {
		scanner           *bufio.Scanner
		record            *Record
		handleShortLines  bool
		ignoreExtraSize   bool
		ignoreEmptyCols   bool
		trimLeadingSpace  bool
		trimTrailingSpace bool
		numLines          int
		numCols           int
		lineSize          int
		Errors            []ParsingError
		done              bool
	}
	tests := []struct {
		name        string
		fields      fields
		input       string
		wantRecords [][]string
		wantErr     bool
		err         error
	}{
		{
			name: "Regular",
			fields: fields{
				Errors:            []ParsingError{{line: 6, colName: "Street", err: ErrNumCols}},
				handleShortLines:  false,
				ignoreExtraSize:   false,
				ignoreEmptyCols:   true,
				trimLeadingSpace:  true,
				trimTrailingSpace: true,
			},
			input: `city1str15   T
city2str24537T
city3str3 24 F
city4str4125 T
city5str52   T
city6    5   T`,
			wantRecords: [][]string{
				{"city1", "str1", "5", "T"},
				{"city2", "str2", "4537", "T"},
				{"city3", "str3", "24", "F"},
				{"city4", "str4", "125", "T"},
				{"city5", "str5", "2", "T"},
				{"city6", "", "5", "T"},
			},
		}, {
			name: "Ignore extra size",
			fields: fields{
				Errors:            []ParsingError{{line: 1, err: ErrLineTooLong}},
				handleShortLines:  false,
				ignoreExtraSize:   true,
				ignoreEmptyCols:   true,
				trimLeadingSpace:  true,
				trimTrailingSpace: true,
			},
			input: `city1str15   Tabc`,
			wantRecords: [][]string{
				{"city1", "str1", "5", "T"},
			},
		}, {
			name: "Don't ignore extra size",
			fields: fields{
				Errors:            []ParsingError{{line: 1, err: ErrLineTooLong}},
				handleShortLines:  false,
				ignoreExtraSize:   false,
				ignoreEmptyCols:   true,
				trimLeadingSpace:  true,
				trimTrailingSpace: true,
			},
			input:       `city1str15   Tabc`,
			wantRecords: nil,
			wantErr:     true,
			err:         &ParsingError{line: 1, err: ErrLineTooLong},
		}, {
			name: "Don't ignore empty cols",
			fields: fields{
				Errors:            []ParsingError{{line: 1, colName: "Street", err: ErrNumCols}},
				handleShortLines:  false,
				ignoreExtraSize:   false,
				ignoreEmptyCols:   false,
				trimLeadingSpace:  true,
				trimTrailingSpace: true,
			},
			input:       `city6    5   T`,
			wantRecords: nil,
			wantErr:     true,
			err:         &ParsingError{line: 1, colName: "Street", err: ErrNumCols},
		}, {
			name: "Handle short lines",
			fields: fields{
				Errors: []ParsingError{
					{line: 1, err: ErrLineTooShort},
					{line: 1, colName: "Main", err: ErrNumCols},
				},
				handleShortLines:  true,
				ignoreExtraSize:   false,
				ignoreEmptyCols:   true,
				trimLeadingSpace:  true,
				trimTrailingSpace: true,
			},
			input: `city1str15   `,
			wantRecords: [][]string{
				{"city1", "str1", "5", ""},
			},
		}, {
			name: "Don't handle short lines",
			fields: fields{
				Errors: []ParsingError{
					{line: 1, err: ErrLineTooShort},
				},
				handleShortLines:  false,
				ignoreExtraSize:   false,
				ignoreEmptyCols:   true,
				trimLeadingSpace:  true,
				trimTrailingSpace: true,
			},
			input:       `city1str15   `,
			wantRecords: nil,
			wantErr:     true,
			err:         &ParsingError{line: 1, err: ErrLineTooShort},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaFile, err := os.Open("../schema.json")
			if err != nil {
				t.Fatal("error reading schema file")
			}
			defer schemaFile.Close()
			parser := NewParser(strings.NewReader(tt.input), schemaFile)
			parser.handleShortLines = tt.fields.handleShortLines
			parser.ignoreExtraSize = tt.fields.ignoreExtraSize
			parser.ignoreEmptyCols = tt.fields.ignoreEmptyCols
			parser.trimLeadingSpace = tt.fields.trimLeadingSpace
			parser.trimTrailingSpace = tt.fields.trimTrailingSpace

			gotRecords, err := parser.All()
			if (err != nil) != tt.wantErr {
				t.Errorf("All() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("All() err = %v, wantErr %v", err, tt.err)
			}
			if !reflect.DeepEqual(gotRecords, tt.wantRecords) {
				t.Errorf("All() gotRecords = %v, want %v", gotRecords, tt.wantRecords)
			}
			if !reflect.DeepEqual(parser.Errors, tt.fields.Errors) {
				t.Errorf("All() parser.Errors = %v, want %v", parser.Errors, tt.fields.Errors)
			}
			if !parser.done {
				t.Errorf("All() done = %v, expected done %v", parser.done, true)
			}
		})
	}
}

func TestParser_UnmarshallNext(t *testing.T) {
	type fields struct {
		scanner           *bufio.Scanner
		record            *Record
		handleShortLines  bool
		ignoreExtraSize   bool
		ignoreEmptyCols   bool
		trimLeadingSpace  bool
		trimTrailingSpace bool
		numLines          int
		numCols           int
		lineSize          int
		Errors            []ParsingError
		done              bool
	}
	type Address struct {
		City   string
		Street string
		Number int
		Main   bool
	}
	tests := []struct {
		name        string
		fields      fields
		input       string
		wantRecords []Address
		wantErr     bool
	}{
		{
			name: "Regular",
			fields: fields{
				Errors:            []ParsingError{{line: 6, colName: "Street", err: ErrNumCols}},
				handleShortLines:  false,
				ignoreExtraSize:   false,
				ignoreEmptyCols:   true,
				trimLeadingSpace:  true,
				trimTrailingSpace: true,
			},
			input: `city1str15   T
city2str24537T
city3str3 24 F
city4str4125 T
city5str52   T
city6    5   T`,
			wantRecords: []Address{
				{City: "city1", Street: "str1", Number: 5, Main: true},
				{City: "city2", Street: "str2", Number: 4537, Main: true},
				{City: "city3", Street: "str3", Number: 24, Main: false},
				{City: "city4", Street: "str4", Number: 125, Main: true},
				{City: "city5", Street: "str5", Number: 2, Main: true},
				{City: "city6", Street: "", Number: 5, Main: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaFile, err := os.Open("../schema.json")
			if err != nil {
				t.Fatal("error reading schema file")
			}
			defer schemaFile.Close()
			parser := NewParser(strings.NewReader(tt.input), schemaFile)
			parser.handleShortLines = tt.fields.handleShortLines
			parser.ignoreExtraSize = tt.fields.ignoreExtraSize
			parser.ignoreEmptyCols = tt.fields.ignoreEmptyCols
			parser.trimLeadingSpace = tt.fields.trimLeadingSpace
			parser.trimTrailingSpace = tt.fields.trimTrailingSpace

			for i := 0; parser.HasNext(); i++ {
				var addr Address
				if err := parser.UnmarshallNext(&addr); (err != nil) != tt.wantErr {
					t.Errorf("UnmarshallNext() error = %v, wantErr %v", err, tt.wantErr)
				}
				if !reflect.DeepEqual(addr, tt.wantRecords[i]) {
					t.Errorf("UnmarshallNext() record = %v, wantRecord %v", addr, tt.wantRecords[i])
				}
				fmt.Println("unmarshalled:", addr)
			}

			if !parser.done {
				t.Errorf("UnmarshallNext() done = %v, expected done %v", parser.done, true)
			}
		})
	}
}

func TestParser_UnmarshallNextMissingCol(t *testing.T) {
	type fields struct {
		scanner           *bufio.Scanner
		record            *Record
		handleShortLines  bool
		ignoreExtraSize   bool
		ignoreEmptyCols   bool
		trimLeadingSpace  bool
		trimTrailingSpace bool
		numLines          int
		numCols           int
		lineSize          int
		Errors            []ParsingError
		done              bool
	}
	type Address struct {
		City   string
		Street string
	}
	tests := []struct {
		name        string
		fields      fields
		input       string
		wantRecords []Address
		wantErr     bool
	}{
		{
			name: "Unmarshall with missing col",
			fields: fields{
				Errors:            []ParsingError{{line: 6, colName: "Street", err: ErrNumCols}},
				handleShortLines:  false,
				ignoreExtraSize:   false,
				ignoreEmptyCols:   true,
				trimLeadingSpace:  true,
				trimTrailingSpace: true,
			},
			input: `city1str15   T
city2str24537T
city3str3 24 F
city4str4125 T
city5str52   T
city6    5   T`,
			wantRecords: []Address{
				{City: "city1", Street: "str1"},
				{City: "city2", Street: "str2"},
				{City: "city3", Street: "str3"},
				{City: "city4", Street: "str4"},
				{City: "city5", Street: "str5"},
				{City: "city6", Street: ""},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaFile, err := os.Open("../schema.json")
			if err != nil {
				t.Fatal("error reading schema file")
			}
			defer schemaFile.Close()
			parser := NewParser(strings.NewReader(tt.input), schemaFile)
			parser.handleShortLines = tt.fields.handleShortLines
			parser.ignoreExtraSize = tt.fields.ignoreExtraSize
			parser.ignoreEmptyCols = tt.fields.ignoreEmptyCols
			parser.trimLeadingSpace = tt.fields.trimLeadingSpace
			parser.trimTrailingSpace = tt.fields.trimTrailingSpace

			for i := 0; parser.HasNext(); i++ {
				var addr Address
				if err := parser.UnmarshallNext(&addr); (err != nil) != tt.wantErr {
					t.Errorf("UnmarshallNext() error = %v, wantErr %v", err, tt.wantErr)
				}
				if !reflect.DeepEqual(addr, tt.wantRecords[i]) {
					t.Errorf("UnmarshallNext() record = %v, wantRecord %v", addr, tt.wantRecords[i])
				}
				fmt.Println("unmarshalled:", addr)
			}

			if !parser.done {
				t.Errorf("UnmarshallNext() done = %v, expected done %v", parser.done, true)
			}
		})
	}
}
