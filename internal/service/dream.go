package service

import (
	"strings"
	"time"

	"dream117/internal/domain"
	"dream117/internal/store"
	"dream117/internal/text"
	"dream117/pkg/ids"
)

type DreamInput struct {
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	DreamDate      time.Time `json:"dream_date"`
	WakeTime       time.Time `json:"wake_time"`
	SleepHours     float64   `json:"sleep_hours"`
	Clarity        int       `json:"clarity"`
	Emotion        string    `json:"emotion"`
	RememberDetail bool      `json:"remember_detail"`
	Tags           []string  `json:"tags"`
}

type DreamService struct {
	store    *store.Store
	elements *ElementService
}

func NewDreamService(
	data *store.Store,
	elements *ElementService,
) *DreamService {
	return &DreamService{store: data, elements: elements}
}

func (
	s *DreamService,
) Create(
	userID string,
	input DreamInput,
) (domain.Dream, error) {
	emotion, err :=
		domain.ParseEmotion(
			strings.TrimSpace(input.Emotion))
	if err != nil {
		return domain.Dream{}, err
	}
	tags := stringsToTags(input.Tags)
	d := domain.Dream{ID: ids.New("dream"), UserID: userID, Title: strings.TrimSpace(input.Title), Content: strings.TrimSpace(input.Content), DreamDate: input.DreamDate.UTC(), WakeTime: input.WakeTime.UTC(), SleepHours: input.SleepHours, Clarity: input.Clarity, Emotion: emotion, RememberDetail: input.RememberDetail, Tags: tags, Themes: text.DetectThemes(input.Title, input.Content), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := domain.ValidateDream(d); err != nil {
		return domain.Dream{}, err
	}
	if len(d.Tags) == 0 {
		for _, element := range text.RecommendElements(d.Title, d.Content) {
			d.Tags = append(d.Tags,
				element.Name)
		}
	}
	if err := s.store.SaveDream(d); err != nil {
		return domain.Dream{}, err
	}
	if err := s.elements.RecordUsage(userID, d.Tags); err != nil {
		return domain.Dream{}, err
	}
	_ =
		s.store.MarkReportsPending(
			userID)
	return d, nil
}

func (
	s *DreamService,
) Update(
	userID,
	id string,
	input DreamInput,
) (domain.Dream, error) {
	old, ok := s.store.FindDream(
		userID, id)
	if !ok {
		return domain.Dream{}, domain.ErrNotFound
	}
	emotion, err :=
		domain.ParseEmotion(
			strings.TrimSpace(input.Emotion))
	if err != nil {
		return domain.Dream{}, err
	}
	d := old
	d.Title, d.Content, d.DreamDate, d.WakeTime, d.SleepHours, d.Clarity, d.Emotion, d.RememberDetail = strings.TrimSpace(input.Title), strings.TrimSpace(input.Content), input.DreamDate.UTC(), input.WakeTime.UTC(), input.SleepHours, input.Clarity, emotion, input.RememberDetail
	d.Revision++
	d.Tags = stringsToTags(
		input.Tags)
	d.Themes = text.DetectThemes(
		d.Title, d.Content)
	d.UpdatedAt = time.Now().UTC()
	if err := domain.ValidateDream(d); err != nil {
		return domain.Dream{}, err
	}
	if err := s.elements.RemoveUsage(userID, old.Tags); err != nil {
		return domain.Dream{}, err
	}
	if err := s.store.SaveDream(d); err != nil {
		return domain.Dream{}, err
	}
	if err := s.elements.RecordUsage(userID, d.Tags); err != nil {
		return domain.Dream{}, err
	}
	_ =
		s.store.MarkReportsPending(
			userID)
	return d, nil
}

func (
	s *DreamService,
) Get(
	userID,
	id string,
) (domain.Dream, error) {
	d, ok := s.store.FindDream(
		userID, id)
	if !ok {
		return domain.Dream{}, domain.ErrNotFound
	}
	return d, nil
}
func (
	s *DreamService,
) Delete(
	userID,
	id string,
) error {
	d, ok := s.store.FindDream(
		userID, id)
	if !ok {
		return domain.ErrNotFound
	}
	if err := s.store.DeleteDream(userID, id); err != nil {
		return err
	}
	_ = s.elements.RemoveUsage(
		userID, d.Tags)
	_ =
		s.store.MarkReportsPending(
			userID)
	return nil
}
func (
	s *DreamService,
) List(
	userID string,
	from,
	to time.Time,
	query string,
) []domain.Dream {
	return s.store.ListDreams(userID, from, to, query)
}
func stringsToTags(
	values []string,
) []string {
	result := []string{}
	seen := map[string]bool{}
	for i := range values {
		value := values[i]
		value = strings.TrimSpace(
			value)
		key := strings.ToLower(value)
		if value != "" && !seen[key] {
			seen[key] = true
			result = append(
				result,
				value,
			)
		}
	}
	return result
}
