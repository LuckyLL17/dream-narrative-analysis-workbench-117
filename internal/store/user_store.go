package store

import (
	"strings"

	"dream117/internal/domain"
)

func (
	s *Store,
) FindUserByEmail(
	email string,
) (domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.data.Users {
		user := s.data.Users[i]
		if strings.EqualFold(user.Email, email) {
			return user, true
		}
	}
	return domain.User{}, false
}

func (
	s *Store,
) FindUser(
	id string,
) (domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.data.Users[id]
	return user, ok
}

func (
	s *Store,
) SaveUser(
	user domain.User,
) error {
	return s.Update(func(data *domain.Database) error { data.Users[user.ID] = user; return nil })
}

func (
	s *Store,
) DeleteUser(
	id string,
) error {
	return s.Update(func(data *domain.Database) error { delete(data.Users, id); return nil })
}

func (s *Store) UserCount() int { s.mu.RLock(); defer s.mu.RUnlock(); return len(s.data.Users) }
