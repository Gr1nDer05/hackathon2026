package repository

import (
	"database/sql"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

type AppRepository struct {
	db *sql.DB
}

func NewAppRepository(db *sql.DB) *AppRepository {
	return &AppRepository{db: db}
}

func (r *AppRepository) GetStatus() domain.AppStatus {
	return domain.AppStatus{
		Status: "ok",
		Name:   "Hackathon2026",
	}
}
