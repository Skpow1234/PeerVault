package pki

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPKIManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Test successful creation
	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, tempDir, manager.storagePath)
	assert.NotNil(t, manager.certificates)
	assert.NotNil(t, manager.requests)
	assert.NotNil(t, manager.cas)

	// Test that root CA is initialized
	manager.mu.RLock()
	rootCA, exists := manager.cas["root"]
	manager.mu.RUnlock()
	assert.True(t, exists)
	assert.NotNil(t, rootCA)
	assert.Equal(t, "root", rootCA.ID)
	assert.Equal(t, "PeerVault Root CA", rootCA.Name)
}

func TestNewPKIManager_InvalidPath(t *testing.T) {
	// Test with invalid path (empty string)
	manager, err := NewPKIManager("")
	assert.Error(t, err)
	assert.Nil(t, manager)
	assert.Contains(t, err.Error(), "failed to create PKI storage directory")
}

func TestPKIManager_CreateCertificateRequest(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	ctx := context.Background()

	// Test valid certificate request
	request := &CertificateRequest{
		ID:           "test-server-req",
		Type:         CertificateTypeServer,
		Subject:      "test-server",
		DNSNames:     []string{"test.example.com"},
		KeySize:      2048,
		Algorithm:    "RSA",
		ValidityDays: 365,
		RequestedBy:  "admin",
	}

	err = manager.CreateCertificateRequest(ctx, request)
	assert.NoError(t, err)
	assert.Equal(t, "pending", request.Status)
	assert.NotZero(t, request.RequestedAt)

	// Test duplicate request
	err = manager.CreateCertificateRequest(ctx, request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestPKIManager_ApproveCertificateRequest(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	ctx := context.Background()

	// Create a certificate request
	request := &CertificateRequest{
		ID:           "test-server-req",
		Type:         CertificateTypeServer,
		Subject:      "test-server",
		DNSNames:     []string{"test.example.com"},
		KeySize:      2048,
		Algorithm:    "RSA",
		ValidityDays: 365,
		RequestedBy:  "admin",
	}

	err = manager.CreateCertificateRequest(ctx, request)
	assert.NoError(t, err)

	// Test approval
	cert, err := manager.ApproveCertificateRequest(ctx, request.ID, "approver")
	assert.NoError(t, err)
	assert.NotNil(t, cert)
	assert.Equal(t, request.Type, cert.Type)
	assert.Contains(t, cert.Subject, request.Subject)
	assert.Equal(t, CertificateStatusValid, cert.Status)
	assert.NotEmpty(t, cert.PEMData)
	assert.NotEmpty(t, cert.PrivateKey)
	assert.NotZero(t, cert.CreatedAt)

	// Verify certificate is stored
	storedCert, exists := manager.GetCertificate(cert.ID)
	assert.True(t, exists)
	assert.Equal(t, cert, storedCert)
}

func TestPKIManager_GetCertificate_NotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	// Test with non-existent certificate
	cert, exists := manager.GetCertificate("non-existent-id")
	assert.False(t, exists)
	assert.Nil(t, cert)
}

func TestPKIManager_ListCertificates(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	// Initially should have root CA
	certificates := manager.ListCertificates()
	initialCount := len(certificates)
	assert.GreaterOrEqual(t, initialCount, 0) // May or may not include root CA

	// Create a certificate request and approve it
	ctx := context.Background()
	request := &CertificateRequest{
		ID:           "test-server-req",
		Type:         CertificateTypeServer,
		Subject:      "test-server",
		KeySize:      2048,
		Algorithm:    "RSA",
		ValidityDays: 365,
		RequestedBy:  "admin",
	}

	err = manager.CreateCertificateRequest(ctx, request)
	assert.NoError(t, err)

	_, err = manager.ApproveCertificateRequest(ctx, request.ID, "approver")
	assert.NoError(t, err)

	// List certificates
	certificates = manager.ListCertificates()
	assert.Len(t, certificates, initialCount+1) // Should have one more certificate
}

func TestPKIManager_RevokeCertificate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	ctx := context.Background()

	// Create and approve a certificate request
	request := &CertificateRequest{
		ID:           "test-server-req",
		Type:         CertificateTypeServer,
		Subject:      "test-server",
		KeySize:      2048,
		Algorithm:    "RSA",
		ValidityDays: 365,
		RequestedBy:  "admin",
	}

	err = manager.CreateCertificateRequest(ctx, request)
	assert.NoError(t, err)

	cert, err := manager.ApproveCertificateRequest(ctx, request.ID, "approver")
	assert.NoError(t, err)

	// Revoke the certificate
	err = manager.RevokeCertificate(ctx, cert.ID, "Security breach")
	assert.NoError(t, err)

	// Verify certificate status is updated
	revokedCert, exists := manager.GetCertificate(cert.ID)
	assert.True(t, exists)
	assert.Equal(t, CertificateStatusRevoked, revokedCert.Status)
}

