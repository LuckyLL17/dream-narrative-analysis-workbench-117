package service

import (
	"time"

	"dream117/internal/analysis"
	"dream117/internal/domain"
	"dream117/internal/store"
	"dream117/internal/text"
)

type AnalysisService struct {
	store    *store.Store
	overview analysis.OverviewEngine
	emotions analysis.EmotionEngine
}

func NewAnalysisService(
	data *store.Store,
) *AnalysisService {
	return &AnalysisService{store: data}
}

func (
	s *AnalysisService,
) items(
	userID string,
	window analysis.Window,
) []domain.Dream {
	return s.store.ListDreams(userID, window.From, window.To, "")
}

func (
	s *AnalysisService,
) Analyze(
	userID,
	path string,
	window analysis.Window,
) interface{} {
	items := s.items(userID, window)
	switch path {
	case "overview":
		return s.overview.Build(items, s.store.ListElements(userID), window, time.Now().UTC())
	case "signals":
		return analysis.BuildSignalReport(items, window)
	case "patterns":
		return analysis.BuildPatternReport(items)
	case "facets":
		return analysis.BuildFacets(items)
	case "review":
		return analysis.BuildReviewDigest(items, window)
	case "themes":
		return analysis.ThemeStats(items, 20)
	case "emotions":
		return map[string]interface{}{"points": s.emotions.Trend(items), "counts": analysis.EmotionCounts(items), "sleep_buckets": analysis.SleepBuckets(items), "sleep_clarity_correlation": analysis.SleepClarityCorrelation(items)}
	case "wordcloud":
		return analysis.WordCloud(items, 40)
	case "mood":
		return analysis.BuildMoodReport(items)
	case "sleep-contrast":
		return analysis.SleepContrast(items)
	case "theme-pairs":
		return analysis.ThemeCooccurrence(items, 24)
	case "theme-timeline":
		return analysis.ThemeTimeline(items, 12)
	case "calendar":
		return analysis.BuildCalendar(items, window.From, window.To)
	case "phrases":
		return text.PhrasesAcrossDreams(items, 36)
	case "narratives":
		return s.narratives(items)
	default:
		return nil
	}
}

func (
	s *AnalysisService,
) Reflection(
	userID,
	dreamID string,
) (analysis.ReflectionReport, error) {
	d, ok := s.store.FindDream(
		userID, dreamID)
	if !ok {
		return analysis.ReflectionReport{}, domain.ErrNotFound
	}
	return analysis.BuildReflection(d), nil
}

func (
	s *AnalysisService,
) Similar(
	userID,
	dreamID string,
	limit int,
) ([]analysis.SimilarDream, error) {
	target, ok :=
		s.store.FindDream(
			userID, dreamID)
	if !ok {
		return nil, domain.ErrNotFound
	}
	return analysis.SimilarDreams(target, s.store.AllDreams(userID), limit), nil
}
func (
	s *AnalysisService,
) Search(
	userID string,
	filter domain.DreamFilter,
) domain.DreamPage {
	return s.store.SearchPage(userID, filter)
}

func (
	s *AnalysisService,
) Lexicon(
	query string,
) []text.LexiconEntry {
	return text.FindLexicon(query)
}

type NarrativeSample struct {
	ID      string
	Date    string
	Title   string
	Profile text.NarrativeProfile
	Phrases []text.Phrase
}

func (
	s *AnalysisService,
) Narratives(
	userID string,
	window analysis.Window,
) []NarrativeSample {
	return s.narratives(s.items(userID, window))
}

func (
	s *AnalysisService,
) narratives(
	items []domain.Dream,
) []NarrativeSample {
	result := make(
		[]NarrativeSample,
		0,
		len(items),
	)
	for i := range items {
		d := items[i]
		phrases := text.Phrases(d.Title, d.Content, 6)
		result = append(result,
			NarrativeSample{ID: d.ID, Date: d.DreamDate.Format("2006-01-02"), Title: d.Title, Profile: text.Profile(d.Title, d.Content), Phrases: phrases})
	}
	return result
}
