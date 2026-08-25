package service

import (
	"strings"
	"time"

	"dream117/internal/domain"
	"dream117/internal/store"
	"dream117/internal/text"
	"dream117/pkg/ids"
)

type ElementService struct{ store *store.Store }

func NewElementService(
	data *store.Store,
) *ElementService {
	return &ElementService{store: data}
}

func (
	s *ElementService,
) RecordUsage(
	userID string,
	names []string,
) error {
	now := time.Now().UTC()
	for i := range names {
		name := strings.TrimSpace(names[i])
		if name == "" {
			continue
		}
		kind := text.KindFor(name)
		element := domain.Element{ID: ids.New("element"), UserID: userID, Name: name, Kind: kind, Count: 1, CreatedAt: now, UpdatedAt: now}
		existing := s.find(userID, name, kind)
		if existing.ID != "" {
			element = existing
			element.Count++
			element.UpdatedAt = now
		}
		if err := s.store.UpsertElement(element); err != nil {
			return err
		}
	}
	return nil
}
func (
	s *ElementService,
) RemoveUsage(
	userID string,
	names []string,
) error {
	return s.store.RemoveElementsForDream(userID, names)
}
func (
	s *ElementService,
) List(
	userID string,
) []domain.Element {
	return s.store.ListElements(userID)
}
func (
	s *ElementService,
) Suggestions(
	userID,
	title,
	content string,
) []domain.Element {
	recommended := text.RecommendElements(title, content)
	known := s.List(userID)
	combined := append(recommended, known...)
	result := make(
		[]domain.Element, 0, len(recommended)+len(known))
	seen := map[string]bool{}
	for i := range combined {
		item := combined[i]
		key := strings.ToLower(strings.TrimSpace(item.Name))
		if !seen[key] {
			seen[key] = true
			result = append(
				result,
				item,
			)
		}
	}
	return result
}
func (
	s *ElementService,
) find(
	userID,
	name string,
	kind domain.ElementKind,
) domain.Element {
	items := s.store.ListElements(userID)
	for i := range items {
		item := items[i]
		if strings.EqualFold(item.Name, name) && item.Kind == kind {
			return item
		}
	}
	return domain.Element{}
}
