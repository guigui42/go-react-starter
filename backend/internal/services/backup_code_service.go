package services

import (
"crypto/rand"
"encoding/base32"
"fmt"
"time"

"github.com/google/uuid"
"github.com/guigui42/go-react-starter/internal/models"
"golang.org/x/crypto/bcrypt"
"gorm.io/gorm"
)

// BackupCodeService handles backup code generation and validation.
type BackupCodeService struct {
db         *gorm.DB
codeCount  int
codeLength int
bcryptCost int
}

// NewBackupCodeService creates a new backup code service.
func NewBackupCodeService(db *gorm.DB, codeCount, codeLength int) *BackupCodeService {
return &BackupCodeService{
db:         db,
codeCount:  codeCount,
codeLength: codeLength,
bcryptCost: models.BcryptCost,
}
}

// GenerateBackupCodes generates new backup codes for a user.
func (s *BackupCodeService) GenerateBackupCodes(userID uuid.UUID) ([]string, error) {
if err := s.db.Where("user_id = ?", userID).Delete(&models.UserBackupCode{}).Error; err != nil {
return nil, fmt.Errorf("failed to delete existing codes: %w", err)
}

codes := make([]string, s.codeCount)
backupCodes := make([]*models.UserBackupCode, s.codeCount)

for i := 0; i < s.codeCount; i++ {
code, err := s.generateRandomCode()
if err != nil {
return nil, fmt.Errorf("failed to generate code: %w", err)
}

hash, err := bcrypt.GenerateFromPassword([]byte(code), s.bcryptCost)
if err != nil {
return nil, fmt.Errorf("failed to hash code: %w", err)
}

codes[i] = s.formatCode(code)
backupCodes[i] = &models.UserBackupCode{
ID:        models.NewID(),
UserID:    userID,
CodeHash:  string(hash),
Used:      false,
CreatedAt: time.Now().UTC(),
}
}

if err := s.db.Create(&backupCodes).Error; err != nil {
return nil, fmt.Errorf("failed to save backup codes: %w", err)
}

return codes, nil
}

// VerifyAndConsumeBackupCode verifies a backup code and marks it as used.
func (s *BackupCodeService) VerifyAndConsumeBackupCode(userID uuid.UUID, code string) error {
cleanCode := s.cleanCode(code)

var backupCodes []models.UserBackupCode
if err := s.db.Where("user_id = ? AND used = ?", userID, false).Find(&backupCodes).Error; err != nil {
return fmt.Errorf("failed to load backup codes: %w", err)
}

for _, bc := range backupCodes {
if err := bcrypt.CompareHashAndPassword([]byte(bc.CodeHash), []byte(cleanCode)); err == nil {
now := time.Now().UTC()
bc.Used = true
bc.UsedAt = &now

if err := s.db.Save(&bc).Error; err != nil {
return fmt.Errorf("failed to mark code as used: %w", err)
}

return nil
}
}

return fmt.Errorf("invalid or already used backup code")
}

// CountUnusedCodes returns the number of unused backup codes for a user.
func (s *BackupCodeService) CountUnusedCodes(userID uuid.UUID) (int64, error) {
var count int64
err := s.db.Model(&models.UserBackupCode{}).
Where("user_id = ? AND used = ?", userID, false).
Count(&count).Error
return count, err
}

func (s *BackupCodeService) generateRandomCode() (string, error) {
bytes := make([]byte, s.codeLength)
if _, err := rand.Read(bytes); err != nil {
return "", err
}

code := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)
if len(code) > s.codeLength {
code = code[:s.codeLength]
}

return code, nil
}

func (s *BackupCodeService) formatCode(code string) string {
formatted := ""
for i, char := range code {
if i > 0 && i%4 == 0 {
formatted += "-"
}
formatted += string(char)
}
return formatted
}

func (s *BackupCodeService) cleanCode(code string) string {
cleaned := ""
for _, char := range code {
if char != '-' && char != ' ' {
cleaned += string(char)
}
}
return cleaned
}
