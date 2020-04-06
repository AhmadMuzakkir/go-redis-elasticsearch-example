package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store"
	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store/mock"

	"github.com/stretchr/testify/assert"
)

func TestTrack(t *testing.T) {
	expectedID := "1"
	var lastID string

	mockStore := &mock.Store{
		ViewTrackerStore: &mock.ViewTracker{
			OnTrack: func(ctx context.Context, v store.ViewTrack) error {
				lastID = v.ID

				return nil
			},
		},
	}

	handler := NewHandler(mockStore.ViewTracker(), nil, nil)

	type request struct {
		ID string `json:"id"`
	}

	tests := []struct {
		name     string
		body     func() []byte
		wantCode int
	}{
		{
			name:     "empty body",
			wantCode: http.StatusBadRequest,
			body: func() []byte {
				return nil
			},
		},
		{
			name:     "empty id",
			wantCode: http.StatusBadRequest,
			body: func() []byte {
				req := request{}

				body, err := json.Marshal(req)
				if err != nil {
					t.Fatal(err)
				}

				return body
			},
		},
		{
			name:     "success",
			wantCode: http.StatusOK,
			body: func() []byte {
				req := request{}
				req.ID = expectedID

				body, err := json.Marshal(req)
				if err != nil {
					t.Fatal(err)
				}

				return body
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := tc.body()

			request := httptest.NewRequest("POST", "/analytics", bytes.NewReader(body))
			request.Header.Add("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, request)

			response, err := ioutil.ReadAll(w.Body)
			assert.Nil(t, err)

			assert.Equal(t, tc.wantCode, w.Code, "status code")

			// If success, we expect response equal to request.
			if w.Code == 200 {
				assert.NoError(t, compareJSON(body, response))

				assert.Equal(t, expectedID, lastID)
			}
		})
	}
}

func TestRetrieve(t *testing.T) {
	mockStore := &mock.Store{
		ViewRetrieverStore: &mock.ViewRetriever{
			OnRetrieve: func(ctx context.Context, id string, ranges ...store.Range) (counts []store.ViewCount, err error) {
				if id == "1" {
					return []store.ViewCount{
						{
							Description: "1 minute ago",
							Count:       0,
						},
					}, nil
				}
				if id == "2" {
					return []store.ViewCount{
						{
							Description: "1 minute ago",
							Count:       1,
						},
					}, nil
				}
				return
			},
		},
	}

	type response struct {
		ID     string            `json:"id"`
		Counts []store.ViewCount `json:"counts"`
	}

	handler := NewHandler(nil, mockStore.ViewRetriever(), nil)

	tests := []struct {
		name     string
		ID       string
		wantCode int
		response func() []byte
	}{
		{
			name:     "zero view",
			ID:       "1",
			wantCode: http.StatusOK,
			response: func() []byte {
				var res response
				res.ID = "1"
				res.Counts = []store.ViewCount{
					{
						Description: "1 minute ago",
						Count:       0,
					},
				}

				b, err := json.Marshal(res)
				if err != nil {
					t.Fatal(err)
				}

				return b
			},
		},
		{
			name:     "1 view",
			ID:       "2",
			wantCode: http.StatusOK,
			response: func() []byte {
				var res response
				res.ID = "2"
				res.Counts = []store.ViewCount{
					{
						Description: "1 minute ago",
						Count:       1,
					},
				}

				b, err := json.Marshal(res)
				if err != nil {
					t.Fatal(err)
				}

				return b
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/analytics/"+tc.ID, nil)
			request.Header.Add("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, request)

			response, err := ioutil.ReadAll(w.Body)
			assert.Nil(t, err)

			assert.Equal(t, tc.wantCode, w.Code, "status code")

			// If success, we expect response equal to request.
			if w.Code == 200 {
				assert.NoError(t, compareJSON(tc.response(), response))
			}
		})
	}
}

func compareJSON(expected, actual []byte) error {
	if bytes.Equal(bytes.TrimSpace(actual), expected) {
		return nil
	}

	return fmt.Errorf("expected response `%s`, found `%s`", string(expected), string(actual))
}
