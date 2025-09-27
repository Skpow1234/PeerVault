package privacy

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPrivacyManager(t *testing.T) {
	pm := NewPrivacyManager()
	
	assert.NotNil(t, pm)
	assert.NotNil(t, pm.subjects)
	assert.NotNil(t, pm.assets)
	assert.NotNil(t, pm.activities)
	assert.NotNil(t, pm.assessments)
	assert.NotNil(t, pm.policies)
	
	// Check that default policies are initialized
	policies := pm.ListDataRetentionPolicies()
	assert.GreaterOrEqual(t, len(policies), 3) // At least 3 default policies
	
	// Verify default policies exist
	policyIDs := make(map[string]bool)
	for _, policy := range policies {
		policyIDs[policy.ID] = true
	}
	assert.True(t, policyIDs["personal_data_retention"])
	assert.True(t, policyIDs["financial_data_retention"])
	assert.True(t, policyIDs["technical_data_retention"])
}

func TestPrivacyManager_DataSubjectOperations(t *testing.T) {
	pm := NewPrivacyManager()
	
	subject := &DataSubject{
		ID:         "subject1",
		Identifier: "user123",
		Type:       "individual",
		Attributes: map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		},
		Consent: map[string]Consent{
			"marketing": {
				Purpose:     "Marketing communications",
				LawfulBasis: "Consent",
				Given:       true,
				GivenAt:     time.Now(),
			},
		},
	}
	
	// Test creating data subject
	err := pm.CreateDataSubject(subject)
	assert.NoError(t, err)
	
	// Test retrieving data subject
	retrievedSubject, exists := pm.GetDataSubject("subject1")
	assert.True(t, exists)
	assert.Equal(t, "subject1", retrievedSubject.ID)
	assert.Equal(t, "user123", retrievedSubject.Identifier)
	assert.Equal(t, "individual", retrievedSubject.Type)
	assert.Equal(t, "John Doe", retrievedSubject.Attributes["name"])
	assert.NotZero(t, retrievedSubject.CreatedAt)
	assert.NotZero(t, retrievedSubject.UpdatedAt)
	
	// Test updating data subject
	subject.Attributes["name"] = "Jane Doe"
	err = pm.UpdateDataSubject(subject)
	assert.NoError(t, err)
	
	updatedSubject, exists := pm.GetDataSubject("subject1")
	assert.True(t, exists)
	assert.Equal(t, "Jane Doe", updatedSubject.Attributes["name"])
	
	// Test deleting data subject
	err = pm.DeleteDataSubject("subject1")
	assert.NoError(t, err)
	
	_, exists = pm.GetDataSubject("subject1")
	assert.False(t, exists)
	
	// Test creating duplicate data subject
	err = pm.CreateDataSubject(subject)
	assert.NoError(t, err)
	
	err = pm.CreateDataSubject(subject)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	
	// Test updating non-existent data subject
	nonExistentSubject := &DataSubject{ID: "nonexistent"}
	err = pm.UpdateDataSubject(nonExistentSubject)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	
	// Test deleting non-existent data subject
	err = pm.DeleteDataSubject("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPrivacyManager_DataAssetOperations(t *testing.T) {
	pm := NewPrivacyManager()
	
	asset := &DataAsset{
		ID:              "asset1",
		Name:            "Customer Database",
		Description:     "Main customer database",
		Classification:  DataClassificationConfidential,
		Type:            DataTypePersonal,
		Location:        "AWS US-East-1",
		Owner:           "Data Team",
		Custodian:       "IT Team",
		RetentionPeriod: 7 * 365 * 24 * time.Hour,
		Tags:            []string{"customer", "database"},
		Metadata: map[string]interface{}{
			"encryption": "AES-256",
			"backup":     "daily",
		},
	}
	
	// Test creating data asset
	err := pm.CreateDataAsset(asset)
	assert.NoError(t, err)
	
	// Test retrieving data asset
	retrievedAsset, exists := pm.GetDataAsset("asset1")
	assert.True(t, exists)
	assert.Equal(t, "asset1", retrievedAsset.ID)
	assert.Equal(t, "Customer Database", retrievedAsset.Name)
	assert.Equal(t, DataClassificationConfidential, retrievedAsset.Classification)
	assert.Equal(t, DataTypePersonal, retrievedAsset.Type)
	assert.Equal(t, "AWS US-East-1", retrievedAsset.Location)
	assert.NotZero(t, retrievedAsset.CreatedAt)
	assert.NotZero(t, retrievedAsset.UpdatedAt)
	
	// Test updating data asset
	asset.Description = "Updated customer database"
	err = pm.UpdateDataAsset(asset)
	assert.NoError(t, err)
	
	updatedAsset, exists := pm.GetDataAsset("asset1")
	assert.True(t, exists)
	assert.Equal(t, "Updated customer database", updatedAsset.Description)
	
	// Test deleting data asset
	err = pm.DeleteDataAsset("asset1")
	assert.NoError(t, err)
	
	_, exists = pm.GetDataAsset("asset1")
	assert.False(t, exists)
	
	// Test creating duplicate data asset
	err = pm.CreateDataAsset(asset)
	assert.NoError(t, err)
	
	err = pm.CreateDataAsset(asset)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	
	// Test updating non-existent data asset
	nonExistentAsset := &DataAsset{ID: "nonexistent"}
	err = pm.UpdateDataAsset(nonExistentAsset)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	
	// Test deleting non-existent data asset
	err = pm.DeleteDataAsset("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPrivacyManager_DataProcessingActivityOperations(t *testing.T) {
	pm := NewPrivacyManager()
	
	activity := &DataProcessingActivity{
		ID:               "activity1",
		Name:             "Customer Analytics",
		Description:      "Analyze customer behavior for product improvement",
		Purpose:          "Product development",
		LawfulBasis:      "Legitimate interest",
		DataSubjects:     []string{"subject1", "subject2"},
		DataCategories:   []DataType{DataTypePersonal, DataTypeBehavioral},
		Recipients:       []string{"Product Team", "Analytics Team"},
		Transfers:        []string{"US", "EU"},
		RetentionPeriod:  2 * 365 * 24 * time.Hour,
		SecurityMeasures: []string{"encryption", "access_controls", "audit_logs"},
		Metadata: map[string]interface{}{
			"frequency": "daily",
			"volume":    "1000 records",
		},
	}
	
	// Test creating data processing activity
	err := pm.CreateDataProcessingActivity(activity)
	assert.NoError(t, err)
	
	// Test retrieving data processing activity
	retrievedActivity, exists := pm.GetDataProcessingActivity("activity1")
	assert.True(t, exists)
	assert.Equal(t, "activity1", retrievedActivity.ID)
	assert.Equal(t, "Customer Analytics", retrievedActivity.Name)
	assert.Equal(t, "Product development", retrievedActivity.Purpose)
	assert.Equal(t, "Legitimate interest", retrievedActivity.LawfulBasis)
	assert.Equal(t, 2, len(retrievedActivity.DataSubjects))
	assert.Equal(t, 2, len(retrievedActivity.DataCategories))
	assert.NotZero(t, retrievedActivity.CreatedAt)
	assert.NotZero(t, retrievedActivity.UpdatedAt)
	
	// Test updating data processing activity
	activity.Description = "Updated customer analytics"
	err = pm.UpdateDataProcessingActivity(activity)
	assert.NoError(t, err)
	
	updatedActivity, exists := pm.GetDataProcessingActivity("activity1")
	assert.True(t, exists)
	assert.Equal(t, "Updated customer analytics", updatedActivity.Description)
	
	// Test deleting data processing activity
	err = pm.DeleteDataProcessingActivity("activity1")
	assert.NoError(t, err)
	
	_, exists = pm.GetDataProcessingActivity("activity1")
	assert.False(t, exists)
	
	// Test creating duplicate data processing activity
	err = pm.CreateDataProcessingActivity(activity)
	assert.NoError(t, err)
	
	err = pm.CreateDataProcessingActivity(activity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	
	// Test updating non-existent data processing activity
	nonExistentActivity := &DataProcessingActivity{ID: "nonexistent"}
	err = pm.UpdateDataProcessingActivity(nonExistentActivity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	
	// Test deleting non-existent data processing activity
	err = pm.DeleteDataProcessingActivity("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPrivacyManager_PrivacyImpactAssessmentOperations(t *testing.T) {
	pm := NewPrivacyManager()
	
	assessment := &PrivacyImpactAssessment{
		ID:         "assessment1",
		ActivityID: "activity1",
		Title:      "PIA for Customer Analytics",
		Description: "Privacy impact assessment for customer analytics activity",
		RiskLevel:  "medium",
		Risks: []PrivacyRisk{
			{
				ID:          "risk1",
				Description: "Data breach risk",
				Likelihood:  "low",
				Impact:      "high",
				RiskLevel:   "medium",
			},
		},
		Mitigations: []PrivacyMitigation{
			{
				ID:          "mit1",
				Description: "Implement encryption",
				Type:        "technical",
				Status:      "planned",
			},
		},
		Status:   "draft",
		Assessor: "Privacy Officer",
		Reviewer: "Legal Team",
		Metadata: map[string]interface{}{
			"version": "1.0",
		},
	}
	
	// Test creating privacy impact assessment
	err := pm.CreatePrivacyImpactAssessment(assessment)
	assert.NoError(t, err)
	
	// Test retrieving privacy impact assessment
	retrievedAssessment, exists := pm.GetPrivacyImpactAssessment("assessment1")
	assert.True(t, exists)
	assert.Equal(t, "assessment1", retrievedAssessment.ID)
	assert.Equal(t, "activity1", retrievedAssessment.ActivityID)
	assert.Equal(t, "PIA for Customer Analytics", retrievedAssessment.Title)
	assert.Equal(t, "medium", retrievedAssessment.RiskLevel)
	assert.Equal(t, 1, len(retrievedAssessment.Risks))
	assert.Equal(t, 1, len(retrievedAssessment.Mitigations))
	assert.Equal(t, "draft", retrievedAssessment.Status)
	assert.NotZero(t, retrievedAssessment.CreatedAt)
	assert.NotZero(t, retrievedAssessment.UpdatedAt)
	
	// Test updating privacy impact assessment
	assessment.Status = "approved"
	err = pm.UpdatePrivacyImpactAssessment(assessment)
	assert.NoError(t, err)
	
	updatedAssessment, exists := pm.GetPrivacyImpactAssessment("assessment1")
	assert.True(t, exists)
	assert.Equal(t, "approved", updatedAssessment.Status)
	
	// Test deleting privacy impact assessment
	err = pm.DeletePrivacyImpactAssessment("assessment1")
	assert.NoError(t, err)
	
	_, exists = pm.GetPrivacyImpactAssessment("assessment1")
	assert.False(t, exists)
	
	// Test creating duplicate privacy impact assessment
	err = pm.CreatePrivacyImpactAssessment(assessment)
	assert.NoError(t, err)
	
	err = pm.CreatePrivacyImpactAssessment(assessment)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	
	// Test updating non-existent privacy impact assessment
	nonExistentAssessment := &PrivacyImpactAssessment{ID: "nonexistent"}
	err = pm.UpdatePrivacyImpactAssessment(nonExistentAssessment)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	
	// Test deleting non-existent privacy impact assessment
	err = pm.DeletePrivacyImpactAssessment("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPrivacyManager_DataRetentionPolicyOperations(t *testing.T) {
	pm := NewPrivacyManager()
	
	policy := &DataRetentionPolicy{
		ID:              "custom_policy",
		Name:            "Custom Data Retention",
		Description:     "Custom retention policy for specific data",
		DataTypes:       []DataType{DataTypeLocation},
		RetentionPeriod: 1 * 365 * 24 * time.Hour, // 1 year
		LegalBasis:      "Consent",
		DisposalMethod:  "Secure deletion",
		Owner:           "Data Protection Officer",
		Metadata: map[string]interface{}{
			"category": "location_data",
		},
	}
	
	// Test creating data retention policy
	err := pm.CreateDataRetentionPolicy(policy)
	assert.NoError(t, err)
	
	// Test retrieving data retention policy
	retrievedPolicy, exists := pm.GetDataRetentionPolicy("custom_policy")
	assert.True(t, exists)
	assert.Equal(t, "custom_policy", retrievedPolicy.ID)
	assert.Equal(t, "Custom Data Retention", retrievedPolicy.Name)
	assert.Equal(t, DataTypeLocation, retrievedPolicy.DataTypes[0])
	assert.Equal(t, 1*365*24*time.Hour, retrievedPolicy.RetentionPeriod)
	assert.Equal(t, "Consent", retrievedPolicy.LegalBasis)
	assert.NotZero(t, retrievedPolicy.CreatedAt)
	assert.NotZero(t, retrievedPolicy.UpdatedAt)
	
	// Test updating data retention policy
	policy.Description = "Updated custom retention policy"
	err = pm.UpdateDataRetentionPolicy(policy)
	assert.NoError(t, err)
	
	updatedPolicy, exists := pm.GetDataRetentionPolicy("custom_policy")
	assert.True(t, exists)
	assert.Equal(t, "Updated custom retention policy", updatedPolicy.Description)
	
	// Test deleting data retention policy
	err = pm.DeleteDataRetentionPolicy("custom_policy")
	assert.NoError(t, err)
	
	_, exists = pm.GetDataRetentionPolicy("custom_policy")
	assert.False(t, exists)
	
	// Test creating duplicate data retention policy
	err = pm.CreateDataRetentionPolicy(policy)
	assert.NoError(t, err)
	
	err = pm.CreateDataRetentionPolicy(policy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	
	// Test updating non-existent data retention policy
	nonExistentPolicy := &DataRetentionPolicy{ID: "nonexistent"}
	err = pm.UpdateDataRetentionPolicy(nonExistentPolicy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	
	// Test deleting non-existent data retention policy
	err = pm.DeleteDataRetentionPolicy("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPrivacyManager_AssessPrivacyRisk(t *testing.T) {
	pm := NewPrivacyManager()
	ctx := context.Background()
	
	// Create a data processing activity
	activity := &DataProcessingActivity{
		ID:             "activity1",
		Name:           "Health Data Processing",
		Description:    "Process health data for research",
		DataCategories: []DataType{DataTypeHealth, DataTypePersonal},
		Transfers:      []string{"US"},
	}
	
	err := pm.CreateDataProcessingActivity(activity)
	assert.NoError(t, err)
	
	// Assess privacy risk
	assessment, err := pm.AssessPrivacyRisk(ctx, "activity1")
	assert.NoError(t, err)
	assert.NotNil(t, assessment)
	assert.Equal(t, "activity1", assessment.ActivityID)
	assert.Contains(t, assessment.Title, "Health Data Processing")
	assert.Equal(t, "draft", assessment.Status)
	assert.NotEmpty(t, assessment.Risks)
	assert.NotEmpty(t, assessment.Mitigations)
	
	// Verify risk assessment for health data
	hasHealthRisk := false
	for _, risk := range assessment.Risks {
		if risk.Description == "Processing of sensitive health data" {
			hasHealthRisk = true
			assert.Equal(t, "high", risk.RiskLevel)
			break
		}
	}
	assert.True(t, hasHealthRisk)
	
	// Verify risk assessment for personal data
	hasPersonalRisk := false
	for _, risk := range assessment.Risks {
		if risk.Description == "Processing of personal data" {
			hasPersonalRisk = true
			assert.Equal(t, "medium", risk.RiskLevel)
			break
		}
	}
	assert.True(t, hasPersonalRisk)
	
	// Verify risk assessment for international transfers
	hasTransferRisk := false
	for _, risk := range assessment.Risks {
		if risk.Description == "International data transfers" {
			hasTransferRisk = true
			assert.Equal(t, "medium", risk.RiskLevel)
			break
		}
	}
	assert.True(t, hasTransferRisk)
	
	// Test assessing risk for non-existent activity
	_, err = pm.AssessPrivacyRisk(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPrivacyManager_CheckDataRetention(t *testing.T) {
	pm := NewPrivacyManager()
	ctx := context.Background()
	
	// Test with personal data (should use default 7-year retention)
	createdAt := time.Now().Add(-2 * 365 * 24 * time.Hour) // 2 years ago
	shouldRetain, policy, err := pm.CheckDataRetention(ctx, DataTypePersonal, createdAt)
	assert.NoError(t, err)
	assert.True(t, shouldRetain) // Should still be within retention period
	assert.NotNil(t, policy)
	assert.Equal(t, "personal_data_retention", policy.ID)
	
	// Test with financial data (should use 10-year retention)
	createdAt = time.Now().Add(-5 * 365 * 24 * time.Hour) // 5 years ago
	shouldRetain, policy, err = pm.CheckDataRetention(ctx, DataTypeFinancial, createdAt)
	assert.NoError(t, err)
	assert.True(t, shouldRetain) // Should still be within retention period
	assert.NotNil(t, policy)
	assert.Equal(t, "financial_data_retention", policy.ID)
	
	// Test with technical data (should use 3-year retention)
	createdAt = time.Now().Add(-4 * 365 * 24 * time.Hour) // 4 years ago
	shouldRetain, policy, err = pm.CheckDataRetention(ctx, DataTypeTechnical, createdAt)
	assert.NoError(t, err)
	assert.False(t, shouldRetain) // Should be beyond retention period
	assert.NotNil(t, policy)
	assert.Equal(t, "technical_data_retention", policy.ID)
	
	// Test with data type that has no specific policy (should use default)
	createdAt = time.Now().Add(-1 * 365 * 24 * time.Hour) // 1 year ago
	shouldRetain, policy, err = pm.CheckDataRetention(ctx, DataTypeLocation, createdAt)
	assert.NoError(t, err)
	assert.True(t, shouldRetain) // Should still be within default retention period
	assert.Nil(t, policy) // No specific policy found
}

func TestPrivacyManager_ListOperations(t *testing.T) {
	pm := NewPrivacyManager()
	
	// Test empty lists
	subjects := pm.ListDataSubjects()
	assert.Equal(t, 0, len(subjects))
	
	assets := pm.ListDataAssets()
	assert.Equal(t, 0, len(assets))
	
	activities := pm.ListDataProcessingActivities()
	assert.Equal(t, 0, len(activities))
	
	assessments := pm.ListPrivacyImpactAssessments()
	assert.Equal(t, 0, len(assessments))
	
	policies := pm.ListDataRetentionPolicies()
	assert.GreaterOrEqual(t, len(policies), 3) // At least 3 default policies
	
	// Add some test data
	subject := &DataSubject{ID: "subject1", Identifier: "user1"}
	asset := &DataAsset{ID: "asset1", Name: "Test Asset", Classification: DataClassificationPublic, Type: DataTypeTechnical}
	activity := &DataProcessingActivity{ID: "activity1", Name: "Test Activity", DataCategories: []DataType{DataTypePersonal}}
	assessment := &PrivacyImpactAssessment{ID: "assessment1", ActivityID: "activity1", Title: "Test PIA", RiskLevel: "low", Status: "draft"}
	
	pm.CreateDataSubject(subject)
	pm.CreateDataAsset(asset)
	pm.CreateDataProcessingActivity(activity)
	pm.CreatePrivacyImpactAssessment(assessment)
	
	// Test populated lists
	subjects = pm.ListDataSubjects()
	assert.Equal(t, 1, len(subjects))
	assert.Equal(t, "subject1", subjects[0].ID)
	
	assets = pm.ListDataAssets()
	assert.Equal(t, 1, len(assets))
	assert.Equal(t, "asset1", assets[0].ID)
	
	activities = pm.ListDataProcessingActivities()
	assert.Equal(t, 1, len(activities))
	assert.Equal(t, "activity1", activities[0].ID)
	
	assessments = pm.ListPrivacyImpactAssessments()
	assert.Equal(t, 1, len(assessments))
	assert.Equal(t, "assessment1", assessments[0].ID)
}

func TestPrivacyManager_ConvenienceFunctions(t *testing.T) {
	ctx := context.Background()
	
	// Test convenience functions
	subject := &DataSubject{ID: "subject1", Identifier: "user1"}
	err := CreateDataSubject(subject)
	assert.NoError(t, err)
	
	retrievedSubject, exists := GetDataSubject("subject1")
	assert.True(t, exists)
	assert.Equal(t, "subject1", retrievedSubject.ID)
	
	asset := &DataAsset{ID: "asset1", Name: "Test Asset", Classification: DataClassificationPublic, Type: DataTypeTechnical}
	err = CreateDataAsset(asset)
	assert.NoError(t, err)
	
	retrievedAsset, exists := GetDataAsset("asset1")
	assert.True(t, exists)
	assert.Equal(t, "asset1", retrievedAsset.ID)
	
	activity := &DataProcessingActivity{ID: "activity1", Name: "Test Activity", DataCategories: []DataType{DataTypePersonal}}
	err = CreateDataProcessingActivity(activity)
	assert.NoError(t, err)
	
	retrievedActivity, exists := GetDataProcessingActivity("activity1")
	assert.True(t, exists)
	assert.Equal(t, "activity1", retrievedActivity.ID)
	
	// Test privacy risk assessment convenience function
	assessment, err := AssessPrivacyRisk(ctx, "activity1")
	assert.NoError(t, err)
	assert.NotNil(t, assessment)
	assert.Equal(t, "activity1", assessment.ActivityID)
	
	// Test data retention check convenience function
	createdAt := time.Now().Add(-1 * 365 * 24 * time.Hour)
	shouldRetain, policy, err := CheckDataRetention(ctx, DataTypePersonal, createdAt)
	assert.NoError(t, err)
	assert.True(t, shouldRetain)
	assert.NotNil(t, policy)
}

func TestPrivacyManager_DataClassificationConstants(t *testing.T) {
	// Test data classification constants
	assert.Equal(t, DataClassification("public"), DataClassificationPublic)
	assert.Equal(t, DataClassification("internal"), DataClassificationInternal)
	assert.Equal(t, DataClassification("confidential"), DataClassificationConfidential)
	assert.Equal(t, DataClassification("restricted"), DataClassificationRestricted)
	assert.Equal(t, DataClassification("top_secret"), DataClassificationTopSecret)
}

func TestPrivacyManager_DataTypeConstants(t *testing.T) {
	// Test data type constants
	assert.Equal(t, DataType("personal"), DataTypePersonal)
	assert.Equal(t, DataType("financial"), DataTypeFinancial)
	assert.Equal(t, DataType("health"), DataTypeHealth)
	assert.Equal(t, DataType("biometric"), DataTypeBiometric)
	assert.Equal(t, DataType("location"), DataTypeLocation)
	assert.Equal(t, DataType("behavioral"), DataTypeBehavioral)
	assert.Equal(t, DataType("technical"), DataTypeTechnical)
	assert.Equal(t, DataType("business"), DataTypeBusiness)
}

func TestPrivacyManager_PrivacyLawConstants(t *testing.T) {
	// Test privacy law constants
	assert.Equal(t, PrivacyLaw("gdpr"), PrivacyLawGDPR)
	assert.Equal(t, PrivacyLaw("ccpa"), PrivacyLawCCPA)
	assert.Equal(t, PrivacyLaw("hipaa"), PrivacyLawHIPAA)
	assert.Equal(t, PrivacyLaw("pipeda"), PrivacyLawPIPEDA)
	assert.Equal(t, PrivacyLaw("lgpd"), PrivacyLawLGPD)
	assert.Equal(t, PrivacyLaw("pdpa"), PrivacyLawPDPA)
}

func TestPrivacyManager_EdgeCases(t *testing.T) {
	pm := NewPrivacyManager()
	
	// Test with nil data subject
	err := pm.CreateDataSubject(nil)
	assert.Error(t, err)
	
	// Test with empty ID
	emptySubject := &DataSubject{ID: ""}
	err = pm.CreateDataSubject(emptySubject)
	assert.NoError(t, err) // Should not error, but may not be useful
	
	// Test with very long retention period
	longRetentionPolicy := &DataRetentionPolicy{
		ID:              "long_retention",
		Name:            "Long Retention",
		DataTypes:       []DataType{DataTypePersonal},
		RetentionPeriod: 100 * 365 * 24 * time.Hour, // 100 years
	}
	
	err = pm.CreateDataRetentionPolicy(longRetentionPolicy)
	assert.NoError(t, err)
	
	// Test data retention with very old data
	veryOldData := time.Now().Add(-200 * 365 * 24 * time.Hour) // 200 years ago
	shouldRetain, _, err := pm.CheckDataRetention(context.Background(), DataTypePersonal, veryOldData)
	assert.NoError(t, err)
	assert.False(t, shouldRetain) // Should not retain very old data
}
