package store

import (
	"strings"

	"dream117/internal/domain"
	"dream117/pkg/collections"
)

func (
	s *Store,
) UpsertElement(
	element domain.Element,
) error {
	return s.Update(func(data *domain.Database) error {
		for id, current := range data.Elements {
			if current.UserID == element.UserID && strings.EqualFold(current.Name, element.Name) && current.Kind == element.Kind {
				element.ID, element.CreatedAt = id, current.CreatedAt
			}
		}
		data.Elements[element.ID] = element
		return nil
	})
}

func (
	s *Store,
) ListElements(
	userID string,
) []domain.Element {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := collections.FilterMap(s.data.Elements, func(_ string, element domain.Element) bool { return element.UserID == userID })
	return collections.SortByCountName(items,
		func(item domain.Element) int { return item.Count },
		func(item domain.Element) string { return item.Name })
}

func (
	s *Store,
) RemoveElementsForDream(
	userID string,
	tags []string,
) error {
	return s.Update(func(data *domain.Database) error {
		for i := range tags {
			tag := tags[i]
			for id, element := range data.Elements {
				if element.UserID == userID && strings.EqualFold(element.Name, tag) && element.Count > 0 {
					element.Count--
					data.Elements[id] = element
				}
			}
		}
		return nil
	})
}
