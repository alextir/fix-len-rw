package fixlenrw

import (
	"bufio"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type Writer struct {
	record   *Record
	buffer   *bufio.Writer
	numCols  int
	lineSize int
	padChar  rune
	userCrlf bool
}

func NewWriter(outputWriter io.Writer, schemaReader io.Reader) *Writer {
	record := readSchema(schemaReader)
	lineSize := 0
	for _, col := range record.Columns {
		lineSize += col.Length
	}
	return &Writer{record: record, numCols: len(record.Columns), lineSize: lineSize, buffer: bufio.NewWriter(outputWriter), padChar: ' '}
}

func (writer *Writer) Write(v interface{}) error {
	s := reflect.ValueOf(v).Elem()
	fieldToIdx := make(map[string]int, s.NumField())
	for i := 0; i < s.NumField(); i++ {
		fieldToIdx[s.Type().Field(i).Name] = i
	}
	for _, col := range writer.record.Columns {
		fieldIdx := fieldToIdx[col.Name]
		var sval string
		switch v := s.Field(fieldIdx); v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			sval = strconv.FormatInt(v.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			sval = strconv.FormatUint(v.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			sval = strconv.FormatFloat(v.Float(), 'E', -1, 64)
		case reflect.String:
			sval = v.String()
		case reflect.Bool:
			sval = strconv.FormatBool(v.Bool())
		default:
			log.Fatal("unsupported type", v.Kind())
		}
		if err := appendToken(sval, writer.padChar, writer.buffer, &col); err != nil {
			return err
		}
	}
	var err error
	if writer.userCrlf {
		_, err = writer.buffer.WriteString("\r\n")
	} else {
		err = writer.buffer.WriteByte('\n')
	}
	return err
}

func (writer *Writer) WriteAll(v []interface{}) error {
	for i := range v {
		if err := writer.Write(i); err != nil {
			return err
		}
	}
	return writer.buffer.Flush()
}

func (writer *Writer) Flush() error {
	return writer.buffer.Flush()
}

func format(s string, pad rune, col *Column) string {
	min := min(len(s), col.Length)
	s = s[:min]
	if min < col.Length {
		s = s + strings.Repeat(string(pad), col.Length-min)
	}
	return s
}

func appendToken(s string, pad rune, buffer *bufio.Writer, col *Column) error {
	token := format(s, pad, col)
	_, err := buffer.WriteString(token)
	return err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
