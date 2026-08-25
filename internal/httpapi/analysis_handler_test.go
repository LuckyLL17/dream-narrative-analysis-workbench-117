package httpapi

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"dream117/internal/domain"
)

// TestWriteJSONPreservesWordStatOrder guards the HTTP output chain: the
// wordcloud slice is already ordered by the unified weight, and writeJSON
// must marshal that order verbatim so the API, the report and the sentence
// summary all observe the same sequence. A regression that re-sorts (e.g.
// by count) at the handler layer would break this test.
func TestWriteJSONPreservesWordStatOrder(t *testing.T) {
	words := []domain.WordStat{
		{Word: "时钟", Count: 2, Weight: 164},
		{Word: "大雨", Count: 7, Weight: 129},
		{Word: "桥上", Count: 1, Weight: 24},
	}

	recorder := httptest.NewRecorder()
	writeJSON(recorder, 200, words)

	var payload struct {
		Data []struct {
			Word   string `json:"word"`
			Count  int    `json:"count"`
			Weight int    `json:"weight"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v (body=%s)", err, recorder.Body.String())
	}
	if len(payload.Data) != len(words) {
		t.Fatalf("expected %d items, got %d", len(words), len(payload.Data))
	}
	for i, want := range words {
		got := payload.Data[i]
		if got.Word != want.Word || got.Count != want.Count || got.Weight != want.Weight {
			t.Fatalf("item %d mismatch: got %+v want %+v", i, got, want)
		}
	}

	body := recorder.Body.String()
	clockPos := strings.Index(body, `"时钟"`)
	rainPos := strings.Index(body, `"大雨"`)
	if clockPos == -1 || rainPos == -1 {
		t.Fatalf("missing words in body: %s", body)
	}
	if clockPos > rainPos {
		t.Fatalf("title term 时钟 must appear before body term 大雨 in JSON; body=%s", body)
	}
}
