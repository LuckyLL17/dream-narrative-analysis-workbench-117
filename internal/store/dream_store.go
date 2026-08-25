package store

import (
	"strings"
	"time"

	"dream117/internal/domain"
	"dream117/pkg/collections"
)

func (
	s *Store,
) SaveDream(
	d domain.Dream,
) error {
	return s.Update(func(data *domain.Database) error { data.Dreams[d.ID] = d; return nil })
}

func (
	s *Store,
) DeleteDream(
	userID,
	dreamID string,
) error {
	return s.Update(func(data *domain.Database) error {
		if d, ok := data.Dreams[dreamID]; !ok || d.UserID != userID {
			return domain.ErrNotFound
		}
		delete(data.Dreams, dreamID)
		return nil
	})
}

func (
	s *Store,
) FindDream(
	userID,
	dreamID string,
) (domain.Dream, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.data.Dreams[dreamID]
	return d, ok && d.UserID == userID
}

func (
	s *Store,
) ListDreams(
	userID string,
	from,
	to time.Time,
	query string,
) []domain.Dream {
	s.mu.RLock()
	defer s.mu.RUnlock()
	query = strings.ToLower(
		strings.TrimSpace(query))
	items := collections.FilterMap(s.data.Dreams, func(_ string, d domain.Dream) bool {
		if d.UserID != userID || d.DreamDate.Before(from) || d.DreamDate.After(to) {
			return false
		}
		return query == "" || strings.Contains(strings.ToLower(d.Title+" "+d.Content+" "+strings.Join(d.Tags, " ")), query)
	})
	collections.SortBy(
		items,
		newestDreamFirst,
	)
	return items
}

func (
	s *Store,
) AllDreams(
	userID string,
) []domain.Dream {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := collections.FilterMap(s.data.Dreams, func(_ string, d domain.Dream) bool { return d.UserID == userID })
	collections.SortBy(
		items,
		oldestDreamFirst,
	)
	return items
}

func (
	s *Store,
) DreamCount(
	userID string,
) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(collections.FilterMap(s.data.Dreams, func(_ string, d domain.Dream) bool { return d.UserID == userID }))
}
