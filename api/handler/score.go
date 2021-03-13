package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-rel/gin-example/scores"
	"github.com/go-rel/rel"
)

// Score for score endpoints.
type Score struct {
	repository rel.Repository
}

// Index handle GET /
func (s Score) Index(c *gin.Context) {
	var (
		result scores.Score
	)

	s.repository.Find(c, &result)
	render(c, result, 200)
}

// Points handle Get /points
func (s Score) Points(c *gin.Context) {
	var (
		result []scores.Point
	)

	s.repository.FindAll(c, &result)
	render(c, result, 200)
}

// Mount handlers to router group.
func (s Score) Mount(router *gin.RouterGroup) {
	router.GET("/", s.Index)
	router.GET("/points", s.Points)
}

// NewScore handler.
func NewScore(repository rel.Repository) Score {
	return Score{
		repository: repository,
	}
}
