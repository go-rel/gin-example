package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-rel/gin-example/api/handler"
	"github.com/go-rel/gin-example/scores"
	"github.com/go-rel/reltest"
	"github.com/stretchr/testify/assert"
)

func TestScore_Index(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		path     string
		response string
		mockRepo func(repo *reltest.Repository)
	}{
		{
			name:     "ok",
			status:   http.StatusOK,
			path:     "/",
			response: `{"id":1, "total_point":10, "created_at":"0001-01-01T00:00:00Z", "updated_at":"0001-01-01T00:00:00Z"}`,
			mockRepo: func(repo *reltest.Repository) {
				repo.ExpectFind().Result(scores.Score{ID: 1, TotalPoint: 10})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				router     = gin.New()
				repository = reltest.New()
				handler    = handler.NewScore(repository)
				req, _     = http.NewRequest("GET", test.path, nil)
				rr         = httptest.NewRecorder()
			)

			if test.mockRepo != nil {
				test.mockRepo(repository)
			}

			handler.Mount(router.Group("/"))
			router.ServeHTTP(rr, req)

			assert.Equal(t, test.status, rr.Code)
			assert.JSONEq(t, test.response, rr.Body.String())

			repository.AssertExpectations(t)
		})
	}
}

func TestScore_Points(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		path     string
		response string
		mockRepo func(repo *reltest.Repository)
	}{
		{
			name:     "ok",
			status:   http.StatusOK,
			path:     "/points",
			response: `[{"id":1, "name": "todo completed", "count":1, "score_id": 0, "created_at":"0001-01-01T00:00:00Z", "updated_at":"0001-01-01T00:00:00Z"}]`,
			mockRepo: func(repo *reltest.Repository) {
				repo.ExpectFindAll().Result([]scores.Point{{ID: 1, Name: "todo completed", Count: 1}})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				router     = gin.New()
				repository = reltest.New()
				handler    = handler.NewScore(repository)
				req, _     = http.NewRequest("GET", test.path, nil)
				rr         = httptest.NewRecorder()
			)

			if test.mockRepo != nil {
				test.mockRepo(repository)
			}

			handler.Mount(router.Group("/"))
			router.ServeHTTP(rr, req)

			assert.Equal(t, test.status, rr.Code)
			assert.JSONEq(t, test.response, rr.Body.String())

			repository.AssertExpectations(t)
		})
	}
}
