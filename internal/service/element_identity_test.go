package service

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dream117/internal/domain"
	"dream117/internal/store"
)

// newTestStore opens a store backed by an isolated temp file so each test
// starts from an empty database without touching real user data.
func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	dir := t.TempDir()
	data, err := store.Open(filepath.Join(dir, "dreams.json"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return data
}

// elementByName finds the first stored element whose name matches the given
// name case-insensitively, mirroring how a user scans the element list.
func elementByName(t *testing.T, elements *ElementService, userID, name string) (domain.Element, bool) {
	t.Helper()
	for _, e := range elements.List(userID) {
		if strings.EqualFold(e.Name, name) {
			return e, true
		}
	}
	return domain.Element{}, false
}

// TestElementIdentityIsCaseInsensitive walks the full element usage chain that
// the bug report describes: two dreams submitted with names that differ only
// in letter case ("Moon" then "moon") must collapse into one semantic element
// with a total count of 2, and deleting the tag using yet another case form
// ("MOON") must decrement that single element to 1 instead of silently leaving
// a duplicate or failing to subtract.
func TestElementIdentityIsCaseInsensitive(t *testing.T) {
	store := newTestStore(t)
	elements := NewElementService(store)
	const userID = "user-1"

	// First dream records "Moon".
	if err := elements.RecordUsage(userID, []string{"Moon"}); err != nil {
		t.Fatalf("record Moon: %v", err)
	}

	// Second dream records the same concept with a different case. Before the
	// fix this created a second element because Upsert/find matched exactly.
	if err := elements.RecordUsage(userID, []string{"moon"}); err != nil {
		t.Fatalf("record moon: %v", err)
	}

	list := elements.List(userID)
	if got := len(list); got != 1 {
		t.Fatalf("expected a single semantic element, got %d: %+v", got, list)
	}
	got, ok := elementByName(t, elements, userID, "moon")
	if !ok {
		t.Fatalf("expected a moon element to exist")
	}
	if got.Count != 2 {
		t.Fatalf("expected total count 2, got %d (element %+v)", got.Count, got)
	}

	// Deleting using a third case form must decrement the one element rather
	// than leave a dangling duplicate or miss the count.
	if err := elements.RemoveUsage(userID, []string{"MOON"}); err != nil {
		t.Fatalf("remove MOON: %v", err)
	}
	after, ok := elementByName(t, elements, userID, "moon")
	if !ok {
		t.Fatalf("expected moon element to remain after one removal")
	}
	if after.Count != 1 {
		t.Fatalf("expected count 1 after removal, got %d (element %+v)", after.Count, after)
	}
}

// TestElementIdentityMixedWithOtherElements ensures the case-folding fix only
// merges truly identical names: a different name and a genuinely different
// concept must still be tracked as separate elements with independent counts.
func TestElementIdentityMixedWithOtherElements(t *testing.T) {
	store := newTestStore(t)
	elements := NewElementService(store)
	const userID = "user-2"

	if err := elements.RecordUsage(userID, []string{"Moon", "太阳"}); err != nil {
		t.Fatalf("record first batch: %v", err)
	}
	if err := elements.RecordUsage(userID, []string{"MOON"}); err != nil {
		t.Fatalf("record MOON: %v", err)
	}

	list := elements.List(userID)
	if got := len(list); got != 2 {
		t.Fatalf("expected two distinct elements, got %d: %+v", got, list)
	}
	moon, ok := elementByName(t, elements, userID, "moon")
	if !ok {
		t.Fatalf("expected moon element to exist")
	}
	if moon.Count != 2 {
		t.Fatalf("expected moon count 2, got %d", moon.Count)
	}
	sun, ok := elementByName(t, elements, userID, "太阳")
	if !ok {
		t.Fatalf("expected 太阳 element to exist")
	}
	if sun.Count != 1 {
		t.Fatalf("expected 太阳 count 1, got %d", sun.Count)
	}
}

// TestElementIdentityTrimsWhitespace asserts that surrounding whitespace does
// not fragment identity either, since RecordUsage trims names and the dream
// tag dedup path already normalizes away leading/trailing spaces.
func TestElementIdentityTrimsWhitespace(t *testing.T) {
	store := newTestStore(t)
	elements := NewElementService(store)
	const userID = "user-3"

	if err := elements.RecordUsage(userID, []string{" Moon "}); err != nil {
		t.Fatalf("record ' Moon ': %v", err)
	}
	if err := elements.RecordUsage(userID, []string{"moon"}); err != nil {
		t.Fatalf("record moon: %v", err)
	}

	list := elements.List(userID)
	if got := len(list); got != 1 {
		t.Fatalf("expected one element after trimming, got %d: %+v", got, list)
	}
	if list[0].Count != 2 {
		t.Fatalf("expected count 2, got %d", list[0].Count)
	}
}

// TestDreamServiceDeleteDecrementsElement reproduces the exact end-to-end flow
// from the bug report: two dreams carrying the same concept under different
// cases are created through DreamService, then one dream is deleted, and the
// element count must drop to 1 — the path that previously failed because the
// delete entry used a case that did not match the originally stored element.
func TestDreamServiceDeleteDecrementsElement(t *testing.T) {
	data := newTestStore(t)
	elements := NewElementService(data)
	dreams := NewDreamService(data, elements)
	const userID = "user-4"
	base := time.Date(2026, 8, 25, 7, 0, 0, 0, time.UTC)

	first := makeDream(t, dreams, userID, base, "Moon")
	second := makeDream(t, dreams, userID, base.Add(24*time.Hour), "moon")

	list := elements.List(userID)
	if got := len(list); got != 1 {
		t.Fatalf("expected one element after two dreams, got %d: %+v", got, list)
	}
	if list[0].Count != 2 {
		t.Fatalf("expected element count 2 after two dreams, got %d", list[0].Count)
	}

	// Deleting the second dream must remove one usage of the shared element.
	if err := dreams.Delete(userID, second.ID); err != nil {
		t.Fatalf("delete second dream: %v", err)
	}
	remaining, ok := elementByName(t, elements, userID, "moon")
	if !ok {
		t.Fatalf("expected moon element to survive deletion")
	}
	if remaining.Count != 1 {
		t.Fatalf("expected count 1 after deleting one dream, got %d", remaining.Count)
	}

	// The first dream's own tag keeps the original case the user typed; it
	// should still match the stored element regardless of case.
	storedFirst, ok := data.FindDream(userID, first.ID)
	if !ok {
		t.Fatalf("expected first dream to persist")
	}
	if len(storedFirst.Tags) != 1 {
		t.Fatalf("expected first dream to keep one tag, got %v", storedFirst.Tags)
	}
}

// TestDreamServiceUpdateReusesElement covers the update path: when a dream is
// edited and its tag changes case, the element count must not duplicate or
// lose track. The old usage is removed and the new (case-different) usage is
// recorded against the same semantic element.
func TestDreamServiceUpdateReusesElement(t *testing.T) {
	data := newTestStore(t)
	elements := NewElementService(data)
	dreams := NewDreamService(data, elements)
	const userID = "user-5"
	base := time.Date(2026, 8, 25, 7, 0, 0, 0, time.UTC)

	original := makeDream(t, dreams, userID, base, "Moon")
	before, ok := elementByName(t, elements, userID, "moon")
	if !ok {
		t.Fatalf("expected moon element before update")
	}
	if before.Count != 1 {
		t.Fatalf("expected count 1 before update, got %d", before.Count)
	}

	// Edit the same dream to use a different case for its only tag.
	updated, err := dreams.Update(userID, original.ID, dreamInput(base, "moon"))
	if err != nil {
		t.Fatalf("update dream: %v", err)
	}
	if len(updated.Tags) != 1 {
		t.Fatalf("expected updated dream to keep one tag, got %v", updated.Tags)
	}

	list := elements.List(userID)
	if got := len(list); got != 1 {
		t.Fatalf("expected one element after update, got %d: %+v", got, list)
	}
	after := list[0]
	if after.Count != 1 {
		t.Fatalf("expected count to stay 1 after case-only update, got %d", after.Count)
	}
}

func dreamInput(date time.Time, tags ...string) DreamInput {
	return DreamInput{
		Title:      "梦到夜空",
		Content:    "夜里看到发光的天体，非常清晰。",
		DreamDate:  date,
		WakeTime:   date.Add(8 * time.Hour),
		SleepHours: 7.5,
		Clarity:    8,
		Emotion:    "平静",
		Tags:       tags,
	}
}

func makeDream(t *testing.T, dreams *DreamService, userID string, date time.Time, tag string) domain.Dream {
	t.Helper()
	d, err := dreams.Create(userID, dreamInput(date, tag))
	if err != nil {
		t.Fatalf("create dream: %v", err)
	}
	return d
}
