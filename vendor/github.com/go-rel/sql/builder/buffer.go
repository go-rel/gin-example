package builder

import (
	"database/sql/driver"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-rel/sql"
)

// UnescapeCharacter disable field escaping when it starts with this character.
var UnescapeCharacter byte = '^'

var escapeCache sync.Map

type escapeCacheKey struct {
	table  string
	value  string
	quoter Quoter
}

// Buffer is used to build query string.
type Buffer struct {
	strings.Builder
	Quoter              Quoter
	ValueConverter      driver.ValueConverter
	ArgumentPlaceholder string
	AllowTableSchema    bool
	ArgumentOrdinal     bool
	InlineValues        bool
	BoolTrueValue       string
	BoolFalseValue      string
	valueCount          int
	arguments           []any
}

// WriteValue query placeholder and append value to argument.
func (b *Buffer) WriteValue(value any) {
	if !b.InlineValues {
		b.WritePlaceholder()
		b.arguments = append(b.arguments, value)
		return
	}

	// Detect float bits to not lose precision after converting to float64
	floatBits := 64
	if value != nil && reflect.TypeOf(value).Kind() == reflect.Float32 {
		floatBits = 32
	}

	if v, err := b.ValueConverter.ConvertValue(value); err != nil {
		log.Printf("[WARN] unsupported inline value %v: %v", value, err)
	} else {
		value = v
	}

	if value == nil {
		b.WriteString("NULL")
		return
	}

	switch v := value.(type) {
	case string:
		b.WriteString(b.Quoter.Value(v))
		return
	case []byte:
		b.WriteString(b.Quoter.Value(string(v)))
		return
	case time.Time:
		b.WriteString(b.Quoter.Value(v.Format(sql.DefaultTimeLayout)))
		return
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.WriteString(strconv.FormatInt(rv.Int(), 10))
		return
	case reflect.Float32, reflect.Float64:
		b.WriteString(strconv.FormatFloat(rv.Float(), 'g', -1, floatBits))
		return
	case reflect.Bool:
		if rv.Bool() {
			b.WriteString(b.BoolTrueValue)
		} else {
			b.WriteString(b.BoolFalseValue)
		}
		return
	}
	b.WriteString(fmt.Sprintf("%v", value))
}

// WritePlaceholder without adding argument.
// argument can be added later using AddArguments function.
func (b *Buffer) WritePlaceholder() {
	b.valueCount++
	b.WriteString(b.ArgumentPlaceholder)
	if b.ArgumentOrdinal {
		b.WriteString(strconv.Itoa(b.valueCount))
	}
}

// WriteField writes table and field name.
func (b *Buffer) WriteField(table, field string) {
	b.WriteString(b.escape(table, field))
}

// WriteTable writes table name.
func (b *Buffer) WriteTable(table string) {
	b.WriteString(b.escape(table, ""))
}

// WriteEscape string.
func (b *Buffer) WriteEscape(value string) {
	b.WriteString(b.escape("", value))
}

func (b Buffer) escape(table, value string) string {
	if table == "" && value == "*" {
		return value
	}

	key := escapeCacheKey{table: table, value: value, quoter: b.Quoter}
	escapedValue, ok := escapeCache.Load(key)
	if ok {
		return escapedValue.(string)
	}

	table, alias := extractAlias(table)
	var escapedTable string
	if table != "" {
		if table != alias {
			if value == "" {
				return b.escape(table, "") + " AS " + b.Quoter.ID(alias)
			} else {
				escapedTable = b.Quoter.ID(alias)
			}
		}
		if b.AllowTableSchema && strings.IndexByte(table, '.') >= 0 {
			parts := strings.Split(table, ".")
			for i, part := range parts {
				part = strings.TrimSpace(part)
				parts[i] = b.Quoter.ID(part)
			}
			escapedTable = strings.Join(parts, ".")
		} else {
			escapedTable = b.Quoter.ID(strings.ReplaceAll(table, ".", "_"))
		}
	}

	if value == "" {
		escapedValue = escapedTable
	} else if value == "*" {
		escapedValue = escapedTable + ".*"
	} else if len(value) > 0 && value[0] == UnescapeCharacter {
		escapedValue = value[1:]
	} else if _, err := strconv.Atoi(value); err == nil {
		escapedValue = value
	} else if i := strings.Index(strings.ToLower(value), " as "); i > -1 {
		escapedValue = b.escape(alias, value[:i]) + " AS " + b.Quoter.ID(value[i+4:])
	} else if start, end := strings.IndexRune(value, '('), strings.IndexRune(value, ')'); start >= 0 && end >= 0 && end > start {
		escapedValue = value[:start+1] + b.escape(alias, value[start+1:end]) + value[end:]
	} else {
		parts := strings.Split(value, ".")
		for i, part := range parts {
			part = strings.TrimSpace(part)
			if part == "*" && i == len(parts)-1 {
				break
			}
			parts[i] = b.Quoter.ID(part)
		}
		result := strings.Join(parts, ".")
		if len(parts) == 1 && table != "" {
			result = escapedTable + "." + result
		}
		escapedValue = result
	}

	escapeCache.Store(key, escapedValue)
	return escapedValue.(string)
}

// AddArguments appends multiple arguments without writing placeholder query..
func (b *Buffer) AddArguments(args ...any) {
	if b.arguments == nil {
		b.arguments = args
	} else {
		b.arguments = append(b.arguments, args...)
	}
}

func (b Buffer) Arguments() []any {
	return b.arguments
}

// Reset buffer.
func (b *Buffer) Reset() {
	b.Builder.Reset()
	b.valueCount = 0
	b.arguments = nil
}

// BufferFactory is used to create buffer based on shared settings.
type BufferFactory struct {
	Quoter              Quoter
	ValueConverter      driver.ValueConverter
	AllowTableSchema    bool
	ArgumentPlaceholder string
	ArgumentOrdinal     bool
	InlineValues        bool
	BoolTrueValue       string
	BoolFalseValue      string
}

func (bf BufferFactory) Create() Buffer {
	conv := bf.ValueConverter
	if conv == nil {
		conv = driver.DefaultParameterConverter
	}
	return Buffer{
		Quoter:              bf.Quoter,
		ValueConverter:      conv,
		AllowTableSchema:    bf.AllowTableSchema,
		ArgumentPlaceholder: bf.ArgumentPlaceholder,
		ArgumentOrdinal:     bf.ArgumentOrdinal,
		InlineValues:        bf.InlineValues,
		BoolTrueValue:       bf.BoolTrueValue,
		BoolFalseValue:      bf.BoolFalseValue,
	}
}

// extract alias in the form of table as alias
// if no alias, table will be returned as alias
func extractAlias(input string) (string, string) {
	if i := strings.Index(strings.ToLower(input), " as "); i > -1 {
		return input[:i], input[i+4:]
	}

	return input, input
}
