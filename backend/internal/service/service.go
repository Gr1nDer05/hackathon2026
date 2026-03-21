package service

import (
	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/repository"
)

type AppService struct {
	repo   *repository.AppRepository
	mailer Mailer
}

func NewAppService(repo *repository.AppRepository, mailer Mailer) *AppService {
	return &AppService{repo: repo, mailer: mailer}
}

func (s *AppService) Status() domain.AppStatus {
	return s.repo.GetStatus()
}
