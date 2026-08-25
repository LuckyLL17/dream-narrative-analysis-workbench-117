package service

import (
	"strings"
	"time"

	"dream117/internal/domain"
	"dream117/internal/security"
	"dream117/internal/store"
	"dream117/pkg/ids"
)

type AuthService struct {
	store  *store.Store
	tokens security.TokenCodec
}

func NewAuthService(
	data *store.Store,
	tokens security.TokenCodec,
) *AuthService {
	return &AuthService{store: data, tokens: tokens}
}

func (
	s *AuthService,
) Register(
	email,
	name,
	password string,
) (domain.User, string, error) {
	email, name = strings.ToLower(strings.TrimSpace(email)), strings.TrimSpace(name)
	if err := domain.ValidateUser(email, name, password); err != nil {
		return domain.User{}, "", err
	}
	if _, exists := s.store.FindUserByEmail(email); exists {
		return domain.User{}, "", domain.ErrConflict
	}
	hash, err :=
		security.HashPassword(
			password)
	if err != nil {
		return domain.User{}, "", err
	}
	now := time.Now().UTC()
	user := domain.User{ID: ids.New("user"), Email: email, DisplayName: name, PasswordHash: hash, CreatedAt: now, UpdatedAt: now}
	if err := s.store.SaveUser(user); err != nil {
		return domain.User{}, "", err
	}
	token, err := s.tokens.Issue(
		user.ID, user.DisplayName)
	return user, token, err
}

func (
	s *AuthService,
) Login(
	email,
	password string,
) (domain.User, string, error) {
	user, ok :=
		s.store.FindUserByEmail(
			strings.ToLower(strings.TrimSpace(email)))
	if !ok || !security.CheckPassword(user.PasswordHash, password) {
		return domain.User{}, "", domain.ErrUnauthorized
	}
	token, err := s.tokens.Issue(
		user.ID, user.DisplayName)
	return user, token, err
}

func (
	s *AuthService,
) Resolve(
	token string,
) (domain.User, error) {
	claims, err := s.tokens.Parse(
		token)
	if err != nil {
		return domain.User{}, domain.ErrUnauthorized
	}
	user, ok := s.store.FindUser(
		claims.UserID)
	if !ok {
		return domain.User{}, domain.ErrUnauthorized
	}
	return user, nil
}
