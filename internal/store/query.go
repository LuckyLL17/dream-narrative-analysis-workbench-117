package store

import (
	"time"

	"dream117/internal/domain"
)

func FilterDreams(
	items []domain.Dream,
	from,
	to time.Time,
) []domain.Dream {
	filtered := make(
		[]domain.Dream,
		0,
		len(items),
	)
	for i := range items {
		d := items[i]
		if !d.DreamDate.Before(from) && !d.DreamDate.After(to) {
			filtered = append(
				filtered,
				d,
			)
		}
	}
	return filtered
}

func GroupByMonth(
	items []domain.Dream,
) map[string][]domain.Dream {
	groups := map[string][]domain.Dream{}
	for i := range items {
		d := items[i]
		key := d.DreamDate.Format("2006-01")
		groups[key] = append(groups[key], d)
	}
	return groups
}

func GroupByEmotion(
	items []domain.Dream,
) map[domain.Emotion][]domain.Dream {
	groups := map[domain.Emotion][]domain.Dream{}
	for i := range items {
		d := items[i]
		groups[d.Emotion] = append(groups[d.Emotion], d)
	}
	return groups
}
