package store

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"dream117/internal/domain"
	"dream117/pkg/jsonutil"
)

type Store struct {
	mu   sync.RWMutex
	path string
	data domain.Database
}

func Open(
	path string,
) (*Store, error) {
	store := &Store{path: path, data: domain.NewDatabase()}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) load() error {
	content, err := os.ReadFile(
		s.path)
	if errors.Is(err, os.ErrNotExist) {
		return s.persist()
	}
	if err != nil {
		return err
	}
	if len(content) == 0 {
		return s.persist()
	}
	var data domain.Database
	if err := jsonutil.Unmarshal(content, &data); err != nil {
		return err
	}
	ensureMaps(&data)
	s.data = data
	return nil
}

func ensureMaps(
	data *domain.Database,
) {
	value := reflect.ValueOf(data).Elem()
	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		if field.Kind() == reflect.Map && field.IsNil() {
			field.Set(reflect.MakeMap(field.Type()))
		}
	}
	if data.Schema == 0 {
		data.Schema = 1
	}
}

func (s *Store) persist() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	s.data.UpdatedAt = time.Now().UTC()
	content, err :=
		jsonutil.MarshalIndent(
			s.data, "", "  ")
	if err != nil {
		return err
	}
	temporary := s.path + ".tmp"
	if err := os.WriteFile(temporary, content, 0o600); err != nil {
		return err
	}
	return os.Rename(temporary, s.path)
}

func (s *Store) Snapshot() domain.Database {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneDatabase(s.data)
}

func (
	s *Store,
) Update(
	mutator func(*domain.Database) error,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := mutator(&s.data); err != nil {
		return err
	}
	ensureMaps(&s.data)
	return s.persist()
}

func cloneDatabase(
	source domain.Database,
) domain.Database {
	copyValue := domain.NewDatabase()
	copyValue.Schema, copyValue.UpdatedAt = source.Schema, source.UpdatedAt
	copyDatabaseMaps(&copyValue, &source)
	return copyValue
}

func copyDatabaseMaps(target, source *domain.Database) {
	destination := reflect.ValueOf(target).Elem()
	origin := reflect.ValueOf(source).Elem()
	for index := 0; index < origin.NumField(); index++ {
		from, to := origin.Field(index), destination.Field(index)
		if from.Kind() != reflect.Map {
			continue
		}
		to.Set(reflect.MakeMapWithSize(from.Type(), from.Len()))
		for _, key := range from.MapKeys() {
			to.SetMapIndex(key, from.MapIndex(key))
		}
	}
}
