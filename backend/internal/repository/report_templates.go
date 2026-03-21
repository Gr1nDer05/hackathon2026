package repository

import (
	"context"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
)

func (r *AppRepository) CreateReportTemplate(ctx context.Context, createdByUserID int64, input domain.CreateReportTemplateInput) (domain.ReportTemplate, error) {
	row := r.db.QueryRowContext(
		ctx,
		`INSERT INTO report_templates (name, description, template_body, created_by)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, name, description, template_body, created_by, created_at, updated_at`,
		input.Name,
		input.Description,
		input.TemplateBody,
		createdByUserID,
	)

	return scanReportTemplate(row)
}

func (r *AppRepository) ListReportTemplates(ctx context.Context, createdByUserID int64) ([]domain.ReportTemplate, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, name, description, template_body, created_by, created_at, updated_at
		 FROM report_templates
		 WHERE created_by = $1
		 ORDER BY updated_at DESC, id DESC`,
		createdByUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]domain.ReportTemplate, 0)
	for rows.Next() {
		template, scanErr := scanReportTemplate(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		templates = append(templates, template)
	}

	return templates, rows.Err()
}

func (r *AppRepository) GetReportTemplateByID(ctx context.Context, templateID int64, createdByUserID int64) (domain.ReportTemplate, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, name, description, template_body, created_by, created_at, updated_at
		 FROM report_templates
		 WHERE id = $1
		   AND created_by = $2`,
		templateID,
		createdByUserID,
	)

	return scanReportTemplate(row)
}

func (r *AppRepository) UpdateReportTemplate(ctx context.Context, templateID int64, createdByUserID int64, input domain.UpdateReportTemplateInput) (domain.ReportTemplate, error) {
	row := r.db.QueryRowContext(
		ctx,
		`UPDATE report_templates
		 SET name = $3,
		 	description = $4,
		 	template_body = $5,
		 	updated_at = NOW()
		 WHERE id = $1
		   AND created_by = $2
		 RETURNING id, name, description, template_body, created_by, created_at, updated_at`,
		templateID,
		createdByUserID,
		input.Name,
		input.Description,
		input.TemplateBody,
	)

	return scanReportTemplate(row)
}

func (r *AppRepository) DeleteReportTemplate(ctx context.Context, templateID int64, createdByUserID int64) (bool, error) {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM report_templates
		 WHERE id = $1
		   AND created_by = $2`,
		templateID,
		createdByUserID,
	)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func scanReportTemplate(scanner rowScanner) (domain.ReportTemplate, error) {
	var template domain.ReportTemplate
	var createdAt time.Time
	var updatedAt time.Time

	if err := scanner.Scan(
		&template.ID,
		&template.Name,
		&template.Description,
		&template.TemplateBody,
		&template.CreatedBy,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.ReportTemplate{}, err
	}

	template.CreatedAt = createdAt.Format(time.RFC3339)
	template.UpdatedAt = updatedAt.Format(time.RFC3339)
	return template, nil
}
