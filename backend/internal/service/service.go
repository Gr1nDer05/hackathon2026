package service

import (
	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/repository"
)

type AppService struct {
	repo                         *repository.AppRepository
	reportTemplateDraftGenerator reportTemplateDraftGenerator
}

func NewAppService(repo *repository.AppRepository) *AppService {
	return &AppService{
		repo:                         repo,
		reportTemplateDraftGenerator: newReportTemplateDraftGeneratorFromEnv(),
	}
}

func (s *AppService) Status() domain.AppStatus {
	return s.repo.GetStatus()
}
