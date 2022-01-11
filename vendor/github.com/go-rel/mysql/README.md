# mysql

[![GoDoc](https://godoc.org/github.com/go-rel/mysql?status.svg)](https://pkg.go.dev/github.com/go-rel/mysql)
[![Tesst](https://github.com/go-rel/mysql/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/go-rel/mysql/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-rel/mysql)](https://goreportcard.com/report/github.com/go-rel/mysql)
[![codecov](https://codecov.io/gh/go-rel/mysql/branch/main/graph/badge.svg?token=56qOCsVPJF)](https://codecov.io/gh/go-rel/mysql)
[![Gitter chat](https://badges.gitter.im/go-rel/rel.png)](https://gitter.im/go-rel/rel)

MySQL adapter for REL.

## Example

```go
package main

import (
	"context"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-rel/mysql"
	"github.com/go-rel/rel"
)

func main() {
	// open mysql connection.
	// note: `clientFoundRows=true` is required for update and delete to works correctly.
	adapter, err := mysql.Open("root@(127.0.0.1:3306)/rel_test?clientFoundRows=true&charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	defer adapter.Close()

	// initialize REL's repo.
	repo := rel.New(adapter)
	repo.Ping(context.TODO())
}
```

## Example Replication (Source/Replica)

```go
package main

import (
	"context"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-rel/primaryreplica"
	"github.com/go-rel/mysql"
	"github.com/go-rel/rel"
)

func main() {
	// open mysql connection.
	// note: `clientFoundRows=true` is required for update and delete to works correctly.
	adapter := primaryreplica.New(
		mysql.MustOpen("root@(source:23306)/rel_test?charset=utf8&parseTime=True&loc=Local"),
		mysql.MustOpen("root@(replica:23307)/rel_test?charset=utf8&parseTime=True&loc=Local"),
	)
	defer adapter.Close()

	// initialize REL's repo.
	repo := rel.New(adapter)
	repo.Ping(context.TODO())
}
```

## Supported Driver

- github.com/go-sql-driver/mysql

## Supported Database

- MySQL 5 and 8
- MariaDB 10

## Testing

### Start MariaDB server in Docker

```console
docker run -it --rm -p 3307:3306 -e "MARIADB_ROOT_PASSWORD=test" -e "MARIADB_DATABASE=rel_test" mariadb:10
```

### Run tests

```console
MYSQL_DATABASE="root:test@tcp(localhost:3307)/rel_test" go test ./...
```
