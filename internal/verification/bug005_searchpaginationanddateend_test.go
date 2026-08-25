package verification

import (
	"dream117/internal/domain"
	"dream117/internal/store"
	"testing"
	"time"
)

func TestBug005Searchpaginationanddateend(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/dreams.json")
	if err != nil {
		t.Fatal(err)
	}
	end := time.Date(2026, 8, 2, 23, 59, 59, 999999999, time.UTC)
	for _, d := range []domain.Dream{
		{ID: "p1", UserID: "u", Title: "分页一", DreamDate: time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)},
		{ID: "p2", UserID: "u", Title: "分页二", DreamDate: time.Date(2026, 8, 2, 11, 0, 0, 0, time.UTC)},
		{ID: "p3", UserID: "u", Title: "分页三", DreamDate: end},
	} {
		if err := s.SaveDream(d); err != nil {
			t.Fatal(err)
		}
	}
	page := s.SearchPage("u", domain.DreamFilter{From: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), To: end, Query: "分页", Page: 2, PageSize: 2})
	if page.Total != 3 || page.TotalPages != 2 {
		t.Fatalf("pagination metadata wrong: total=%d total_pages=%d", page.Total, page.TotalPages)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "p1" {
		t.Fatalf("second page wrong: items=%v", page.Items)
	}
}

func TestBug005RegressionHealth(t *testing.T) {
	if got := (domain.DreamFilter{Page: 0, PageSize: 0}).Normalized().Page; got != 1 {
		t.Fatalf("normalized page=%d", got)
	}
}
