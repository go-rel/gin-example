package handler

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-rel/gin-example/todos"
	"github.com/go-rel/rel"
	"github.com/go-rel/rel/where"
	"go.uber.org/zap"
)

type ctx int

const (
	loadKey string = "todosLoadKey"
)

// Todos for todos endpoints.
type Todos struct {
	repository rel.Repository
	todos      todos.Service
}

// Index handle GET /.
func (t Todos) Index(c *gin.Context) {
	var (
		result []todos.Todo
		filter = todos.Filter{
			Keyword: c.Query("keyword"),
		}
	)

	if str := c.Query("completed"); str != "" {
		completed := str == "true"
		filter.Completed = &completed
	}

	t.todos.Search(c, &result, filter)
	render(c, result, 200)
}

// Create handle POST /
func (t Todos) Create(c *gin.Context) {
	var (
		todo todos.Todo
	)

	if err := c.ShouldBindJSON(&todo); err != nil {
		logger.Warn("decode error", zap.Error(err))
		render(c, ErrBadRequest, 400)
		return
	}

	if err := t.todos.Create(c, &todo); err != nil {
		render(c, err, 422)
		return
	}

	c.Header("Location", fmt.Sprint(c.Request.RequestURI, "/", todo.ID))
	render(c, todo, 201)
}

// Show handle GET /{ID}
func (t Todos) Show(c *gin.Context) {
	var (
		todo = c.MustGet(loadKey).(todos.Todo)
	)

	render(c, todo, 200)
}

// Update handle PATCH /{ID}
func (t Todos) Update(c *gin.Context) {
	var (
		todo    = c.MustGet(loadKey).(todos.Todo)
		changes = rel.NewChangeset(&todo)
	)

	if err := c.ShouldBindJSON(&todo); err != nil {
		logger.Warn("decode error", zap.Error(err))
		render(c, ErrBadRequest, 400)
		return
	}

	if err := t.todos.Update(c, &todo, changes); err != nil {
		render(c, err, 422)
		return
	}

	render(c, todo, 200)
}

// Destroy handle DELETE /{ID}
func (t Todos) Destroy(c *gin.Context) {
	var (
		todo = c.MustGet(loadKey).(todos.Todo)
	)

	t.todos.Delete(c, &todo)
	render(c, nil, 204)
}

// Clear handle DELETE /
func (t Todos) Clear(c *gin.Context) {
	t.todos.Clear(c)
	render(c, nil, 204)
}

// Load is middleware that loads todos to context.
func (t Todos) Load(c *gin.Context) {
	var (
		id, _ = strconv.Atoi(c.Param("ID"))
		todo  todos.Todo
	)

	if err := t.repository.Find(c, &todo, where.Eq("id", id)); err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			render(c, err, 404)
			c.Abort()
			return
		}
		panic(err)
	}

	c.Set(loadKey, todo)
	c.Next()
}

// Mount handlers to router group.
func (t Todos) Mount(router *gin.RouterGroup) {
	router.GET("/", t.Index)
	router.POST("/", t.Create)
	router.GET("/:ID", t.Load, t.Show)
	router.PATCH("/:ID", t.Load, t.Update)
	router.DELETE("/:ID", t.Load, t.Destroy)
	router.DELETE("/", t.Clear)
}

// NewTodos handler.
func NewTodos(repository rel.Repository, todos todos.Service) Todos {
	return Todos{
		repository: repository,
		todos:      todos,
	}
}
