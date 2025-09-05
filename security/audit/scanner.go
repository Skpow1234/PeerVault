package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// VulnerabilityLevel represents the severity level of a vulnerability
type VulnerabilityLevel string

const (
	VulnerabilityLevelCritical VulnerabilityLevel = "critical"
	VulnerabilityLevelHigh     VulnerabilityLevel = "high"
	VulnerabilityLevelMedium   VulnerabilityLevel = "medium"
	VulnerabilityLevelLow      VulnerabilityLevel = "low"
	VulnerabilityLevelInfo     VulnerabilityLevel = "info"
)

// VulnerabilityType represents the type of vulnerability
type VulnerabilityType string

const (
	VulnerabilityTypeSQLInjection        VulnerabilityType = "sql_injection"
	VulnerabilityTypeXSS                 VulnerabilityType = "xss"
	VulnerabilityTypeCSRF                VulnerabilityType = "csrf"
	VulnerabilityTypeInsecureAuth        VulnerabilityType = "insecure_auth"
	VulnerabilityTypeWeakCrypto          VulnerabilityType = "weak_crypto"
	VulnerabilityTypeInsecureStorage     VulnerabilityType = "insecure_storage"
	VulnerabilityTypeInsecureNetwork     VulnerabilityType = "insecure_network"
	VulnerabilityTypePrivilegeEscalation VulnerabilityType = "privilege_escalation"
	VulnerabilityTypeDataExposure        VulnerabilityType = "data_exposure"
	VulnerabilityTypeInsecureConfig      VulnerabilityType = "insecure_config"
)

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID          string             `json:"id"`
	Type        VulnerabilityType  `json:"type"`
	Level       VulnerabilityLevel `json:"level"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	File        string             `json:"file,omitempty"`
	Line        int                `json:"line,omitempty"`
	Function    string             `json:"function,omitempty"`
	Severity    int                `json:"severity"` // 1-10 scale
	CVSS        float64            `json:"cvss,omitempty"`
	Remediation string             `json:"remediation"`
	References  []string           `json:"references,omitempty"`
	DetectedAt  time.Time          `json:"detected_at"`
}

// SecurityScanResult represents the result of a security scan
type SecurityScanResult struct {
	ScanID          string          `json:"scan_id"`
	StartTime       time.Time       `json:"start_time"`
	EndTime         time.Time       `json:"end_time"`
	Duration        time.Duration   `json:"duration"`
	Target          string          `json:"target"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	Summary         ScanSummary     `json:"summary"`
	Status          string          `json:"status"`
	Error           string          `json:"error,omitempty"`
}

// ScanSummary provides a summary of scan results
type ScanSummary struct {
	TotalVulnerabilities int `json:"total_vulnerabilities"`
	CriticalCount        int `json:"critical_count"`
	HighCount            int `json:"high_count"`
	MediumCount          int `json:"medium_count"`
	LowCount             int `json:"low_count"`
	InfoCount            int `json:"info_count"`
	RiskScore            int `json:"risk_score"` // 0-100
}

// SecurityScanner performs security vulnerability scans
type SecurityScanner struct {
	patterns map[VulnerabilityType][]string
	mu       sync.RWMutex
}

// NewSecurityScanner creates a new security scanner
func NewSecurityScanner() *SecurityScanner {
	scanner := &SecurityScanner{
		patterns: make(map[VulnerabilityType][]string),
	}

	// Initialize vulnerability patterns
	scanner.initializePatterns()

	return scanner
}