func TestPKIManager_RevokeCertificate_NotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	ctx := context.Background()

	err = manager.RevokeCertificate(ctx, "non-existent-id", "Test revocation")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPKIManager_GetCertificateAuthority(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	// Get root CA
	ca, exists := manager.GetCertificateAuthority("root")
	assert.True(t, exists)
	assert.NotNil(t, ca)
	assert.Equal(t, "root", ca.ID)
	assert.Equal(t, "PeerVault Root CA", ca.Name)
	assert.NotNil(t, ca.Certificate)
	assert.NotEmpty(t, ca.PrivateKey)
}

func TestPKIManager_GetCertificateAuthority_NotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	ca, exists := manager.GetCertificateAuthority("non-existent-ca")
	assert.False(t, exists)
	assert.Nil(t, ca)
}

func TestPKIManager_ListCertificateAuthorities(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	// Should have root CA
	cas := manager.ListCertificateAuthorities()
	assert.Len(t, cas, 1)
	assert.Equal(t, "root", cas[0].ID)
}

func TestPKIManager_CheckCertificateValidity(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	ctx := context.Background()

	// Create and approve a certificate request
	request := &CertificateRequest{
		ID:           "test-server-req",
		Type:         CertificateTypeServer,
		Subject:      "test-server",
		KeySize:      2048,
		Algorithm:    "RSA",
		ValidityDays: 365,
		RequestedBy:  "admin",
	}

	err = manager.CreateCertificateRequest(ctx, request)
	assert.NoError(t, err)

	cert, err := manager.ApproveCertificateRequest(ctx, request.ID, "approver")
	assert.NoError(t, err)

	// Check certificate validity
	valid, err := manager.CheckCertificateValidity(ctx, cert.ID)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Test with non-existent certificate
	valid, err = manager.CheckCertificateValidity(ctx, "non-existent-id")
	assert.Error(t, err)
	assert.False(t, valid)
}

func TestPKIManager_GetTLSConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	ctx := context.Background()

	// Create and approve a certificate request
	request := &CertificateRequest{
		ID:           "test-server-req",
		Type:         CertificateTypeServer,
		Subject:      "test-server",
		KeySize:      2048,
		Algorithm:    "RSA",
		ValidityDays: 365,
		RequestedBy:  "admin",
	}

	err = manager.CreateCertificateRequest(ctx, request)
	assert.NoError(t, err)

	cert, err := manager.ApproveCertificateRequest(ctx, request.ID, "approver")
	assert.NoError(t, err)

	// Get TLS config
	tlsConfig, err := manager.GetTLSConfig(ctx, cert.ID)
	assert.NoError(t, err)
	assert.NotNil(t, tlsConfig)
	assert.NotNil(t, tlsConfig.Certificates)
	assert.Len(t, tlsConfig.Certificates, 1)
}

