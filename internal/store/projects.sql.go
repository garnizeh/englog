// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: projects.sql

package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createProject = `-- name: CreateProject :one

INSERT INTO projects (
    name, description, color, status, start_date, end_date, created_by, is_default
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING id, name, description, color, status, start_date, end_date, created_by, is_default, created_at, updated_at
`

type CreateProjectParams struct {
	Name        string      `db:"name" json:"name"`
	Description pgtype.Text `db:"description" json:"description"`
	Color       pgtype.Text `db:"color" json:"color"`
	Status      pgtype.Text `db:"status" json:"status"`
	StartDate   pgtype.Date `db:"start_date" json:"start_date"`
	EndDate     pgtype.Date `db:"end_date" json:"end_date"`
	CreatedBy   uuid.UUID   `db:"created_by" json:"created_by"`
	IsDefault   pgtype.Bool `db:"is_default" json:"is_default"`
}

// EngLog Project Management Queries
// Project CRUD operations and statistics
func (q *Queries) CreateProject(ctx context.Context, arg CreateProjectParams) (Project, error) {
	row := q.db.QueryRow(ctx, createProject,
		arg.Name,
		arg.Description,
		arg.Color,
		arg.Status,
		arg.StartDate,
		arg.EndDate,
		arg.CreatedBy,
		arg.IsDefault,
	)
	var i Project
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Color,
		&i.Status,
		&i.StartDate,
		&i.EndDate,
		&i.CreatedBy,
		&i.IsDefault,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteProject = `-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1 AND created_by = $2
`

type DeleteProjectParams struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedBy uuid.UUID `db:"created_by" json:"created_by"`
}

func (q *Queries) DeleteProject(ctx context.Context, arg DeleteProjectParams) error {
	_, err := q.db.Exec(ctx, deleteProject, arg.ID, arg.CreatedBy)
	return err
}

const getActiveProjectsByUser = `-- name: GetActiveProjectsByUser :many
SELECT id, name, description, color, status, start_date, end_date, created_by, is_default, created_at, updated_at FROM projects
WHERE created_by = $1 AND status = 'active'
ORDER BY is_default DESC, name ASC
`