// initializePatterns initializes vulnerability detection patterns
func (ss *SecurityScanner) initializePatterns() {
	ss.patterns[VulnerabilityTypeSQLInjection] = []string{
		"SELECT.*FROM.*WHERE.*%s",
		"INSERT.*INTO.*VALUES.*%s",
		"UPDATE.*SET.*WHERE.*%s",
		"DELETE.*FROM.*WHERE.*%s",
		"db.Query.*%s",
		"db.Exec.*%s",
	}

	ss.patterns[VulnerabilityTypeXSS] = []string{
		"innerHTML.*%s",
		"document.write.*%s",
		"eval.*%s",
		"setTimeout.*%s",
		"setInterval.*%s",
	}

	ss.patterns[VulnerabilityTypeInsecureAuth] = []string{
		"password.*=.*\"\"",
		"token.*=.*\"\"",
		"secret.*=.*\"\"",
		"api_key.*=.*\"\"",
		"auth.*=.*\"\"",
	}

	ss.patterns[VulnerabilityTypeWeakCrypto] = []string{
		"md5\\(",
		"sha1\\(",
		"DES\\(",
		"RC4\\(",
		"MD5\\(",
		"SHA1\\(",
	}

	ss.patterns[VulnerabilityTypeInsecureStorage] = []string{
		"localStorage\\.setItem",
		"sessionStorage\\.setItem",
		"document\\.cookie.*=",
		"os\\.Setenv.*password",
		"os\\.Setenv.*secret",
	}

	ss.patterns[VulnerabilityTypeInsecureNetwork] = []string{
		"http://",
		"ftp://",
		"telnet://",
		"insecure.*transport",
		"skip.*tls.*verify",
	}
}

// ScanDirectory scans a directory for security vulnerabilities
func (ss *SecurityScanner) ScanDirectory(ctx context.Context, targetDir string) (*SecurityScanResult, error) {
	scanID := fmt.Sprintf("scan_%d", time.Now().UnixNano())
	startTime := time.Now()

	result := &SecurityScanResult{
		ScanID:    scanID,
		StartTime: startTime,
		Target:    targetDir,
		Status:    "running",
	}

	// Scan files
	vulnerabilities, err := ss.scanFiles(ctx, targetDir)
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	result.Vulnerabilities = vulnerabilities
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Status = "completed"
	result.Summary = ss.calculateSummary(vulnerabilities)

	return result, nil
}

// scanFiles scans all files in a directory for vulnerabilities
func (ss *SecurityScanner) scanFiles(ctx context.Context, targetDir string) ([]Vulnerability, error) {
	var vulnerabilities []Vulnerability
	var mu sync.Mutex

	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-source files
		if info.IsDir() || !ss.isSourceFile(path) {
			return nil
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Scan file for vulnerabilities
		fileVulns, err := ss.scanFile(path)
		if err != nil {
			return err
		}

		mu.Lock()
		vulnerabilities = append(vulnerabilities, fileVulns...)
		mu.Unlock()

		return nil
	})

	return vulnerabilities, err
}

// isSourceFile checks if a file is a source code file
func (ss *SecurityScanner) isSourceFile(path string) bool {
	sourceExts := []string{".go", ".js", ".ts", ".py", ".java", ".cpp", ".c", ".h", ".hpp", ".cs", ".php", ".rb", ".swift", ".kt"}

	for _, sourceExt := range sourceExts {
		if strings.HasSuffix(strings.ToLower(path), sourceExt) {
			return true
		}
	}

	return false
}

// scanFile scans a single file for vulnerabilities
func (ss *SecurityScanner) scanFile(filePath string) ([]Vulnerability, error) {
	var vulnerabilities []Vulnerability

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")

	// Check each vulnerability type
	for vulnType, patterns := range ss.patterns {
		for lineNum, line := range lines {
			for _, pattern := range patterns {
				if ss.matchesPattern(line, pattern) {
					vuln := Vulnerability{
						ID:          fmt.Sprintf("%s_%d_%d", vulnType, time.Now().UnixNano(), lineNum),
						Type:        vulnType,
						Level:       ss.getVulnerabilityLevel(vulnType),
						Title:       ss.getVulnerabilityTitle(vulnType),
						Description: ss.getVulnerabilityDescription(vulnType),
						File:        filePath,
						Line:        lineNum + 1,
						Severity:    ss.getVulnerabilitySeverity(vulnType),
						Remediation: ss.getVulnerabilityRemediation(vulnType),
						DetectedAt:  time.Now(),
					}
					vulnerabilities = append(vulnerabilities, vuln)
				}
			}
		}
	}

	return vulnerabilities, nil
}

