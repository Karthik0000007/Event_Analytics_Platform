package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/storage"
)

type fakeLister struct {
	letters             []storage.DeadLetter
	total               int
	err                 error
	gotLimit, gotOffset int
}

func (f *fakeLister) ListDeadLetters(_ context.Context, limit, offset int) ([]storage.DeadLetter, int, error) {
	f.gotLimit, f.gotOffset = limit, offset
	return f.letters, f.total, f.err
}

func getDLQ(h *DLQHandler, url string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.List(rec, httptest.NewRequest(http.MethodGet, url, nil))
	return rec
}

func TestDLQList_ReturnsPage(t *testing.T) {
	fake := &fakeLister{
		letters: []storage.DeadLetter{{ID: 1, ErrorKind: "permanent", FailedAt: time.Now()}},
		total:   7,
	}
	rec := getDLQ(&DLQHandler{Store: fake}, "/v1/dlq?limit=1&offset=2")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body struct {
		DeadLetters []storage.DeadLetter `json:"dead_letters"`
		Total       int                  `json:"total"`
		Limit       int                  `json:"limit"`
		Offset      int                  `json:"offset"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(body.DeadLetters) != 1 || body.Total != 7 || body.Limit != 1 || body.Offset != 2 {
		t.Errorf("unexpected body: %+v", body)
	}
	if fake.gotLimit != 1 || fake.gotOffset != 2 {
		t.Errorf("paging not passed to store: limit=%d offset=%d", fake.gotLimit, fake.gotOffset)
	}
}

func TestDLQList_ClampsBadPaging(t *testing.T) {
	fake := &fakeLister{}
	getDLQ(&DLQHandler{Store: fake}, "/v1/dlq?limit=abc&offset=-4")

	if fake.gotLimit != storage.DefaultDeadLetterLimit || fake.gotOffset != 0 {
		t.Errorf("expected default paging, got limit=%d offset=%d", fake.gotLimit, fake.gotOffset)
	}
}

func TestDLQList_EmptyIsArrayNotNull(t *testing.T) {
	rec := getDLQ(&DLQHandler{Store: &fakeLister{}}, "/v1/dlq")
	var body map[string]json.RawMessage
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if string(body["dead_letters"]) != "[]" {
		t.Errorf("expected an empty array, got %s", body["dead_letters"])
	}
}

func TestDLQList_StoreErrorReturns500(t *testing.T) {
	rec := getDLQ(&DLQHandler{Store: &fakeLister{err: errors.New("db down")}}, "/v1/dlq")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}
