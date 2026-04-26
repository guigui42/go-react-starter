package services

import (
"encoding/json"
"fmt"
"html/template"
"os"
"path/filepath"
"strings"
"sync"
texttemplate "text/template"
)

// EmailTemplateService handles rendering email templates with language support.
type EmailTemplateService struct {
templatesPath string
htmlCache     map[string]*template.Template
textCache     map[string]string
subjectsCache map[string]map[string]string
mu            sync.RWMutex
}

// NewEmailTemplateService creates a new email template service.
func NewEmailTemplateService(templatesPath string) *EmailTemplateService {
return &EmailTemplateService{
templatesPath: templatesPath,
htmlCache:     make(map[string]*template.Template),
textCache:     make(map[string]string),
subjectsCache: make(map[string]map[string]string),
}
}

// RenderedEmail contains the rendered email content.
type RenderedEmail struct {
Subject  string
HTMLBody string
TextBody string
}

// RenderTemplate renders an email template with the given data.
func (s *EmailTemplateService) RenderTemplate(templateName, lang string, data interface{}) (*RenderedEmail, error) {
languages := []string{lang, "en"}
if lang == "en" {
languages = []string{"en"}
}

var htmlBody, textBody, subject string
var err error

for _, l := range languages {
htmlBody, err = s.renderHTML(templateName, l, data)
if err == nil {
break
}
}
if htmlBody == "" {
return nil, fmt.Errorf("HTML template not found for %s", templateName)
}

for _, l := range languages {
textBody, err = s.renderText(templateName, l, data)
if err == nil {
break
}
}

for _, l := range languages {
subject, err = s.getSubject(templateName, l)
if err == nil && subject != "" {
break
}
}
if subject == "" {
return nil, fmt.Errorf("subject not found for %s", templateName)
}

return &RenderedEmail{
Subject:  subject,
HTMLBody: htmlBody,
TextBody: textBody,
}, nil
}

func (s *EmailTemplateService) renderHTML(templateName, lang string, data interface{}) (string, error) {
cacheKey := fmt.Sprintf("%s_%s", templateName, lang)

s.mu.RLock()
tmpl, exists := s.htmlCache[cacheKey]
s.mu.RUnlock()

if !exists {
filename := fmt.Sprintf("%s_%s.html", templateName, lang)
path := filepath.Join(s.templatesPath, filename)

content, err := os.ReadFile(path)
if err != nil {
return "", err
}

tmpl, err = template.New(filename).Parse(string(content))
if err != nil {
return "", fmt.Errorf("failed to parse HTML template: %w", err)
}

s.mu.Lock()
s.htmlCache[cacheKey] = tmpl
s.mu.Unlock()
}

var buf strings.Builder
if err := tmpl.Execute(&buf, data); err != nil {
return "", fmt.Errorf("failed to execute HTML template: %w", err)
}

return buf.String(), nil
}

func (s *EmailTemplateService) renderText(templateName, lang string, data interface{}) (string, error) {
cacheKey := fmt.Sprintf("%s_%s", templateName, lang)

s.mu.RLock()
content, exists := s.textCache[cacheKey]
s.mu.RUnlock()

if !exists {
filename := fmt.Sprintf("%s_%s.txt", templateName, lang)
path := filepath.Join(s.templatesPath, filename)

contentBytes, err := os.ReadFile(path)
if err != nil {
return "", err
}

content = string(contentBytes)

s.mu.Lock()
s.textCache[cacheKey] = content
s.mu.Unlock()
}

tmpl, err := texttemplate.New("text").Parse(content)
if err != nil {
return "", fmt.Errorf("failed to parse text template: %w", err)
}

var buf strings.Builder
if err := tmpl.Execute(&buf, data); err != nil {
return "", fmt.Errorf("failed to execute text template: %w", err)
}

return buf.String(), nil
}

func (s *EmailTemplateService) getSubject(templateName, lang string) (string, error) {
s.mu.RLock()
subjects, exists := s.subjectsCache[lang]
s.mu.RUnlock()

if !exists {
filename := fmt.Sprintf("subjects_%s.json", lang)
path := filepath.Join(s.templatesPath, filename)

content, err := os.ReadFile(path)
if err != nil {
return "", err
}

subjects = make(map[string]string)
if err := json.Unmarshal(content, &subjects); err != nil {
return "", fmt.Errorf("failed to parse subjects file: %w", err)
}

s.mu.Lock()
s.subjectsCache[lang] = subjects
s.mu.Unlock()
}

subject, ok := subjects[templateName]
if !ok {
return "", fmt.Errorf("subject not found for template: %s", templateName)
}

return subject, nil
}

// ClearCache clears the template cache.
func (s *EmailTemplateService) ClearCache() {
s.mu.Lock()
defer s.mu.Unlock()

s.htmlCache = make(map[string]*template.Template)
s.textCache = make(map[string]string)
s.subjectsCache = make(map[string]map[string]string)
}
