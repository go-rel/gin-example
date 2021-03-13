package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	logger, _ = zap.NewProduction(zap.Fields(zap.String("type", "handler")))
	// ErrBadRequest error.
	ErrBadRequest = errors.New("Bad Request")
)

func render(c *gin.Context, body interface{}, status int) {
	switch v := body.(type) {
	case string:
		c.JSON(status, struct {
			Message string `json:"message"`
		}{
			Message: v,
		})
	case error:
		c.JSON(status, struct {
			Error string `json:"error"`
		}{
			Error: v.Error(),
		})
	case nil:
		c.Status(status)
	default:
		c.JSON(status, body)
	}
}