// matchesPattern checks if a line matches a vulnerability pattern
func (ss *SecurityScanner) matchesPattern(line, pattern string) bool {
	// Simple pattern matching - in a real implementation, this would use regex
	line = strings.ToLower(strings.TrimSpace(line))
	pattern = strings.ToLower(pattern)

	// Replace %s with wildcard matching
	pattern = strings.ReplaceAll(pattern, "%s", ".*")

	// Simple contains check for now
	return strings.Contains(line, strings.ReplaceAll(pattern, ".*", ""))
}

// getVulnerabilityLevel returns the severity level for a vulnerability type
func (ss *SecurityScanner) getVulnerabilityLevel(vulnType VulnerabilityType) VulnerabilityLevel {
	switch vulnType {
	case VulnerabilityTypeSQLInjection, VulnerabilityTypePrivilegeEscalation:
		return VulnerabilityLevelCritical
	case VulnerabilityTypeXSS, VulnerabilityTypeInsecureAuth, VulnerabilityTypeDataExposure:
		return VulnerabilityLevelHigh
	case VulnerabilityTypeCSRF, VulnerabilityTypeWeakCrypto, VulnerabilityTypeInsecureStorage:
		return VulnerabilityLevelMedium
	case VulnerabilityTypeInsecureNetwork, VulnerabilityTypeInsecureConfig:
		return VulnerabilityLevelLow
	default:
		return VulnerabilityLevelInfo
	}
}

// getVulnerabilityTitle returns a title for a vulnerability type
func (ss *SecurityScanner) getVulnerabilityTitle(vulnType VulnerabilityType) string {
	switch vulnType {
	case VulnerabilityTypeSQLInjection:
		return "SQL Injection Vulnerability"
	case VulnerabilityTypeXSS:
		return "Cross-Site Scripting (XSS) Vulnerability"
	case VulnerabilityTypeCSRF:
		return "Cross-Site Request Forgery (CSRF) Vulnerability"
	case VulnerabilityTypeInsecureAuth:
		return "Insecure Authentication"
	case VulnerabilityTypeWeakCrypto:
		return "Weak Cryptographic Algorithm"
	case VulnerabilityTypeInsecureStorage:
		return "Insecure Data Storage"
	case VulnerabilityTypeInsecureNetwork:
		return "Insecure Network Communication"
	case VulnerabilityTypePrivilegeEscalation:
		return "Privilege Escalation Vulnerability"
	case VulnerabilityTypeDataExposure:
		return "Data Exposure Vulnerability"
	case VulnerabilityTypeInsecureConfig:
		return "Insecure Configuration"
	default:
		return "Security Vulnerability"
	}
}

// getVulnerabilityDescription returns a description for a vulnerability type
func (ss *SecurityScanner) getVulnerabilityDescription(vulnType VulnerabilityType) string {
	switch vulnType {
	case VulnerabilityTypeSQLInjection:
		return "Potential SQL injection vulnerability detected. Use parameterized queries to prevent SQL injection attacks."
	case VulnerabilityTypeXSS:
		return "Potential XSS vulnerability detected. Sanitize user input and use proper output encoding."
	case VulnerabilityTypeCSRF:
		return "Potential CSRF vulnerability detected. Implement CSRF tokens and validate requests."
	case VulnerabilityTypeInsecureAuth:
		return "Insecure authentication detected. Use strong authentication mechanisms and secure credential storage."
	case VulnerabilityTypeWeakCrypto:
		return "Weak cryptographic algorithm detected. Use strong, modern cryptographic algorithms."
	case VulnerabilityTypeInsecureStorage:
		return "Insecure data storage detected. Use secure storage mechanisms and encrypt sensitive data."
	case VulnerabilityTypeInsecureNetwork:
		return "Insecure network communication detected. Use encrypted connections (HTTPS/TLS)."
	case VulnerabilityTypePrivilegeEscalation:
		return "Potential privilege escalation vulnerability detected. Implement proper access controls."
	case VulnerabilityTypeDataExposure:
		return "Potential data exposure detected. Ensure sensitive data is properly protected."
	case VulnerabilityTypeInsecureConfig:
		return "Insecure configuration detected. Review and secure configuration settings."
	default:
		return "Security vulnerability detected. Review code for security best practices."
	}
}

