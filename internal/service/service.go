package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"firegoals/internal/auth"
	"firegoals/internal/repo"
)

type Service struct {
	Repo      *repo.Repo
	Auth      *auth.Manager
	TokenTTL  time.Duration
	RefreshTT time.Duration
}

func New(repo *repo.Repo, authManager *auth.Manager) *Service {
	return &Service{Repo: repo, Auth: authManager, TokenTTL: time.Hour, RefreshTT: 7 * 24 * time.Hour}
}

func (s *Service) Register(ctx context.Context, email, password string) (string, error) {
	hash, err := s.Auth.HashPassword(password)
	if err != nil {
		return "", err
	}
	userID, err := s.Repo.CreateUser(ctx, email, hash)
	if err != nil {
		return "", err
	}
	_, err = s.Repo.CreateWorkspace(ctx, "Personal", "personal", userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, string, error) {
	userID, hash, err := s.Repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}
	if err := s.Auth.ComparePassword(hash, password); err != nil {
		return "", "", errors.New("invalid credentials")
	}
	accessToken, err := s.Auth.GenerateToken(userID, s.TokenTTL)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return "", "", err
	}
	if err := s.Repo.CreateSession(ctx, userID, refreshToken, time.Now().Add(s.RefreshTT)); err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (s *Service) generateRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
