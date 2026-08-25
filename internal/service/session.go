package service

import "dream117/internal/domain"

type SessionService struct{ auth *AuthService }

func NewSessionService(
	auth *AuthService,
) *SessionService {
	return &SessionService{auth: auth}
}
func (
	s *SessionService,
) Resolve(
	token string,
) (domain.User, error) {
	return s.auth.Resolve(token)
}
