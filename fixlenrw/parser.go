package fixlenrw

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type Parser struct {
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

type ParsingError struct {
	line    int
	colName string
	data    string
	message string
	err     error
}

func (e *ParsingError) Error() string {
	return fmt.Sprintf("error parsing line %d: %v", e.line, e.err)
}

var (
	ErrLineTooLong  = errors.New("line is too long")
	ErrLineTooShort = errors.New("line is too short")
	ErrNumCols      = errors.New("line has too few columns")
)

func NewParser(inputReader, schemaReader io.Reader) *Parser {
	scanner := bufio.NewScanner(inputReader)
	record := readSchema(schemaReader)
	lineSize := 0
	for _, col := range record.Columns {
		lineSize += col.Length
	}
	return &Parser{scanner: scanner, record: record,
		numCols: len(record.Columns), lineSize: lineSize, Errors: make([]ParsingError, 0)}
}

func (parser *Parser) WithTrimLeadingSpace() *Parser {
	parser.trimLeadingSpace = true
	return parser
}
func (parser *Parser) WithTrimTrailingSpace() *Parser {
	parser.trimTrailingSpace = true
	return parser
}
func (parser *Parser) WithIgnoreEmptyCols() *Parser {
	parser.ignoreEmptyCols = true
	return parser
}

func parseLine(line string, record *Record, parser *Parser) ([]string, error) {
	var err error
	tokens := make([]string, 0, parser.numCols)
	var v []rune
	parser.numLines++
	if len(line) > parser.lineSize {
		err := ParsingError{line: parser.numLines, err: ErrLineTooLong}
		parser.Errors = append(parser.Errors, err)
		if parser.ignoreExtraSize {
			v = []rune(line)
			line = string(v[:parser.lineSize])
		} else {
			parser.done = true
			return nil, &err
		}
	} else if len(line) < parser.lineSize {
		err := ParsingError{line: parser.numLines, err: ErrLineTooShort}
		parser.Errors = append(parser.Errors, err)
		if parser.handleShortLines {
			line = line + strings.Repeat(" ", parser.lineSize-len(line))
			v = []rune(line)
		} else {
			parser.done = true
			return nil, &err
		}
	} else {
		v = []rune(line)
	}
	idx := 0
	for _, column := range record.Columns {
		token := string(v[idx : idx+column.Length])
		if parser.trimLeadingSpace {
			token = strings.TrimLeftFunc(token, unicode.IsSpace)
		}
		if parser.trimTrailingSpace {
			token = strings.TrimRightFunc(token, unicode.IsSpace)
		}
		if len(token) == 0 {
			err := ParsingError{line: parser.numLines, colName: column.Name, err: ErrNumCols}
			parser.Errors = append(parser.Errors, err)
			if !parser.ignoreEmptyCols {
				parser.done = true
				return nil, &err
			}
		}
		tokens = append(tokens, token)
		idx += column.Length
	}
	return tokens, err
}

func (parser *Parser) All() (records [][]string, err error) {
	for parser.HasNext() {
		tokens, err := parser.Next()
		if err != nil {
			return nil, err
		}
		records = append(records, tokens)
	}
	parser.done = true
	return records, nil
}

func (parser *Parser) Next() ([]string, error) {
	line := parser.scanner.Text()
	return parseLine(line, parser.record, parser)
}

func (parser *Parser) UnmarshallNext(v interface{}) error {
	tokens, err := parser.Next()
	if err != nil {
		return err
	}
	s := reflect.ValueOf(v).Elem()
	fieldToIdx := make(map[string]int, s.NumField())
	for i := 0; i < s.NumField(); i++ {
		fieldToIdx[s.Type().Field(i).Name] = i
	}
	for idx, col := range parser.record.Columns {
		value := tokens[idx]
		fieldIdx, contains := fieldToIdx[col.Name]
		if !contains {
			continue
		}
		switch s.Field(fieldIdx).Type().Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			s.Field(fieldIdx).SetInt(intValue)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintValue, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			s.Field(fieldIdx).SetUint(uintValue)
		case reflect.Float32, reflect.Float64:
			floatValue, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			s.Field(fieldIdx).SetFloat(floatValue)
		case reflect.String:
			s.Field(fieldIdx).SetString(value)
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			s.Field(fieldIdx).SetBool(boolValue)
		default:
			log.Fatal("unsupported type", s.Field(idx).Type().Kind())
		}
	}
	return nil
}

func (parser *Parser) HasNext() bool {
	if parser.done {
		return false
	}
	scan := parser.scanner.Scan()
	if !scan {
		parser.done = true
	}
	return scan
}

type RowsIterator interface {
	HasNext() bool
	Next() ([]string, error)
	UnmarshallNext(v interface{}) error
}