func (q *Queries) GetActiveProjectsByUser(ctx context.Context, createdBy uuid.UUID) ([]Project, error) {
	rows, err := q.db.Query(ctx, getActiveProjectsByUser, createdBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Project{}
	for rows.Next() {
		var i Project
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Color,
			&i.Status,
			&i.StartDate,
			&i.EndDate,
			&i.CreatedBy,
			&i.IsDefault,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getProjectByID = `-- name: GetProjectByID :one
SELECT id, name, description, color, status, start_date, end_date, created_by, is_default, created_at, updated_at FROM projects
WHERE id = $1
`

func (q *Queries) GetProjectByID(ctx context.Context, id uuid.UUID) (Project, error) {
	row := q.db.QueryRow(ctx, getProjectByID, id)
	var i Project
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Color,
		&i.Status,
		&i.StartDate,
		&i.EndDate,
		&i.CreatedBy,
		&i.IsDefault,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getProjectStats = `-- name: GetProjectStats :one
SELECT
    COUNT(le.id) as total_entries,
    SUM(le.duration_minutes) as total_minutes,
    AVG(le.duration_minutes) as avg_duration,
    COUNT(DISTINCT le.user_id) as contributors_count,
    MIN(le.start_time) as first_activity,
    MAX(le.start_time) as last_activity
FROM log_entries le
WHERE le.project_id = $1
`

type GetProjectStatsRow struct {
	TotalEntries      int64       `db:"total_entries" json:"total_entries"`
	TotalMinutes      int64       `db:"total_minutes" json:"total_minutes"`
	AvgDuration       float64     `db:"avg_duration" json:"avg_duration"`
	ContributorsCount int64       `db:"contributors_count" json:"contributors_count"`
	FirstActivity     interface{} `db:"first_activity" json:"first_activity"`
	LastActivity      interface{} `db:"last_activity" json:"last_activity"`
}

func (q *Queries) GetProjectStats(ctx context.Context, projectID pgtype.UUID) (GetProjectStatsRow, error) {
	row := q.db.QueryRow(ctx, getProjectStats, projectID)
	var i GetProjectStatsRow
	err := row.Scan(
		&i.TotalEntries,
		&i.TotalMinutes,
		&i.AvgDuration,
		&i.ContributorsCount,
		&i.FirstActivity,
		&i.LastActivity,
	)
	return i, err
}

const getProjectsByUser = `-- name: GetProjectsByUser :many
SELECT id, name, description, color, status, start_date, end_date, created_by, is_default, created_at, updated_at FROM projects
WHERE created_by = $1
ORDER BY is_default DESC, name ASC
`

func (q *Queries) GetProjectsByUser(ctx context.Context, createdBy uuid.UUID) ([]Project, error) {
	rows, err := q.db.Query(ctx, getProjectsByUser, createdBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Project{}
	for rows.Next() {
		var i Project
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Color,
			&i.Status,
			&i.StartDate,
			&i.EndDate,
			&i.CreatedBy,
			&i.IsDefault,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getProjectsWithActivity = `-- name: GetProjectsWithActivity :many
SELECT
    p.id, p.name, p.description, p.color, p.status, p.start_date, p.end_date, p.created_by, p.is_default, p.created_at, p.updated_at,
    COUNT(le.id) as entry_count,
    SUM(le.duration_minutes) as total_minutes
FROM projects p
LEFT JOIN log_entries le ON p.id = le.project_id
WHERE p.created_by = $1
GROUP BY p.id
ORDER BY p.is_default DESC, entry_count DESC
`

type GetProjectsWithActivityRow struct {
	ID           uuid.UUID          `db:"id" json:"id"`
	Name         string             `db:"name" json:"name"`
	Description  pgtype.Text        `db:"description" json:"description"`
	Color        pgtype.Text        `db:"color" json:"color"`
	Status       pgtype.Text        `db:"status" json:"status"`
	StartDate    pgtype.Date        `db:"start_date" json:"start_date"`
	EndDate      pgtype.Date        `db:"end_date" json:"end_date"`
	CreatedBy    uuid.UUID          `db:"created_by" json:"created_by"`
	IsDefault    pgtype.Bool        `db:"is_default" json:"is_default"`
	CreatedAt    pgtype.Timestamptz `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamptz `db:"updated_at" json:"updated_at"`
	EntryCount   int64              `db:"entry_count" json:"entry_count"`
	TotalMinutes int64              `db:"total_minutes" json:"total_minutes"`
}

func (q *Queries) GetProjectsWithActivity(ctx context.Context, createdBy uuid.UUID) ([]GetProjectsWithActivityRow, error) {
	rows, err := q.db.Query(ctx, getProjectsWithActivity, createdBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetProjectsWithActivityRow{}
	for rows.Next() {
		var i GetProjectsWithActivityRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Color,
			&i.Status,
			&i.StartDate,
			&i.EndDate,
			&i.CreatedBy,
			&i.IsDefault,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.EntryCount,
			&i.TotalMinutes,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserDefaultProject = `-- name: GetUserDefaultProject :one
SELECT id, name, description, color, status, start_date, end_date, created_by, is_default, created_at, updated_at FROM projects
WHERE created_by = $1 AND is_default = true
LIMIT 1
`

func (q *Queries) GetUserDefaultProject(ctx context.Context, createdBy uuid.UUID) (Project, error) {
	row := q.db.QueryRow(ctx, getUserDefaultProject, createdBy)
	var i Project
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Color,
		&i.Status,
		&i.StartDate,
		&i.EndDate,
		&i.CreatedBy,
		&i.IsDefault,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const setProjectAsDefault = `-- name: SetProjectAsDefault :exec
BEGIN
`

func (q *Queries) SetProjectAsDefault(ctx context.Context) error {
	_, err := q.db.Exec(ctx, setProjectAsDefault)
	return err
}

const updateProject = `-- name: UpdateProject :one
UPDATE projects
SET name = $2, description = $3, color = $4, status = $5,
    start_date = $6, end_date = $7, is_default = $8, updated_at = NOW()
WHERE id = $1 AND created_by = $9
RETURNING id, name, description, color, status, start_date, end_date, created_by, is_default, created_at, updated_at
`

type UpdateProjectParams struct {
	ID          uuid.UUID   `db:"id" json:"id"`
	Name        string      `db:"name" json:"name"`
	Description pgtype.Text `db:"description" json:"description"`
	Color       pgtype.Text `db:"color" json:"color"`
	Status      pgtype.Text `db:"status" json:"status"`
	StartDate   pgtype.Date `db:"start_date" json:"start_date"`
	EndDate     pgtype.Date `db:"end_date" json:"end_date"`
	IsDefault   pgtype.Bool `db:"is_default" json:"is_default"`
	CreatedBy   uuid.UUID   `db:"created_by" json:"created_by"`
}

func (q *Queries) UpdateProject(ctx context.Context, arg UpdateProjectParams) (Project, error) {
	row := q.db.QueryRow(ctx, updateProject,
		arg.ID,
		arg.Name,
		arg.Description,
		arg.Color,
		arg.Status,
		arg.StartDate,
		arg.EndDate,
		arg.IsDefault,
		arg.CreatedBy,
	)
	var i Project
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Color,
		&i.Status,
		&i.StartDate,
		&i.EndDate,
		&i.CreatedBy,
		&i.IsDefault,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