// getVulnerabilitySeverity returns a severity score for a vulnerability type
func (ss *SecurityScanner) getVulnerabilitySeverity(vulnType VulnerabilityType) int {
	switch vulnType {
	case VulnerabilityTypeSQLInjection, VulnerabilityTypePrivilegeEscalation:
		return 9
	case VulnerabilityTypeXSS, VulnerabilityTypeInsecureAuth, VulnerabilityTypeDataExposure:
		return 7
	case VulnerabilityTypeCSRF, VulnerabilityTypeWeakCrypto, VulnerabilityTypeInsecureStorage:
		return 5
	case VulnerabilityTypeInsecureNetwork, VulnerabilityTypeInsecureConfig:
		return 3
	default:
		return 1
	}
}

// getVulnerabilityRemediation returns remediation advice for a vulnerability type
func (ss *SecurityScanner) getVulnerabilityRemediation(vulnType VulnerabilityType) string {
	switch vulnType {
	case VulnerabilityTypeSQLInjection:
		return "Use parameterized queries or prepared statements. Never concatenate user input directly into SQL queries."
	case VulnerabilityTypeXSS:
		return "Sanitize all user input and use proper output encoding. Implement Content Security Policy (CSP)."
	case VulnerabilityTypeCSRF:
		return "Implement CSRF tokens and validate all state-changing requests. Use SameSite cookie attributes."
	case VulnerabilityTypeInsecureAuth:
		return "Use strong authentication mechanisms, implement proper session management, and store credentials securely."
	case VulnerabilityTypeWeakCrypto:
		return "Use modern cryptographic algorithms (AES-256, SHA-256, etc.) and proper key management."
	case VulnerabilityTypeInsecureStorage:
		return "Encrypt sensitive data at rest and in transit. Use secure storage mechanisms and proper key management."
	case VulnerabilityTypeInsecureNetwork:
		return "Use encrypted connections (HTTPS/TLS) for all network communication. Implement certificate pinning."
	case VulnerabilityTypePrivilegeEscalation:
		return "Implement proper access controls, principle of least privilege, and regular access reviews."
	case VulnerabilityTypeDataExposure:
		return "Implement data classification, access controls, and encryption for sensitive data."
	case VulnerabilityTypeInsecureConfig:
		return "Review and secure all configuration settings. Use secure defaults and regular configuration audits."
	default:
		return "Review code for security best practices and implement appropriate security controls."
	}
}

// calculateSummary calculates a summary of scan results
func (ss *SecurityScanner) calculateSummary(vulnerabilities []Vulnerability) ScanSummary {
	summary := ScanSummary{
		TotalVulnerabilities: len(vulnerabilities),
	}

	riskScore := 0
	for _, vuln := range vulnerabilities {
		switch vuln.Level {
		case VulnerabilityLevelCritical:
			summary.CriticalCount++
			riskScore += 10
		case VulnerabilityLevelHigh:
			summary.HighCount++
			riskScore += 7
		case VulnerabilityLevelMedium:
			summary.MediumCount++
			riskScore += 5
		case VulnerabilityLevelLow:
			summary.LowCount++
			riskScore += 3
		case VulnerabilityLevelInfo:
			summary.InfoCount++
			riskScore += 1
		}
	}

	// Normalize risk score to 0-100
	if len(vulnerabilities) > 0 {
		summary.RiskScore = riskScore / len(vulnerabilities)
	}

	return summary
}

// SaveReport saves a scan result to a file
func (ss *SecurityScanner) SaveReport(result *SecurityScanResult, outputPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}

// LoadReport loads a scan result from a file
func (ss *SecurityScanner) LoadReport(filePath string) (*SecurityScanResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var result SecurityScanResult
	err = json.Unmarshal(data, &result)
	return &result, err
}

// Global security scanner instance
var GlobalSecurityScanner = NewSecurityScanner()

// Convenience functions
func ScanDirectory(ctx context.Context, targetDir string) (*SecurityScanResult, error) {
	return GlobalSecurityScanner.ScanDirectory(ctx, targetDir)
}

func SaveSecurityReport(result *SecurityScanResult, outputPath string) error {
	return GlobalSecurityScanner.SaveReport(result, outputPath)
}

func LoadSecurityReport(filePath string) (*SecurityScanResult, error) {
	return GlobalSecurityScanner.LoadReport(filePath)
}
