// Package mysql wraps mysql driver as an adapter for REL.
//
// Usage:
//
//	// open mysql connection.
//	// note: `clientFoundRows=true` is required for update and delete to works correctly.
//	adapter, err := mysql.Open("root@(127.0.0.1:3306)/rel_test?clientFoundRows=true&charset=utf8&parseTime=True&loc=Local")
//	if err != nil {
//		panic(err)
//	}
//	defer adapter.Close()
//
//	// initialize REL's repo.
//	repo := rel.New(adapter)
package mysql

import (
	db "database/sql"
	"fmt"
	"strings"

	"github.com/go-rel/rel"
	"github.com/go-rel/sql"
	"github.com/go-rel/sql/builder"
)

// New mysql adapter using existing connection.
// Existing connection needs to be created with `clientFoundRows=true` options for update and delete to works correctly.
func New(database *db.DB) rel.Adapter {
	var (
		bufferFactory     = builder.BufferFactory{ArgumentPlaceholder: "?", BoolTrueValue: "true", BoolFalseValue: "false", Quoter: Quote{}, ValueConverter: ValueConvert{}}
		filterBuilder     = builder.Filter{}
		queryBuilder      = builder.Query{BufferFactory: bufferFactory, Filter: filterBuilder}
		onConflictBuilder = builder.OnConflict{Statement: "ON DUPLICATE KEY", UpdateStatement: "UPDATE", UseValues: true}
		InsertBuilder     = builder.Insert{BufferFactory: bufferFactory, InsertDefaultValues: true, OnConflict: onConflictBuilder}
		insertAllBuilder  = builder.InsertAll{BufferFactory: bufferFactory, OnConflict: onConflictBuilder}
		updateBuilder     = builder.Update{BufferFactory: bufferFactory, Query: queryBuilder, Filter: filterBuilder}
		deleteBuilder     = builder.Delete{BufferFactory: bufferFactory, Query: queryBuilder, Filter: filterBuilder}
		ddlBufferFactory  = builder.BufferFactory{InlineValues: true, BoolTrueValue: "true", BoolFalseValue: "false", Quoter: Quote{}, ValueConverter: ValueConvert{}}
		ddlQueryBuilder   = builder.Query{BufferFactory: ddlBufferFactory, Filter: filterBuilder}
		tableBuilder      = builder.Table{BufferFactory: ddlBufferFactory, ColumnMapper: columnMapper, DropKeyMapper: dropKeyMapper}
		indexBuilder      = builder.Index{BufferFactory: ddlBufferFactory, Query: ddlQueryBuilder, Filter: filterBuilder, DropIndexOnTable: true}
	)

	return &sql.SQL{
		QueryBuilder:     queryBuilder,
		InsertBuilder:    InsertBuilder,
		InsertAllBuilder: insertAllBuilder,
		UpdateBuilder:    updateBuilder,
		DeleteBuilder:    deleteBuilder,
		TableBuilder:     tableBuilder,
		IndexBuilder:     indexBuilder,
		Increment:        getIncrement(database),
		ErrorMapper:      errorMapper,
		DB:               database,
	}
}

// Open mysql connection using dsn.
func Open(dsn string) (rel.Adapter, error) {
	var database, err = db.Open("mysql", rewriteDsn(dsn))
	return New(database), err
}

func rewriteDsn(dsn string) string {
	// force clientFoundRows=true
	// this allows not found record check when updating a record.
	if strings.ContainsRune(dsn, '?') {
		dsn += "&clientFoundRows=true"
	} else {
		dsn += "?clientFoundRows=true"
	}

	return dsn
}

// MustOpen mysql connection using dsn.
func MustOpen(dsn string) rel.Adapter {
	var (
		adapter, err = Open(dsn)
	)

	check(err)
	return adapter
}

func getIncrement(database *db.DB) int {
	var (
		variable  string
		increment int
	)

	if database != nil {
		check(database.QueryRow("SHOW VARIABLES LIKE 'auto_increment_increment';").Scan(&variable, &increment))
	}

	return increment
}

func errorMapper(err error) error {
	if err == nil {
		return nil
	}

	var (
		msg          = err.Error()
		errCodeSep   = ':'
		errCodeIndex = strings.IndexRune(msg, errCodeSep)
	)

	if errCodeIndex < 0 {
		errCodeIndex = 0
	}

	switch msg[:errCodeIndex] {
	case "Error 1062":
		return rel.ConstraintError{
			Key:  sql.ExtractString(msg, "key '", "'"),
			Type: rel.UniqueConstraint,
			Err:  err,
		}
	case "Error 1452":
		return rel.ConstraintError{
			Key:  sql.ExtractString(msg, "CONSTRAINT `", "`"),
			Type: rel.ForeignKeyConstraint,
			Err:  err,
		}
	default:
		return err
	}
}

func columnMapper(column *rel.Column) (string, int, int) {
	switch column.Type {
	case rel.JSON:
		return "JSON", 0, 0
	default:
		return sql.ColumnMapper(column)
	}
}

func dropKeyMapper(typ rel.KeyType) string {
	if typ == rel.ForeignKey {
		return "FOREIGN KEY"
	}

	panic(fmt.Sprintf("drop key: unsupported key type `%s`", typ))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
