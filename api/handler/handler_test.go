package handler

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type responseRecorder struct {
	*httptest.ResponseRecorder
}

func (crr *responseRecorder) CloseNotify() <-chan bool {
	return nil
}

func (crr *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

func (crr *responseRecorder) Pusher() http.Pusher {
	return nil
}

func (crr *responseRecorder) Size() int {
	return 0
}

func (crr *responseRecorder) Status() int {
	return 0
}

func (crr *responseRecorder) WriteHeaderNow() {}

func (crr *responseRecorder) Written() bool {
	return true
}

func TestRender(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		response string
	}{
		{
			name:     "message",
			data:     "lorem",
			response: `{"message":"lorem"}`,
		},
		{
			name:     "error",
			data:     errors.New("system error"),
			response: `{"error":"system error"}`,
		},
		{
			name:     "nil",
			data:     nil,
			response: ``,
		},
		{
			name: "struct",
			data: struct {
				ID int `json:"id"`
			}{ID: 1},
			response: `{"id":1}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				rr = httptest.NewRecorder()
				c  = &gin.Context{Writer: &responseRecorder{rr}}
			)

			render(c, test.data, 200)
			if test.response != "" {
				assert.JSONEq(t, test.response, rr.Body.String())
			} else {
				assert.Equal(t, test.response, rr.Body.String())
			}
		})
	}
}
