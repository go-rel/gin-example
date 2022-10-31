package sql

import (
	"github.com/go-rel/rel"
)

type QueryBuilder interface {
	Build(query rel.Query) (string, []any)
}

type InsertBuilder interface {
	Build(table string, primaryField string, mutates map[string]rel.Mutate, onConflict rel.OnConflict) (string, []any)
}

type InsertAllBuilder interface {
	Build(table string, primaryField string, fields []string, bulkMutates []map[string]rel.Mutate, onConflict rel.OnConflict) (string, []any)
}

type UpdateBuilder interface {
	Build(table string, primaryField string, mutates map[string]rel.Mutate, filter rel.FilterQuery) (string, []any)
}

type DeleteBuilder interface {
	Build(table string, filter rel.FilterQuery) (string, []any)
}

type TableBuilder interface {
	Build(table rel.Table) string
}

type IndexBuilder interface {
	Build(index rel.Index) string
}
