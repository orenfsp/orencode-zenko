package competition

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestListPaginationModes(t *testing.T) {
	for _, tc := range []struct {
		query      string
		status     int
		meta       bool
		countCalls int
		hasNext    bool
	}{
		{"?limit=1", 200, false, 0, false},
		{"?limit=1&with_meta=false", 200, false, 0, false},
		{"?limit=1&with_meta=true", 200, true, 1, true},
		{"?limit=1&offset=2&with_meta=true", 200, true, 1, false},
		{"?with_meta=1", 400, false, 0, false},
		{"?with_meta=", 400, false, 0, false},
		{"?with_meta=true&with_meta=false", 400, false, 0, false},
		{"?limit=0&with_meta=true", 400, false, 0, false},
	} {
		t.Run(tc.query, func(t *testing.T) {
			repo := &repositoryStub{items: []Competition{testCompetition()}, total: 3}
			w := httptest.NewRecorder()
			NewHandler(NewService(repo)).List(w, httptest.NewRequest("GET", "/api/v1/competitions"+tc.query, nil))
			if w.Code != tc.status || repo.countCalls != tc.countCalls {
				t.Fatalf("status=%d count calls=%d", w.Code, repo.countCalls)
			}
			if w.Code != 200 {
				return
			}
			var response struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if tc.meta {
				var page ListPageDTO
				if err := json.Unmarshal(response.Data, &page); err != nil {
					t.Fatal(err)
				}
				if page.Total != 3 || page.HasNext != tc.hasNext || page.Items == nil {
					t.Fatalf("page=%+v", page)
				}
			} else {
				var items []ResponseDTO
				if err := json.Unmarshal(response.Data, &items); err != nil || len(items) != 1 {
					t.Fatalf("items=%v err=%v", items, err)
				}
			}
		})
	}
}

func TestEmptyPagePreservesTotal(t *testing.T) {
	repo := &repositoryStub{total: 3}
	page, err := NewService(repo).ListWithMeta(t.Context(), 2, 10)
	if err != nil || page.Total != 3 || page.HasNext || page.Items == nil || len(page.Items) != 0 {
		t.Fatalf("page=%+v err=%v", page, err)
	}
}
