package domain

import (
	"time"

	"dream117/pkg/clock"
)

type User struct {
	ID           string
	Email        string
	DisplayName  string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Element struct {
	ID        string
	UserID    string
	Name      string
	Kind      ElementKind
	Count     int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Dream struct {
	ID             string
	UserID         string
	Title          string
	Content        string
	DreamDate      time.Time
	WakeTime       time.Time
	SleepHours     float64
	Clarity        int
	Emotion        Emotion
	RememberDetail bool
	Tags           []string
	Themes         []ThemeHit
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Revision       int
	Source         string
	AnalysisNote   string
}

type ThemeHit struct {
	Name       string
	Confidence float64
	Evidence   []string
}

type ThemeStat struct {
	Name       string
	Count      int
	Frequency  float64
	LastAppear time.Time
	AvgClarity float64
	AvgSleep   float64
	TopEmotion string
	Evidence   []string
	Trend      string
}

type EmotionPoint struct {
	Date       string
	Emotion    Emotion
	Intensity  float64
	SleepHours float64
	Clarity    int
}

type WordStat struct {
	Word   string
	Count  int
	Weight int
}

type Overview struct {
	From           time.Time
	To             time.Time
	TotalDreams    int
	CurrentMonth   int
	PreviousMonth  int
	MonthDelta     int
	AverageSleep   float64
	AverageClarity float64
	RememberRate   float64
	TopThemes      []ThemeStat
	TopElements    []Element
	EmotionCounts  map[string]int
	Insights       []Insight
	DailyCounts    []DailyCount
	ElementTrend   []ElementPoint
	ThemeMatrix    []ThemeEvidence
	WindowLabel    string
}

type Insight struct {
	Title      string
	Detail     string
	Confidence float64
	Tone       string
	Evidence   []string
}

type DailyCount struct {
	Date        string
	Count       int
	AverageMood float64
	Clarity     float64
	Emotions    map[string]int
}

type ElementPoint struct {
	Name  string
	Kind  string
	Count int
	First string
	Last  string
}

type ThemeEvidence struct {
	Theme       string
	Evidence    []string
	RecordCount int
	Strength    float64
}

type DreamFilter struct {
	From           time.Time
	To             time.Time
	Query          string
	Emotion        Emotion
	Theme          string
	MinimumClarity int
	MaximumSleep   float64
	RememberedOnly bool
	Page           int
	PageSize       int
}

type DreamPage struct {
	Items       []Dream
	Page        int
	PageSize    int
	Total       int
	TotalPages  int
	HasNext     bool
	HasPrevious bool
}

type MoodSummary struct {
	Dominant    Emotion
	Count       int
	Share       float64
	Buckets     map[string]int
	Transitions map[string]int
}

type SleepContrast struct {
	Group          string
	Count          int
	AverageClarity float64
	AverageSleep   float64
	Intensity      float64
	TopEmotion     string
}

type ReportDigest struct {
	WeekStart time.Time
	WeekEnd   time.Time
	Dreams    int
	Theme     string
	Emotion   string
	Change    string
	State     ReportState
}

type ExportBundle struct {
	ExportedAt time.Time
	User       User
	Dreams     []Dream
	Elements   []Element
	Reports    []WeeklyReport
	Format     string
}

type WeeklyReport struct {
	ID          string
	UserID      string
	WeekStart   time.Time
	WeekEnd     time.Time
	State       ReportState
	TotalDreams int
	TopThemes   []ThemeStat
	EmotionDist map[string]int
	AvgClarity  float64
	AvgSleep    float64
	Summary     string
	Suggestions []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AnalysisJob struct {
	ID          string
	UserID      string
	Kind        string
	Status      AnalysisStatus
	Attempts    int
	LastError   string
	ScheduledAt time.Time
	FinishedAt  *time.Time
}

type Database struct {
	Users     map[string]User
	Elements  map[string]Element
	Dreams    map[string]Dream
	Reports   map[string]WeeklyReport
	Jobs      map[string]AnalysisJob
	Schema    int
	UpdatedAt time.Time
}

func NewDatabase() Database {
	return Database{Users: map[string]User{}, Elements: map[string]Element{}, Dreams: map[string]Dream{}, Reports: map[string]WeeklyReport{}, Jobs: map[string]AnalysisJob{}, Schema: 1, UpdatedAt: time.Now().UTC()}
}

func (d Dream) ThemeNames() []string {
	result := make(
		[]string,
		0,
		len(d.Themes),
	)
	for i := range d.Themes {
		theme := d.Themes[i]
		result = append(result,
			theme.Name)
	}
	return result
}

func (
	d Dream,
) HasTheme(
	name string,
) bool {
	for i := range d.Themes {
		theme := d.Themes[i]
		if theme.Name == name {
			return true
		}
	}
	return false
}

func (
	d Dream,
) HasTag(
	name string,
) bool {
	for i := range d.Tags {
		tag := d.Tags[i]
		if tag == name {
			return true
		}
	}
	return false
}

func (f DreamFilter) Normalized() DreamFilter {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 24
	}
	if f.PageSize > 100 {
		f.PageSize = 100
	}
	if f.MinimumClarity < 0 {
		f.MinimumClarity = 0
	}
	if f.MinimumClarity > 10 {
		f.MinimumClarity = 10
	}
	if f.MaximumSleep < 0 {
		f.MaximumSleep = 0
	}
	// Date bounds are inclusive calendar-day ranges: a bare date (e.g.
	// "2026-08-02" parsed to midnight) is floored/ceiled so that every
	// record on the selected end day participates in the search, including
	// one timestamped at the very last nanosecond of that day. A zero bound
	// is left untouched so an unset from/to stays an open bound.
	if !f.From.IsZero() {
		f.From = clock.DayStart(f.From)
	}
	if !f.To.IsZero() {
		f.To = clock.DayEnd(f.To)
	}
	return f
}

func (p DreamPage) Empty() bool { return len(p.Items) == 0 }
