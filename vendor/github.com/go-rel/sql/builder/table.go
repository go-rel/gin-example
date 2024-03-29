package builder

import (
	"log"
	"strconv"

	"github.com/go-rel/rel"
)

type (
	ColumnMapper        func(*rel.Column) (string, int, int)
	ColumnOptionsMapper func(*rel.Column) string
	DropKeyMapper       func(rel.KeyType) string
	DefinitionFilter    func(table rel.Table, def rel.TableDefinition) bool
)

// Table builder.
type Table struct {
	BufferFactory       BufferFactory
	ColumnMapper        ColumnMapper
	ColumnOptionsMapper ColumnOptionsMapper
	DropKeyMapper       DropKeyMapper
	DefinitionFilter    DefinitionFilter
}

// Build SQL query for table creation and modification.
func (t Table) Build(table rel.Table) string {
	buffer := t.BufferFactory.Create()

	switch table.Op {
	case rel.SchemaCreate:
		t.WriteCreateTable(&buffer, table)
	case rel.SchemaAlter:
		t.WriteAlterTable(&buffer, table)
	case rel.SchemaRename:
		t.WriteRenameTable(&buffer, table)
	case rel.SchemaDrop:
		t.WriteDropTable(&buffer, table)
	}

	return buffer.String()
}

// WriteCreateTable query to buffer.
func (t Table) WriteCreateTable(buffer *Buffer, table rel.Table) {
	defs := t.definitions(table)

	buffer.WriteString("CREATE TABLE ")

	if table.Optional {
		buffer.WriteString("IF NOT EXISTS ")
	}

	buffer.WriteTable(table.Name)
	if len(defs) > 0 {
		buffer.WriteString(" (")

		for i, def := range defs {
			if i > 0 {
				buffer.WriteString(", ")
			}
			switch v := def.(type) {
			case rel.Column:
				t.WriteColumn(buffer, v)
			case rel.Key:
				t.WriteKey(buffer, v)
			case rel.Raw:
				buffer.WriteString(string(v))
			}
		}

		buffer.WriteByte(')')
	}
	t.WriteOptions(buffer, table.Options)
	buffer.WriteByte(';')
}

// WriteAlterTable query to buffer.
func (t Table) WriteAlterTable(buffer *Buffer, table rel.Table) {
	defs := t.definitions(table)

	for _, def := range defs {
		buffer.WriteString("ALTER TABLE ")
		buffer.WriteTable(table.Name)
		buffer.WriteByte(' ')

		switch v := def.(type) {
		case rel.Column:
			switch v.Op {
			case rel.SchemaCreate:
				buffer.WriteString("ADD COLUMN ")
				t.WriteColumn(buffer, v)
			case rel.SchemaRename:
				// Add Change
				buffer.WriteString("RENAME COLUMN ")
				buffer.WriteEscape(v.Name)
				buffer.WriteString(" TO ")
				buffer.WriteEscape(v.Rename)
			case rel.SchemaDrop:
				buffer.WriteString("DROP COLUMN ")
				buffer.WriteEscape(v.Name)
			}
		case rel.Key:
			// TODO: Rename and Drop, PR welcomed.
			switch v.Op {
			case rel.SchemaCreate:
				buffer.WriteString("ADD ")
				t.WriteKey(buffer, v)
			case rel.SchemaDrop:
				buffer.WriteString("DROP ")
				buffer.WriteString(t.DropKeyMapper(v.Type))
				buffer.WriteString(" ")
				buffer.WriteEscape(v.Name)
			}
		}

		t.WriteOptions(buffer, table.Options)
		buffer.WriteByte(';')
	}
}

// WriteRenameTable query to buffer.
func (t Table) WriteRenameTable(buffer *Buffer, table rel.Table) {
	buffer.WriteString("ALTER TABLE ")
	buffer.WriteTable(table.Name)
	buffer.WriteString(" RENAME TO ")
	buffer.WriteTable(table.Rename)
	buffer.WriteByte(';')
}

// WriteDropTable query to buffer.
func (t Table) WriteDropTable(buffer *Buffer, table rel.Table) {
	buffer.WriteString("DROP TABLE ")

	if table.Optional {
		buffer.WriteString("IF EXISTS ")
	}

	buffer.WriteTable(table.Name)
	buffer.WriteByte(';')
}

// WriteColumn definition to buffer.
func (t Table) WriteColumn(buffer *Buffer, column rel.Column) {
	typ, m, n := t.ColumnMapper(&column)

	buffer.WriteEscape(column.Name)
	buffer.WriteByte(' ')
	buffer.WriteString(typ)

	if m != 0 {
		buffer.WriteByte('(')
		buffer.WriteString(strconv.Itoa(m))

		if n != 0 {
			buffer.WriteByte(',')
			buffer.WriteString(strconv.Itoa(n))
		}

		buffer.WriteByte(')')
	}

	if opts := t.ColumnOptionsMapper(&column); opts != "" {
		buffer.WriteByte(' ')
		buffer.WriteString(opts)
	}

	if column.Default != nil {
		buffer.WriteString(" DEFAULT ")
		buffer.WriteValue(column.Default)
	}

	t.WriteOptions(buffer, column.Options)
}

// WriteKey definition to buffer.
func (t Table) WriteKey(buffer *Buffer, key rel.Key) {
	typ := string(key.Type)

	buffer.WriteString(typ)

	if key.Name != "" {
		buffer.WriteByte(' ')
		buffer.WriteEscape(key.Name)
	}

	buffer.WriteString(" (")
	for i, col := range key.Columns {
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteEscape(col)
	}
	buffer.WriteString(")")

	if key.Type == rel.ForeignKey {
		buffer.WriteString(" REFERENCES ")
		buffer.WriteTable(key.Reference.Table)

		buffer.WriteString(" (")
		for i, col := range key.Reference.Columns {
			if i > 0 {
				buffer.WriteString(", ")
			}
			buffer.WriteEscape(col)
		}
		buffer.WriteString(")")

		if onDelete := key.Reference.OnDelete; onDelete != "" {
			buffer.WriteString(" ON DELETE ")
			buffer.WriteString(onDelete)
		}

		if onUpdate := key.Reference.OnUpdate; onUpdate != "" {
			buffer.WriteString(" ON UPDATE ")
			buffer.WriteString(onUpdate)
		}
	}

	t.WriteOptions(buffer, key.Options)
}

// WriteOptions sql to buffer.
func (t Table) WriteOptions(buffer *Buffer, options string) {
	if options == "" {
		return
	}

	buffer.WriteByte(' ')
	buffer.WriteString(options)
}

func (t Table) definitions(table rel.Table) []rel.TableDefinition {
	if t.DefinitionFilter == nil {
		return table.Definitions
	}

	result := []rel.TableDefinition{}

	for _, def := range table.Definitions {
		if t.DefinitionFilter(table, def) {
			result = append(result, def)
		} else {
			log.Printf("[REL] An unsupported table definition has been excluded: %T", def)
		}
	}

	return result
}
