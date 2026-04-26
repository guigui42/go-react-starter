package repository

import (
	"context"
	"time"

	"github.com/example/go-react-starter/internal/models"
	"gorm.io/gorm"
)

// CSPViolationRepository handles CSP violation data access.
type CSPViolationRepository struct {
	db *gorm.DB
}

// NewCSPViolationRepository creates a new CSP violation repository.
func NewCSPViolationRepository(db *gorm.DB) *CSPViolationRepository {
	return &CSPViolationRepository{db: db}
}

// Create stores a new CSP violation report.
func (r *CSPViolationRepository) Create(ctx context.Context, violation *models.CSPViolation) error {
	return r.db.WithContext(ctx).Create(violation).Error
}

// GetRecent returns the most recent CSP violations.
func (r *CSPViolationRepository) GetRecent(ctx context.Context, limit int) ([]models.CSPViolation, error) {
	var violations []models.CSPViolation
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Find(&violations).Error
	return violations, err
}

// CountTotal returns the total number of CSP violations.
func (r *CSPViolationRepository) CountTotal(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CSPViolation{}).
		Count(&count).Error
	return count, err
}

// CountSince returns the number of CSP violations since a given time.
func (r *CSPViolationRepository) CountSince(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CSPViolation{}).
		Where("created_at >= ?", since).
		Count(&count).Error
	return count, err
}

// GetTopBlockedURIs returns the most frequently blocked URIs.
func (r *CSPViolationRepository) GetTopBlockedURIs(ctx context.Context, limit int) ([]models.ViolationCount, error) {
	var results []models.ViolationCount
	err := r.db.WithContext(ctx).
		Model(&models.CSPViolation{}).
		Select("blocked_uri as value, COUNT(*) as count").
		Group("blocked_uri").
		Order("count DESC").
		Limit(limit).
		Scan(&results).Error
	return results, err
}

// GetTopViolatedDirectives returns the most frequently violated directives.
func (r *CSPViolationRepository) GetTopViolatedDirectives(ctx context.Context, limit int) ([]models.ViolationCount, error) {
	var results []models.ViolationCount
	err := r.db.WithContext(ctx).
		Model(&models.CSPViolation{}).
		Select("violated_directive as value, COUNT(*) as count").
		Group("violated_directive").
		Order("count DESC").
		Limit(limit).
		Scan(&results).Error
	return results, err
}

// GetSummary returns an aggregated summary of CSP violations.
func (r *CSPViolationRepository) GetSummary(ctx context.Context) (*models.CSPViolationSummary, error) {
	summary := &models.CSPViolationSummary{}

	// Total count
	total, err := r.CountTotal(ctx)
	if err != nil {
		return nil, err
	}
	summary.TotalCount = total

	// Last 24 hours
	last24h, err := r.CountSince(ctx, time.Now().Add(-24*time.Hour))
	if err != nil {
		return nil, err
	}
	summary.Last24HoursCount = last24h

	// Last 7 days
	last7d, err := r.CountSince(ctx, time.Now().Add(-7*24*time.Hour))
	if err != nil {
		return nil, err
	}
	summary.Last7DaysCount = last7d

	// Top blocked domains
	topBlocked, err := r.GetTopBlockedURIs(ctx, 10)
	if err != nil {
		return nil, err
	}
	summary.TopViolatedDomains = topBlocked

	// Top directives
	topDirectives, err := r.GetTopViolatedDirectives(ctx, 10)
	if err != nil {
		return nil, err
	}
	summary.TopDirectives = topDirectives

	// Recent violations
	recent, err := r.GetRecent(ctx, 20)
	if err != nil {
		return nil, err
	}
	summary.RecentViolations = recent

	return summary, nil
}

// DeleteOlderThan removes CSP violations older than the specified time.
// This can be used for cleanup/retention policies.
func (r *CSPViolationRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&models.CSPViolation{})
	return result.RowsAffected, result.Error
}
