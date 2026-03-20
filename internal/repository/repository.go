package repository

import "github.com/Gr1nDer05/Hackathon2026/internal/domain"

type AppRepository struct{}

func NewAppRepository() *AppRepository {
	return &AppRepository{}
}

func (r *AppRepository) GetStatus() domain.AppStatus {
	return domain.AppStatus{
		Status: "ok",
		Name:   "Hackathon2026",
	}
}
