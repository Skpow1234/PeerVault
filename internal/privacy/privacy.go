package privacy

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DataClassification represents the classification level of data
type DataClassification string

const (
	DataClassificationPublic       DataClassification = "public"
	DataClassificationInternal     DataClassification = "internal"
	DataClassificationConfidential DataClassification = "confidential"
	DataClassificationRestricted   DataClassification = "restricted"
	DataClassificationTopSecret    DataClassification = "top_secret"
)

// DataType represents the type of data
type DataType string

const (
	DataTypePersonal   DataType = "personal"
	DataTypeFinancial  DataType = "financial"
	DataTypeHealth     DataType = "health"
	DataTypeBiometric  DataType = "biometric"
	DataTypeLocation   DataType = "location"
	DataTypeBehavioral DataType = "behavioral"
	DataTypeTechnical  DataType = "technical"
	DataTypeBusiness   DataType = "business"
)

// PrivacyLaw represents privacy laws and regulations
type PrivacyLaw string

const (
	PrivacyLawGDPR   PrivacyLaw = "gdpr"
	PrivacyLawCCPA   PrivacyLaw = "ccpa"
	PrivacyLawHIPAA  PrivacyLaw = "hipaa"
	PrivacyLawPIPEDA PrivacyLaw = "pipeda"
	PrivacyLawLGPD   PrivacyLaw = "lgpd"
	PrivacyLawPDPA   PrivacyLaw = "pdpa"
)

// DataSubject represents a data subject (person)
type DataSubject struct {
	ID         string                 `json:"id"`
	Identifier string                 `json:"identifier"`
	Type       string                 `json:"type"` // individual, organization
	Attributes map[string]interface{} `json:"attributes"`
	Consent    map[string]Consent     `json:"consent"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// Consent represents consent for data processing
type Consent struct {
	Purpose     string    `json:"purpose"`
	LawfulBasis string    `json:"lawful_basis"`
	Given       bool      `json:"given"`
	Withdrawn   bool      `json:"withdrawn"`
	GivenAt     time.Time `json:"given_at,omitempty"`
	WithdrawnAt time.Time `json:"withdrawn_at,omitempty"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	Evidence    string    `json:"evidence,omitempty"`
}

// DataAsset represents a data asset
type DataAsset struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Classification  DataClassification     `json:"classification"`
	Type            DataType               `json:"type"`
	Location        string                 `json:"location"`
	Owner           string                 `json:"owner"`
	Custodian       string                 `json:"custodian"`
	RetentionPeriod time.Duration          `json:"retention_period"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Tags            []string               `json:"tags,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// DataProcessingActivity represents a data processing activity
type DataProcessingActivity struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Purpose          string                 `json:"purpose"`
	LawfulBasis      string                 `json:"lawful_basis"`
	DataSubjects     []string               `json:"data_subjects"`
	DataCategories   []DataType             `json:"data_categories"`
	Recipients       []string               `json:"recipients"`
	Transfers        []string               `json:"transfers"`
	RetentionPeriod  time.Duration          `json:"retention_period"`
	SecurityMeasures []string               `json:"security_measures"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// PrivacyImpactAssessment represents a privacy impact assessment
type PrivacyImpactAssessment struct {
	ID          string                 `json:"id"`
	ActivityID  string                 `json:"activity_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	RiskLevel   string                 `json:"risk_level"` // low, medium, high, critical
	Risks       []PrivacyRisk          `json:"risks"`
	Mitigations []PrivacyMitigation    `json:"mitigations"`
	Status      string                 `json:"status"` // draft, review, approved, rejected
	Assessor    string                 `json:"assessor"`
	Reviewer    string                 `json:"reviewer"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ApprovedAt  time.Time              `json:"approved_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PrivacyRisk represents a privacy risk
type PrivacyRisk struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Likelihood  string `json:"likelihood"` // low, medium, high
	Impact      string `json:"impact"`     // low, medium, high
	RiskLevel   string `json:"risk_level"` // low, medium, high, critical
}

// PrivacyMitigation represents a privacy mitigation measure
type PrivacyMitigation struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Type        string `json:"type"`   // technical, organizational, legal
	Status      string `json:"status"` // planned, implemented, verified
}

