package services

import (
"context"
"errors"
"fmt"
"time"

"github.com/guigui42/go-react-starter/internal/models"
"github.com/guigui42/go-react-starter/pkg/logger"
"gorm.io/gorm"
)

// AccountLockoutService manages per-account login attempt tracking and lockout.
type AccountLockoutService struct {
db  *gorm.DB
log *logger.Logger
}

// NewAccountLockoutService creates a new account lockout service.
func NewAccountLockoutService(db *gorm.DB, log *logger.Logger) *AccountLockoutService {
return &AccountLockoutService{db: db, log: log}
}

// IsLocked checks if the given email is currently locked out.
func (s *AccountLockoutService) IsLocked(ctx context.Context, email string) (bool, time.Duration, error) {
var attempt models.LoginAttempt
if err := s.db.WithContext(ctx).Where("email = ?", email).First(&attempt).Error; err != nil {
if errors.Is(err, gorm.ErrRecordNotFound) {
return false, 0, nil
}
if s.log != nil {
s.log.Error("Failed to check account lockout status", err)
}
return false, 0, fmt.Errorf("check lockout status: %w", err)
}
if attempt.IsLocked() {
remaining := time.Until(*attempt.LockedUntil)
return true, remaining, nil
}
return false, 0, nil
}

// RecordFailure records a failed login attempt for the given email.
func (s *AccountLockoutService) RecordFailure(ctx context.Context, email string) error {
var attempt models.LoginAttempt
now := time.Now()

err := s.db.WithContext(ctx).Where("email = ?", email).First(&attempt).Error
if err != nil {
if !errors.Is(err, gorm.ErrRecordNotFound) {
return fmt.Errorf("lookup login attempt: %w", err)
}
attempt = models.LoginAttempt{
Email:        email,
FailedCount:  1,
LastFailedAt: &now,
}
if err := s.db.WithContext(ctx).Create(&attempt).Error; err != nil {
return fmt.Errorf("create login attempt: %w", err)
}
return nil
}

if attempt.LockedUntil != nil && time.Now().After(*attempt.LockedUntil) {
attempt.FailedCount = 0
attempt.LockedUntil = nil
}

attempt.FailedCount++
attempt.LastFailedAt = &now

if attempt.FailedCount >= models.MaxFailedAttempts {
lockUntil := now.Add(models.LockoutDuration)
attempt.LockedUntil = &lockUntil
if s.log != nil {
s.log.Warn("Account locked due to too many failed login attempts: " + email)
}
}

if err := s.db.WithContext(ctx).Save(&attempt).Error; err != nil {
return fmt.Errorf("update login attempt: %w", err)
}
return nil
}

// RecordSuccess resets the failure counter for the given email after a successful login.
func (s *AccountLockoutService) RecordSuccess(ctx context.Context, email string) error {
if err := s.db.WithContext(ctx).Where("email = ?", email).Delete(&models.LoginAttempt{}).Error; err != nil {
return fmt.Errorf("clear login attempts: %w", err)
}
return nil
}

// CleanupExpired removes stale login attempt records.
func (s *AccountLockoutService) CleanupExpired(ctx context.Context) error {
cutoff := time.Now().Add(-1 * time.Hour)
if err := s.db.WithContext(ctx).Where("locked_until IS NOT NULL AND locked_until < ?", cutoff).Delete(&models.LoginAttempt{}).Error; err != nil {
if s.log != nil {
s.log.Error("Failed to cleanup expired lockout records", err)
}
return fmt.Errorf("cleanup expired lockouts: %w", err)
}
oldCutoff := time.Now().Add(-24 * time.Hour)
if err := s.db.WithContext(ctx).Where("locked_until IS NULL AND (last_failed_at IS NULL OR last_failed_at < ?)", oldCutoff).Delete(&models.LoginAttempt{}).Error; err != nil {
if s.log != nil {
s.log.Error("Failed to cleanup stale login attempt records", err)
}
return fmt.Errorf("cleanup stale login attempts: %w", err)
}
return nil
}
