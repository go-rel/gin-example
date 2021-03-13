package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/go-rel/gin-example/api/handler"
	"github.com/go-rel/gin-example/scores"
	"github.com/go-rel/gin-example/todos"
	"github.com/go-rel/rel"
	"go.uber.org/zap"
)

// NewMux api.
func NewMux(repository rel.Repository) *gin.Engine {
	var (
		logger, _      = zap.NewProduction()
		router         = gin.Default()
		scores         = scores.New(repository)
		todos          = todos.New(repository, scores)
		healthzHandler = handler.NewHealthz()
		todosHandler   = handler.NewTodos(repository, todos)
		scoreHandler   = handler.NewScore(repository)
	)

	healthzHandler.Add("database", repository)

	router.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(logger, true))
	router.Use(requestid.New())
	router.Use(cors.Default())

	healthzHandler.Mount(router.Group("/healthz"))
	todosHandler.Mount(router.Group("/todos"))
	scoreHandler.Mount(router.Group("/score"))

	return router
}