// DataRetentionPolicy represents a data retention policy
type DataRetentionPolicy struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	DataTypes       []DataType             `json:"data_types"`
	RetentionPeriod time.Duration          `json:"retention_period"`
	LegalBasis      string                 `json:"legal_basis"`
	DisposalMethod  string                 `json:"disposal_method"`
	Owner           string                 `json:"owner"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PrivacyManager manages privacy and data protection
type PrivacyManager struct {
	subjects    map[string]*DataSubject
	assets      map[string]*DataAsset
	activities  map[string]*DataProcessingActivity
	assessments map[string]*PrivacyImpactAssessment
	policies    map[string]*DataRetentionPolicy
	mu          sync.RWMutex
}

// NewPrivacyManager creates a new privacy manager
func NewPrivacyManager() *PrivacyManager {
	pm := &PrivacyManager{
		subjects:    make(map[string]*DataSubject),
		assets:      make(map[string]*DataAsset),
		activities:  make(map[string]*DataProcessingActivity),
		assessments: make(map[string]*PrivacyImpactAssessment),
		policies:    make(map[string]*DataRetentionPolicy),
	}

	// Initialize default policies
	pm.initializeDefaultPolicies()

	return pm
}

// initializeDefaultPolicies initializes default data retention policies
func (pm *PrivacyManager) initializeDefaultPolicies() {
	policies := []DataRetentionPolicy{
		{
			ID:              "personal_data_retention",
			Name:            "Personal Data Retention",
			Description:     "Retention policy for personal data",
			DataTypes:       []DataType{DataTypePersonal, DataTypeHealth, DataTypeBiometric},
			RetentionPeriod: 7 * 365 * 24 * time.Hour, // 7 years
			LegalBasis:      "Legal obligation",
			DisposalMethod:  "Secure deletion",
			Owner:           "Privacy Officer",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "financial_data_retention",
			Name:            "Financial Data Retention",
			Description:     "Retention policy for financial data",
			DataTypes:       []DataType{DataTypeFinancial},
			RetentionPeriod: 10 * 365 * 24 * time.Hour, // 10 years
			LegalBasis:      "Legal obligation",
			DisposalMethod:  "Secure deletion",
			Owner:           "Finance Officer",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "technical_data_retention",
			Name:            "Technical Data Retention",
			Description:     "Retention policy for technical data",
			DataTypes:       []DataType{DataTypeTechnical},
			RetentionPeriod: 3 * 365 * 24 * time.Hour, // 3 years
			LegalBasis:      "Legitimate interest",
			DisposalMethod:  "Secure deletion",
			Owner:           "IT Officer",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for _, policy := range policies {
		pm.policies[policy.ID] = &policy
	}
}

// CreateDataSubject creates a new data subject
func (pm *PrivacyManager) CreateDataSubject(subject *DataSubject) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.subjects[subject.ID]; exists {
		return fmt.Errorf("data subject %s already exists", subject.ID)
	}

	subject.CreatedAt = time.Now()
	subject.UpdatedAt = time.Now()

	pm.subjects[subject.ID] = subject
	return nil
}

// GetDataSubject retrieves a data subject by ID
func (pm *PrivacyManager) GetDataSubject(subjectID string) (*DataSubject, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	subject, exists := pm.subjects[subjectID]
	return subject, exists
}

// UpdateDataSubject updates an existing data subject
func (pm *PrivacyManager) UpdateDataSubject(subject *DataSubject) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.subjects[subject.ID]; !exists {
		return fmt.Errorf("data subject %s not found", subject.ID)
	}

	subject.UpdatedAt = time.Now()
	pm.subjects[subject.ID] = subject
	return nil
}

// DeleteDataSubject deletes a data subject
func (pm *PrivacyManager) DeleteDataSubject(subjectID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.subjects[subjectID]; !exists {
		return fmt.Errorf("data subject %s not found", subjectID)
	}

	delete(pm.subjects, subjectID)
	return nil
}

// CreateDataAsset creates a new data asset
func (pm *PrivacyManager) CreateDataAsset(asset *DataAsset) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.assets[asset.ID]; exists {
		return fmt.Errorf("data asset %s already exists", asset.ID)
	}

	asset.CreatedAt = time.Now()
	asset.UpdatedAt = time.Now()

	pm.assets[asset.ID] = asset
	return nil
}

// GetDataAsset retrieves a data asset by ID
func (pm *PrivacyManager) GetDataAsset(assetID string) (*DataAsset, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	asset, exists := pm.assets[assetID]
	return asset, exists
}

// UpdateDataAsset updates an existing data asset
func (pm *PrivacyManager) UpdateDataAsset(asset *DataAsset) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.assets[asset.ID]; !exists {
		return fmt.Errorf("data asset %s not found", asset.ID)
	}

	asset.UpdatedAt = time.Now()
	pm.assets[asset.ID] = asset
	return nil
}

// DeleteDataAsset deletes a data asset
func (pm *PrivacyManager) DeleteDataAsset(assetID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.assets[assetID]; !exists {
		return fmt.Errorf("data asset %s not found", assetID)
	}

	delete(pm.assets, assetID)
	return nil
}

// CreateDataProcessingActivity creates a new data processing activity
func (pm *PrivacyManager) CreateDataProcessingActivity(activity *DataProcessingActivity) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.activities[activity.ID]; exists {
		return fmt.Errorf("data processing activity %s already exists", activity.ID)
	}

	activity.CreatedAt = time.Now()
	activity.UpdatedAt = time.Now()

	pm.activities[activity.ID] = activity
	return nil
}

// GetDataProcessingActivity retrieves a data processing activity by ID
func (pm *PrivacyManager) GetDataProcessingActivity(activityID string) (*DataProcessingActivity, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	activity, exists := pm.activities[activityID]
	return activity, exists
}

// UpdateDataProcessingActivity updates an existing data processing activity
func (pm *PrivacyManager) UpdateDataProcessingActivity(activity *DataProcessingActivity) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.activities[activity.ID]; !exists {
		return fmt.Errorf("data processing activity %s not found", activity.ID)
	}

	activity.UpdatedAt = time.Now()
	pm.activities[activity.ID] = activity
	return nil
}

// DeleteDataProcessingActivity deletes a data processing activity
func (pm *PrivacyManager) DeleteDataProcessingActivity(activityID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.activities[activityID]; !exists {
		return fmt.Errorf("data processing activity %s not found", activityID)
	}

	delete(pm.activities, activityID)
	return nil
}

// CreatePrivacyImpactAssessment creates a new privacy impact assessment
func (pm *PrivacyManager) CreatePrivacyImpactAssessment(assessment *PrivacyImpactAssessment) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.assessments[assessment.ID]; exists {
		return fmt.Errorf("privacy impact assessment %s already exists", assessment.ID)
	}

	assessment.CreatedAt = time.Now()
	assessment.UpdatedAt = time.Now()

	pm.assessments[assessment.ID] = assessment
	return nil
}

// GetPrivacyImpactAssessment retrieves a privacy impact assessment by ID
func (pm *PrivacyManager) GetPrivacyImpactAssessment(assessmentID string) (*PrivacyImpactAssessment, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	assessment, exists := pm.assessments[assessmentID]
	return assessment, exists
}

// UpdatePrivacyImpactAssessment updates an existing privacy impact assessment
func (pm *PrivacyManager) UpdatePrivacyImpactAssessment(assessment *PrivacyImpactAssessment) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.assessments[assessment.ID]; !exists {
		return fmt.Errorf("privacy impact assessment %s not found", assessment.ID)
	}

	assessment.UpdatedAt = time.Now()
	pm.assessments[assessment.ID] = assessment
	return nil
}

// DeletePrivacyImpactAssessment deletes a privacy impact assessment
func (pm *PrivacyManager) DeletePrivacyImpactAssessment(assessmentID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.assessments[assessmentID]; !exists {
		return fmt.Errorf("privacy impact assessment %s not found", assessmentID)
	}

	delete(pm.assessments, assessmentID)
	return nil
}

// CreateDataRetentionPolicy creates a new data retention policy
func (pm *PrivacyManager) CreateDataRetentionPolicy(policy *DataRetentionPolicy) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.policies[policy.ID]; exists {
		return fmt.Errorf("data retention policy %s already exists", policy.ID)
	}

	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	pm.policies[policy.ID] = policy
	return nil
}

// GetDataRetentionPolicy retrieves a data retention policy by ID
func (pm *PrivacyManager) GetDataRetentionPolicy(policyID string) (*DataRetentionPolicy, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	policy, exists := pm.policies[policyID]
	return policy, exists
}

// UpdateDataRetentionPolicy updates an existing data retention policy
func (pm *PrivacyManager) UpdateDataRetentionPolicy(policy *DataRetentionPolicy) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.policies[policy.ID]; !exists {
		return fmt.Errorf("data retention policy %s not found", policy.ID)
	}

	policy.UpdatedAt = time.Now()
	pm.policies[policy.ID] = policy
	return nil
}

// DeleteDataRetentionPolicy deletes a data retention policy
func (pm *PrivacyManager) DeleteDataRetentionPolicy(policyID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.policies[policyID]; !exists {
		return fmt.Errorf("data retention policy %s not found", policyID)
	}

	delete(pm.policies, policyID)
	return nil
}

// AssessPrivacyRisk assesses privacy risk for a data processing activity
func (pm *PrivacyManager) AssessPrivacyRisk(ctx context.Context, activityID string) (*PrivacyImpactAssessment, error) {
	activity, exists := pm.GetDataProcessingActivity(activityID)
	if !exists {
		return nil, fmt.Errorf("data processing activity %s not found", activityID)
	}

	assessment := &PrivacyImpactAssessment{
		ID:          fmt.Sprintf("pia_%s_%d", activityID, time.Now().UnixNano()),
		ActivityID:  activityID,
		Title:       fmt.Sprintf("Privacy Impact Assessment for %s", activity.Name),
		Description: fmt.Sprintf("Privacy impact assessment for data processing activity: %s", activity.Description),
		Status:      "draft",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Assess risks based on data types and processing activities
	risks := pm.identifyPrivacyRisks(activity)
	assessment.Risks = risks

	// Determine overall risk level
	assessment.RiskLevel = pm.calculateOverallRiskLevel(risks)

	// Generate mitigations
	mitigations := pm.generateMitigations(risks)
	assessment.Mitigations = mitigations

	return assessment, nil
}

// identifyPrivacyRisks identifies privacy risks for a data processing activity
func (pm *PrivacyManager) identifyPrivacyRisks(activity *DataProcessingActivity) []PrivacyRisk {
	var risks []PrivacyRisk

	// Check for high-risk data types
	for _, dataType := range activity.DataCategories {
		switch dataType {
		case DataTypeHealth, DataTypeBiometric:
			risks = append(risks, PrivacyRisk{
				ID:          fmt.Sprintf("risk_%s_%d", dataType, time.Now().UnixNano()),
				Description: fmt.Sprintf("Processing of sensitive %s data", dataType),
				Likelihood:  "medium",
				Impact:      "high",
				RiskLevel:   "high",
			})
		case DataTypePersonal:
			risks = append(risks, PrivacyRisk{
				ID:          fmt.Sprintf("risk_%s_%d", dataType, time.Now().UnixNano()),
				Description: "Processing of personal data",
				Likelihood:  "high",
				Impact:      "medium",
				RiskLevel:   "medium",
			})
		}
	}

	// Check for international transfers
	if len(activity.Transfers) > 0 {
		risks = append(risks, PrivacyRisk{
			ID:          fmt.Sprintf("risk_transfer_%d", time.Now().UnixNano()),
			Description: "International data transfers",
			Likelihood:  "medium",
			Impact:      "medium",
			RiskLevel:   "medium",
		})
	}

	return risks
}

// calculateOverallRiskLevel calculates the overall risk level
func (pm *PrivacyManager) calculateOverallRiskLevel(risks []PrivacyRisk) string {
	if len(risks) == 0 {
		return "low"
	}

	hasHigh := false
	hasMedium := false

	for _, risk := range risks {
		switch risk.RiskLevel {
		case "critical":
			return "critical"
		case "high":
			hasHigh = true
		case "medium":
			hasMedium = true
		}
	}

	if hasHigh {
		return "high"
	}
	if hasMedium {
		return "medium"
	}

	return "low"
}

// generateMitigations generates mitigation measures for identified risks
func (pm *PrivacyManager) generateMitigations(risks []PrivacyRisk) []PrivacyMitigation {
	var mitigations []PrivacyMitigation

	for _, risk := range risks {
		switch risk.RiskLevel {
		case "high", "critical":
			mitigations = append(mitigations, PrivacyMitigation{
				ID:          fmt.Sprintf("mit_%s_%d", risk.ID, time.Now().UnixNano()),
				Description: "Implement strong encryption for data at rest and in transit",
				Type:        "technical",
				Status:      "planned",
			})
			mitigations = append(mitigations, PrivacyMitigation{
				ID:          fmt.Sprintf("mit_%s_%d", risk.ID, time.Now().UnixNano()+1),
				Description: "Implement access controls and authentication",
				Type:        "technical",
				Status:      "planned",
			})
			mitigations = append(mitigations, PrivacyMitigation{
				ID:          fmt.Sprintf("mit_%s_%d", risk.ID, time.Now().UnixNano()+2),
				Description: "Provide privacy training to staff",
				Type:        "organizational",
				Status:      "planned",
			})
		case "medium":
			mitigations = append(mitigations, PrivacyMitigation{
				ID:          fmt.Sprintf("mit_%s_%d", risk.ID, time.Now().UnixNano()),
				Description: "Implement data minimization practices",
				Type:        "organizational",
				Status:      "planned",
			})
		}
	}

	return mitigations
}

// CheckDataRetention checks if data should be retained based on policies
func (pm *PrivacyManager) CheckDataRetention(ctx context.Context, dataType DataType, createdAt time.Time) (bool, *DataRetentionPolicy, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Find applicable retention policy
	for _, policy := range pm.policies {
		for _, policyType := range policy.DataTypes {
			if policyType == dataType {
				retentionEnd := createdAt.Add(policy.RetentionPeriod)
				shouldRetain := time.Now().Before(retentionEnd)
				return shouldRetain, policy, nil
			}
		}
	}

	// Default retention if no specific policy found
	defaultRetention := 7 * 365 * 24 * time.Hour // 7 years
	retentionEnd := createdAt.Add(defaultRetention)
	shouldRetain := time.Now().Before(retentionEnd)

	return shouldRetain, nil, nil
}

// ListDataSubjects returns all data subjects
func (pm *PrivacyManager) ListDataSubjects() []*DataSubject {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var subjects []*DataSubject
	for _, subject := range pm.subjects {
		subjects = append(subjects, subject)
	}

	return subjects
}

// ListDataAssets returns all data assets
func (pm *PrivacyManager) ListDataAssets() []*DataAsset {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var assets []*DataAsset
	for _, asset := range pm.assets {
		assets = append(assets, asset)
	}

	return assets
}

// ListDataProcessingActivities returns all data processing activities
func (pm *PrivacyManager) ListDataProcessingActivities() []*DataProcessingActivity {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var activities []*DataProcessingActivity
	for _, activity := range pm.activities {
		activities = append(activities, activity)
	}

	return activities
}

// ListPrivacyImpactAssessments returns all privacy impact assessments
func (pm *PrivacyManager) ListPrivacyImpactAssessments() []*PrivacyImpactAssessment {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var assessments []*PrivacyImpactAssessment
	for _, assessment := range pm.assessments {
		assessments = append(assessments, assessment)
	}

	return assessments
}

// ListDataRetentionPolicies returns all data retention policies
func (pm *PrivacyManager) ListDataRetentionPolicies() []*DataRetentionPolicy {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var policies []*DataRetentionPolicy
	for _, policy := range pm.policies {
		policies = append(policies, policy)
	}

	return policies
}

// Global privacy manager instance
var GlobalPrivacyManager = NewPrivacyManager()

// Convenience functions
func CreateDataSubject(subject *DataSubject) error {
	return GlobalPrivacyManager.CreateDataSubject(subject)
}

func GetDataSubject(subjectID string) (*DataSubject, bool) {
	return GlobalPrivacyManager.GetDataSubject(subjectID)
}

func CreateDataAsset(asset *DataAsset) error {
	return GlobalPrivacyManager.CreateDataAsset(asset)
}

func GetDataAsset(assetID string) (*DataAsset, bool) {
	return GlobalPrivacyManager.GetDataAsset(assetID)
}

func CreateDataProcessingActivity(activity *DataProcessingActivity) error {
	return GlobalPrivacyManager.CreateDataProcessingActivity(activity)
}

func GetDataProcessingActivity(activityID string) (*DataProcessingActivity, bool) {
	return GlobalPrivacyManager.GetDataProcessingActivity(activityID)
}

func AssessPrivacyRisk(ctx context.Context, activityID string) (*PrivacyImpactAssessment, error) {
	return GlobalPrivacyManager.AssessPrivacyRisk(ctx, activityID)
}

func CheckDataRetention(ctx context.Context, dataType DataType, createdAt time.Time) (bool, *DataRetentionPolicy, error) {
	return GlobalPrivacyManager.CheckDataRetention(ctx, dataType, createdAt)
}
