package verification

// UpsertElement is reached through RecordUsage so the element identity path stays covered.

import (
	"dream117/internal/service"
	"dream117/internal/store"
	"strings"
	"testing"
)

func TestBug007Elementcaseidentityandremoval(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/dreams.json")
	if err != nil {
		t.Fatal(err)
	}
	elements := service.NewElementService(s)
	if err := elements.RecordUsage("u", []string{"Moon"}); err != nil {
		t.Fatal(err)
	}
	if err := elements.RecordUsage("u", []string{"moon"}); err != nil {
		t.Fatal(err)
	}
	list := elements.List("u")
	if len(list) != 1 || list[0].Count != 2 || !strings.EqualFold(list[0].Name, "moon") {
		t.Fatalf("case variants split identity: %+v", list)
	}
	if err := elements.RemoveUsage("u", []string{"MOON"}); err != nil {
		t.Fatal(err)
	}
	list = elements.List("u")
	if len(list) != 1 || list[0].Count != 1 {
		t.Fatalf("case-insensitive removal failed: %+v", list)
	}
}

func TestBug007RegressionHealth(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/dreams.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.NewElementService(s).RecordUsage("u", []string{"钥匙"}); err != nil {
		t.Fatal(err)
	}
}
