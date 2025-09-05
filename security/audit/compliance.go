package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ComplianceStandard represents a compliance standard
type ComplianceStandard string

const (
	ComplianceStandardSOC2     ComplianceStandard = "soc2"
	ComplianceStandardISO27001 ComplianceStandard = "iso27001"
	ComplianceStandardGDPR     ComplianceStandard = "gdpr"
	ComplianceStandardHIPAA    ComplianceStandard = "hipaa"
	ComplianceStandardPCIDSS   ComplianceStandard = "pci_dss"
	ComplianceStandardNIST     ComplianceStandard = "nist"
)

// ComplianceControl represents a compliance control
type ComplianceControl struct {
	ID          string             `json:"id"`
	Standard    ComplianceStandard `json:"standard"`
	Category    string             `json:"category"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Requirement string             `json:"requirement"`
	Level       string             `json:"level"` // mandatory, recommended, optional
	References  []string           `json:"references,omitempty"`
}

// ComplianceCheck represents a compliance check result
type ComplianceCheck struct {
	ControlID   string    `json:"control_id"`
	Status      string    `json:"status"` // passed, failed, warning, not_applicable
	Message     string    `json:"message"`
	Evidence    []string  `json:"evidence,omitempty"`
	CheckedAt   time.Time `json:"checked_at"`
	Remediation string    `json:"remediation,omitempty"`
}

// ComplianceReport represents a compliance assessment report
type ComplianceReport struct {
	ReportID  string             `json:"report_id"`
	Standard  ComplianceStandard `json:"standard"`
	StartTime time.Time          `json:"start_time"`
	EndTime   time.Time          `json:"end_time"`
	Duration  time.Duration      `json:"duration"`
	Checks    []ComplianceCheck  `json:"checks"`
	Summary   ComplianceSummary  `json:"summary"`
	Status    string             `json:"status"`
	Error     string             `json:"error,omitempty"`
}

// ComplianceSummary provides a summary of compliance results
type ComplianceSummary struct {
	TotalControls         int `json:"total_controls"`
	PassedControls        int `json:"passed_controls"`
	FailedControls        int `json:"failed_controls"`
	WarningControls       int `json:"warning_controls"`
	NotApplicableControls int `json:"not_applicable_controls"`
	ComplianceScore       int `json:"compliance_score"` // 0-100
}

// ComplianceAuditor performs compliance assessments
type ComplianceAuditor struct {
	controls map[ComplianceStandard][]ComplianceControl
}

// NewComplianceAuditor creates a new compliance auditor
func NewComplianceAuditor() *ComplianceAuditor {
	auditor := &ComplianceAuditor{
		controls: make(map[ComplianceStandard][]ComplianceControl),
	}

	// Initialize compliance controls
	auditor.initializeControls()

	return auditor
}

// initializeControls initializes compliance controls for various standards
func (ca *ComplianceAuditor) initializeControls() {
	// SOC 2 Controls
	ca.controls[ComplianceStandardSOC2] = []ComplianceControl{
		{
			ID:          "soc2_cc1_1",
			Standard:    ComplianceStandardSOC2,
			Category:    "Control Environment",
			Title:       "Control Environment",
			Description: "The entity demonstrates a commitment to integrity and ethical values",
			Requirement: "Establish and maintain a control environment that supports the achievement of objectives",
			Level:       "mandatory",
		},
		{
			ID:          "soc2_cc2_1",
			Standard:    ComplianceStandardSOC2,
			Category:    "Communication and Information",
			Title:       "Communication and Information",
			Description: "The entity obtains or generates and uses relevant, quality information",
			Requirement: "Obtain and communicate information necessary for internal control",
			Level:       "mandatory",
		},
		{
			ID:          "soc2_cc3_1",
			Standard:    ComplianceStandardSOC2,
			Category:    "Risk Assessment",
			Title:       "Risk Assessment",
			Description: "The entity specifies suitable objectives and identifies and analyzes risks",
			Requirement: "Identify and analyze risks to achievement of objectives",
			Level:       "mandatory",
		},
		{
			ID:          "soc2_cc4_1",
			Standard:    ComplianceStandardSOC2,
			Category:    "Monitoring Activities",
			Title:       "Monitoring Activities",
			Description: "The entity selects, develops, and performs ongoing and/or separate evaluations",
			Requirement: "Select, develop, and perform ongoing and separate evaluations",
			Level:       "mandatory",
		},
		{
			ID:          "soc2_cc5_1",
			Standard:    ComplianceStandardSOC2,
			Category:    "Control Activities",
			Title:       "Control Activities",
			Description: "The entity selects and develops control activities that contribute to the mitigation of risks",
			Requirement: "Select and develop control activities that mitigate risks",
			Level:       "mandatory",
		},
	}

	// GDPR Controls
	ca.controls[ComplianceStandardGDPR] = []ComplianceControl{
		{
			ID:          "gdpr_art5_1",
			Standard:    ComplianceStandardGDPR,
			Category:    "Data Processing Principles",
			Title:       "Lawfulness, Fairness, and Transparency",
			Description: "Personal data shall be processed lawfully, fairly, and in a transparent manner",
			Requirement: "Ensure data processing is lawful, fair, and transparent",
			Level:       "mandatory",
		},
		{
			ID:          "gdpr_art5_2",
			Standard:    ComplianceStandardGDPR,
			Category:    "Data Processing Principles",
			Title:       "Purpose Limitation",
			Description: "Personal data shall be collected for specified, explicit, and legitimate purposes",
			Requirement: "Limit data collection to specified, explicit, and legitimate purposes",
			Level:       "mandatory",
		},
		{
			ID:          "gdpr_art5_3",
			Standard:    ComplianceStandardGDPR,
			Category:    "Data Processing Principles",
			Title:       "Data Minimization",
			Description: "Personal data shall be adequate, relevant, and limited to what is necessary",
			Requirement: "Collect only adequate, relevant, and necessary data",
			Level:       "mandatory",
		},
		{
			ID:          "gdpr_art5_4",
			Standard:    ComplianceStandardGDPR,
			Category:    "Data Processing Principles",
			Title:       "Accuracy",
			Description: "Personal data shall be accurate and, where necessary, kept up to date",
			Requirement: "Ensure data accuracy and keep data up to date",
			Level:       "mandatory",
		},
		{
			ID:          "gdpr_art5_5",
			Standard:    ComplianceStandardGDPR,
			Category:    "Data Processing Principles",
			Title:       "Storage Limitation",
			Description: "Personal data shall be kept in a form which permits identification for no longer than necessary",
			Requirement: "Limit data storage duration to what is necessary",
			Level:       "mandatory",
		},
		{
			ID:          "gdpr_art5_6",
			Standard:    ComplianceStandardGDPR,
			Category:    "Data Processing Principles",
			Title:       "Integrity and Confidentiality",
			Description: "Personal data shall be processed in a manner that ensures appropriate security",
			Requirement: "Ensure appropriate security of personal data",
			Level:       "mandatory",
		},
	}

	// ISO 27001 Controls
	ca.controls[ComplianceStandardISO27001] = []ComplianceControl{
		{
			ID:          "iso27001_a5_1_1",
			Standard:    ComplianceStandardISO27001,
			Category:    "Information Security Policies",
			Title:       "Information Security Policies",
			Description: "Management direction and support for information security",
			Requirement: "Establish and maintain information security policies",
			Level:       "mandatory",
		},
		{
			ID:          "iso27001_a6_1_1",
			Standard:    ComplianceStandardISO27001,
			Category:    "Organization of Information Security",
			Title:       "Information Security Roles and Responsibilities",
			Description: "All information security responsibilities shall be defined and allocated",
			Requirement: "Define and allocate information security responsibilities",
			Level:       "mandatory",
		},
		{
			ID:          "iso27001_a8_1_1",
			Standard:    ComplianceStandardISO27001,
			Category:    "Asset Management",
			Title:       "Inventory of Assets",
			Description: "Assets associated with information and information processing facilities shall be identified",
			Requirement: "Maintain inventory of information assets",
			Level:       "mandatory",
		},
		{
			ID:          "iso27001_a9_1_1",
			Standard:    ComplianceStandardISO27001,
			Category:    "Access Control",
			Title:       "Access Control Policy",
			Description: "Access to information and information processing facilities shall be controlled",
			Requirement: "Implement access control policy",
			Level:       "mandatory",
		},
		{
			ID:          "iso27001_a10_1_1",
			Standard:    ComplianceStandardISO27001,
			Category:    "Cryptography",
			Title:       "Policy on the Use of Cryptographic Controls",
			Description: "Cryptographic controls shall be used to protect information",
			Requirement: "Implement cryptographic controls",
			Level:       "mandatory",
		},
	}
}

// AssessCompliance performs a compliance assessment for a specific standard
func (ca *ComplianceAuditor) AssessCompliance(ctx context.Context, standard ComplianceStandard, targetDir string) (*ComplianceReport, error) {
	reportID := fmt.Sprintf("compliance_%s_%d", standard, time.Now().UnixNano())
	startTime := time.Now()

	report := &ComplianceReport{
		ReportID:  reportID,
		Standard:  standard,
		StartTime: startTime,
		Status:    "running",
	}

	// Get controls for the standard
	controls, exists := ca.controls[standard]
	if !exists {
		report.Status = "failed"
		report.Error = fmt.Sprintf("unsupported compliance standard: %s", standard)
		report.EndTime = time.Now()
		report.Duration = report.EndTime.Sub(report.StartTime)
		return report, fmt.Errorf("unsupported compliance standard: %s", standard)
	}

	// Perform compliance checks
	var checks []ComplianceCheck
	for _, control := range controls {
		check := ca.performComplianceCheck(ctx, control, targetDir)
		checks = append(checks, check)
	}

	report.Checks = checks
	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)
	report.Status = "completed"
	report.Summary = ca.calculateComplianceSummary(checks)

	return report, nil
}

// performComplianceCheck performs a compliance check for a specific control
func (ca *ComplianceAuditor) performComplianceCheck(ctx context.Context, control ComplianceControl, targetDir string) ComplianceCheck {
	check := ComplianceCheck{
		ControlID: control.ID,
		CheckedAt: time.Now(),
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		check.Status = "failed"
		check.Message = "compliance check cancelled"
		return check
	default:
	}

	// Perform specific checks based on control type
	switch control.Standard {
	case ComplianceStandardSOC2:
		check = ca.checkSOC2Control(control, targetDir)
	case ComplianceStandardGDPR:
		check = ca.checkGDPRControl(control, targetDir)
	case ComplianceStandardISO27001:
		check = ca.checkISO27001Control(control, targetDir)
	default:
		check.Status = "not_applicable"
		check.Message = "control not implemented for this standard"
	}

	return check
}

// checkSOC2Control checks SOC 2 specific controls
func (ca *ComplianceAuditor) checkSOC2Control(control ComplianceControl, targetDir string) ComplianceCheck {
	check := ComplianceCheck{
		ControlID: control.ID,
		CheckedAt: time.Now(),
	}

	switch control.ID {
	case "soc2_cc1_1":
		// Check for control environment
		if ca.hasSecurityPolicies(targetDir) {
			check.Status = "passed"
			check.Message = "Security policies found"
			check.Evidence = []string{"Security policies directory exists"}
		} else {
			check.Status = "failed"
			check.Message = "No security policies found"
			check.Remediation = "Create security policies and procedures"
		}
	case "soc2_cc2_1":
		// Check for communication and information
		if ca.hasAuditLogging(targetDir) {
			check.Status = "passed"
			check.Message = "Audit logging implemented"
			check.Evidence = []string{"Audit logging system found"}
		} else {
			check.Status = "failed"
			check.Message = "No audit logging found"
			check.Remediation = "Implement comprehensive audit logging"
		}
	case "soc2_cc3_1":
		// Check for risk assessment
		if ca.hasRiskAssessment(targetDir) {
			check.Status = "passed"
			check.Message = "Risk assessment implemented"
			check.Evidence = []string{"Risk assessment documentation found"}
		} else {
			check.Status = "warning"
			check.Message = "Limited risk assessment found"
			check.Remediation = "Implement comprehensive risk assessment process"
		}
	case "soc2_cc4_1":
		// Check for monitoring activities
		if ca.hasMonitoring(targetDir) {
			check.Status = "passed"
			check.Message = "Monitoring activities implemented"
			check.Evidence = []string{"Monitoring system found"}
		} else {
			check.Status = "failed"
			check.Message = "No monitoring activities found"
			check.Remediation = "Implement monitoring and alerting system"
		}
	case "soc2_cc5_1":
		// Check for control activities
		if ca.hasAccessControls(targetDir) {
			check.Status = "passed"
			check.Message = "Access controls implemented"
			check.Evidence = []string{"Access control system found"}
		} else {
			check.Status = "failed"
			check.Message = "No access controls found"
			check.Remediation = "Implement access control system"
		}
	default:
		check.Status = "not_applicable"
		check.Message = "Control not implemented"
	}

	return check
}

// checkGDPRControl checks GDPR specific controls
func (ca *ComplianceAuditor) checkGDPRControl(control ComplianceControl, targetDir string) ComplianceCheck {
	check := ComplianceCheck{
		ControlID: control.ID,
		CheckedAt: time.Now(),
	}

	switch control.ID {
	case "gdpr_art5_1", "gdpr_art5_2", "gdpr_art5_3", "gdpr_art5_4", "gdpr_art5_5", "gdpr_art5_6":
		// Check for data protection measures
		if ca.hasDataProtection(targetDir) {
			check.Status = "passed"
			check.Message = "Data protection measures implemented"
			check.Evidence = []string{"Data protection system found"}
		} else {
			check.Status = "failed"
			check.Message = "No data protection measures found"
			check.Remediation = "Implement data protection and privacy controls"
		}
	default:
		check.Status = "not_applicable"
		check.Message = "Control not implemented"
	}

	return check
}

// checkISO27001Control checks ISO 27001 specific controls
func (ca *ComplianceAuditor) checkISO27001Control(control ComplianceControl, targetDir string) ComplianceCheck {
	check := ComplianceCheck{
		ControlID: control.ID,
		CheckedAt: time.Now(),
	}

	switch control.ID {
	case "iso27001_a5_1_1":
		// Check for information security policies
		if ca.hasSecurityPolicies(targetDir) {
			check.Status = "passed"
			check.Message = "Information security policies found"
			check.Evidence = []string{"Security policies directory exists"}
		} else {
			check.Status = "failed"
			check.Message = "No information security policies found"
			check.Remediation = "Create information security policies"
		}
	case "iso27001_a6_1_1":
		// Check for security roles and responsibilities
		if ca.hasSecurityRoles(targetDir) {
			check.Status = "passed"
			check.Message = "Security roles and responsibilities defined"
			check.Evidence = []string{"Security roles documentation found"}
		} else {
			check.Status = "warning"
			check.Message = "Limited security roles documentation"
			check.Remediation = "Define security roles and responsibilities"
		}
	case "iso27001_a8_1_1":
		// Check for asset inventory
		if ca.hasAssetInventory(targetDir) {
			check.Status = "passed"
			check.Message = "Asset inventory implemented"
			check.Evidence = []string{"Asset inventory system found"}
		} else {
			check.Status = "failed"
			check.Message = "No asset inventory found"
			check.Remediation = "Implement asset inventory system"
		}
	case "iso27001_a9_1_1":
		// Check for access control policy
		if ca.hasAccessControls(targetDir) {
			check.Status = "passed"
			check.Message = "Access control policy implemented"
			check.Evidence = []string{"Access control system found"}
		} else {
			check.Status = "failed"
			check.Message = "No access control policy found"
			check.Remediation = "Implement access control policy"
		}
	case "iso27001_a10_1_1":
		// Check for cryptographic controls
		if ca.hasCryptographicControls(targetDir) {
			check.Status = "passed"
			check.Message = "Cryptographic controls implemented"
			check.Evidence = []string{"Cryptographic system found"}
		} else {
			check.Status = "failed"
			check.Message = "No cryptographic controls found"
			check.Remediation = "Implement cryptographic controls"
		}
	default:
		check.Status = "not_applicable"
		check.Message = "Control not implemented"
	}

	return check
}

// Helper functions to check for specific security features
func (ca *ComplianceAuditor) hasSecurityPolicies(targetDir string) bool {
	policyDirs := []string{"security/policies", "docs/security", "policies"}
	for _, dir := range policyDirs {
		if _, err := os.Stat(filepath.Join(targetDir, dir)); err == nil {
			return true
		}
	}
	return false
}

func (ca *ComplianceAuditor) hasAuditLogging(targetDir string) bool {
	auditFiles := []string{"internal/audit", "internal/logging", "audit"}
	for _, file := range auditFiles {
		if _, err := os.Stat(filepath.Join(targetDir, file)); err == nil {
			return true
		}
	}
	return false
}

func (ca *ComplianceAuditor) hasRiskAssessment(targetDir string) bool {
	riskFiles := []string{"security/risk", "docs/risk", "risk-assessment"}
	for _, file := range riskFiles {
		if _, err := os.Stat(filepath.Join(targetDir, file)); err == nil {
			return true
		}
	}
	return false
}

func (ca *ComplianceAuditor) hasMonitoring(targetDir string) bool {
	monitoringFiles := []string{"internal/metrics", "internal/monitoring", "monitoring"}
	for _, file := range monitoringFiles {
		if _, err := os.Stat(filepath.Join(targetDir, file)); err == nil {
			return true
		}
	}
	return false
}

func (ca *ComplianceAuditor) hasAccessControls(targetDir string) bool {
	accessFiles := []string{"internal/auth", "internal/rbac", "internal/access"}
	for _, file := range accessFiles {
		if _, err := os.Stat(filepath.Join(targetDir, file)); err == nil {
			return true
		}
	}
	return false
}

func (ca *ComplianceAuditor) hasDataProtection(targetDir string) bool {
	protectionFiles := []string{"internal/privacy", "internal/encryption", "internal/data-protection"}
	for _, file := range protectionFiles {
		if _, err := os.Stat(filepath.Join(targetDir, file)); err == nil {
			return true
		}
	}
	return false
}

func (ca *ComplianceAuditor) hasSecurityRoles(targetDir string) bool {
	roleFiles := []string{"docs/security-roles", "security/roles", "internal/roles"}
	for _, file := range roleFiles {
		if _, err := os.Stat(filepath.Join(targetDir, file)); err == nil {
			return true
		}
	}
	return false
}

func (ca *ComplianceAuditor) hasAssetInventory(targetDir string) bool {
	assetFiles := []string{"internal/assets", "docs/assets", "asset-inventory"}
	for _, file := range assetFiles {
		if _, err := os.Stat(filepath.Join(targetDir, file)); err == nil {
			return true
		}
	}
	return false
}

func (ca *ComplianceAuditor) hasCryptographicControls(targetDir string) bool {
	cryptoFiles := []string{"internal/crypto", "internal/encryption", "crypto"}
	for _, file := range cryptoFiles {
		if _, err := os.Stat(filepath.Join(targetDir, file)); err == nil {
			return true
		}
	}
	return false
}

// calculateComplianceSummary calculates a summary of compliance results
func (ca *ComplianceAuditor) calculateComplianceSummary(checks []ComplianceCheck) ComplianceSummary {
	summary := ComplianceSummary{
		TotalControls: len(checks),
	}

	score := 0
	for _, check := range checks {
		switch check.Status {
		case "passed":
			summary.PassedControls++
			score += 100
		case "failed":
			summary.FailedControls++
			score += 0
		case "warning":
			summary.WarningControls++
			score += 50
		case "not_applicable":
			summary.NotApplicableControls++
			score += 75 // Partial credit for N/A
		}
	}

	// Calculate compliance score
	if len(checks) > 0 {
		summary.ComplianceScore = score / len(checks)
	}

	return summary
}

// SaveComplianceReport saves a compliance report to a file
func (ca *ComplianceAuditor) SaveComplianceReport(report *ComplianceReport, outputPath string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}

// LoadComplianceReport loads a compliance report from a file
func (ca *ComplianceAuditor) LoadComplianceReport(filePath string) (*ComplianceReport, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var report ComplianceReport
	err = json.Unmarshal(data, &report)
	return &report, err
}

// Global compliance auditor instance
var GlobalComplianceAuditor = NewComplianceAuditor()

// Convenience functions
func AssessCompliance(ctx context.Context, standard ComplianceStandard, targetDir string) (*ComplianceReport, error) {
	return GlobalComplianceAuditor.AssessCompliance(ctx, standard, targetDir)
}

func SaveComplianceReport(report *ComplianceReport, outputPath string) error {
	return GlobalComplianceAuditor.SaveComplianceReport(report, outputPath)
}

func LoadComplianceReport(filePath string) (*ComplianceReport, error) {
	return GlobalComplianceAuditor.LoadComplianceReport(filePath)
}