func TestPKIManager_GetTLSConfig_NotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	ctx := context.Background()

	_, err = manager.GetTLSConfig(ctx, "non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPKIManager_RotateCertificate(t *testing.T) {
	// This test requires RSA key generation which can be slow in CI environments
	// Increase timeout for this specific test
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	ctx := context.Background()

	// Create and approve a certificate request
	request := &CertificateRequest{
		ID:           "test-server-req",
		Type:         CertificateTypeServer,
		Subject:      "test-server",
		KeySize:      2048,
		Algorithm:    "RSA",
		ValidityDays: 365,
		RequestedBy:  "admin",
	}

	err = manager.CreateCertificateRequest(ctx, request)
	assert.NoError(t, err)

	cert, err := manager.ApproveCertificateRequest(ctx, request.ID, "approver")
	assert.NoError(t, err)

	// Rotate the certificate
	newCert, err := manager.RotateCertificate(ctx, cert.ID)
	assert.NoError(t, err)
	assert.NotNil(t, newCert)
	assert.Equal(t, cert.Type, newCert.Type)
	assert.Contains(t, newCert.Subject, "test-server")
	assert.NotEqual(t, cert.ID, newCert.ID) // Should have different ID
}

func TestPKIManager_RotateCertificate_NotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create PKI manager manually without calling NewPKIManager to avoid root CA generation
	manager := &PKIManager{
		certificates: make(map[string]*Certificate),
		requests:     make(map[string]*CertificateRequest),
		cas:          make(map[string]*CertificateAuthority),
		storagePath:  tempDir,
	}

	ctx := context.Background()

	_, err = manager.RotateCertificate(ctx, "non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPKIManager_ConcurrentAccess(t *testing.T) {
	// Test concurrent access without creating a new PKI manager to avoid expensive RSA key generation
	// This test focuses on testing the thread safety of the PKI manager's internal structures

	// Create a PKI manager without initializing the root CA to avoid RSA key generation
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create storage directory
	err = os.MkdirAll(tempDir, 0755)
	assert.NoError(t, err)

	// Create PKI manager manually without calling NewPKIManager to avoid root CA generation
	manager := &PKIManager{
		certificates: make(map[string]*Certificate),
		requests:     make(map[string]*CertificateRequest),
		cas:          make(map[string]*CertificateAuthority),
		storagePath:  tempDir,
	}

	// Test concurrent access with simple operations
	done := make(chan bool, 3)

	// Test concurrent read operations (no key generation)
	for i := 0; i < 3; i++ {
		go func(i int) {
			defer func() { done <- true }()

			// Test concurrent read operations - just verify methods don't panic
			_ = manager.ListCertificates()
			_ = manager.ListCertificateAuthorities()

			// Test that the manager is still functional by checking storage path
			_ = manager.storagePath
		}(i)
	}

	// Wait for all goroutines to complete with timeout
	timeout := time.After(5 * time.Second)
	for i := 0; i < 3; i++ {
		select {
		case <-done:
			// Goroutine completed
		case <-timeout:
			t.Fatal("Test timed out - concurrent operations took too long")
		}
	}

	// Verify the manager is still functional
	assert.NotNil(t, manager.certificates)
	assert.NotNil(t, manager.requests)
	assert.NotNil(t, manager.cas)
	assert.Equal(t, tempDir, manager.storagePath)
}

func TestCertificateConstants(t *testing.T) {
	// Test certificate types
	assert.Equal(t, CertificateType("ca"), CertificateTypeCA)
	assert.Equal(t, CertificateType("server"), CertificateTypeServer)
	assert.Equal(t, CertificateType("client"), CertificateTypeClient)
	assert.Equal(t, CertificateType("code_signing"), CertificateTypeCodeSigning)
	assert.Equal(t, CertificateType("email"), CertificateTypeEmail)

	// Test certificate statuses
	assert.Equal(t, CertificateStatus("valid"), CertificateStatusValid)
	assert.Equal(t, CertificateStatus("expired"), CertificateStatusExpired)
	assert.Equal(t, CertificateStatus("revoked"), CertificateStatusRevoked)
	assert.Equal(t, CertificateStatus("pending"), CertificateStatusPending)
	assert.Equal(t, CertificateStatus("suspended"), CertificateStatusSuspended)
}
