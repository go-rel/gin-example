package mysql

import (
	"database/sql/driver"
	"strings"
	"time"
)

// Quote MySQL identifiers and literals.
type Quote struct{}

func (q Quote) ID(name string) string {
	end := strings.IndexRune(name, 0)
	if end > -1 {
		name = name[:end]
	}
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func (q Quote) Value(v interface{}) string {
	switch v := v.(type) {
	default:
		panic("unsupported value")
	case string:
		// TODO: Need to check on connection for NO_BACKSLASH_ESCAPES
		rv := []rune(v)
		buf := make([]rune, len(rv)*2)
		pos := 0
		for i := 0; i < len(rv); i++ {
			c := rv[i]
			switch c {
			case '\x00':
				buf[pos] = '\\'
				buf[pos+1] = '0'
				pos += 2
			case '\n':
				buf[pos] = '\\'
				buf[pos+1] = 'n'
				pos += 2
			case '\r':
				buf[pos] = '\\'
				buf[pos+1] = 'r'
				pos += 2
			case '\x1a':
				buf[pos] = '\\'
				buf[pos+1] = 'Z'
				pos += 2
			case '\'':
				buf[pos] = '\\'
				buf[pos+1] = '\''
				pos += 2
			case '"':
				buf[pos] = '\\'
				buf[pos+1] = '"'
				pos += 2
			case '\\':
				buf[pos] = '\\'
				buf[pos+1] = '\\'
				pos += 2
			default:
				buf[pos] = c
				pos++
			}
		}

		return "'" + string(buf[:pos]) + "'"
	}
}

// ValueConvert converts values to MySQL literals.
type ValueConvert struct{}

func (c ValueConvert) ConvertValue(v interface{}) (driver.Value, error) {
	v, err := driver.DefaultParameterConverter.ConvertValue(v)
	if err != nil {
		return nil, err
	}
	switch v := v.(type) {
	default:
		return v, nil
	case time.Time:
		return v.Truncate(time.Microsecond).Format("2006-01-02 15:04:05.999999"), nil
	}
}
