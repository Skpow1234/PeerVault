package pki

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPKIManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

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
	defer os.RemoveAll(tempDir)

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
	defer os.RemoveAll(tempDir)

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
	defer os.RemoveAll(tempDir)

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
	defer os.RemoveAll(tempDir)

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
	defer os.RemoveAll(tempDir)

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
	defer os.RemoveAll(tempDir)

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
	defer os.RemoveAll(tempDir)

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
	defer os.RemoveAll(tempDir)

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	ca, exists := manager.GetCertificateAuthority("non-existent-ca")
	assert.False(t, exists)
	assert.Nil(t, ca)
}

func TestPKIManager_ListCertificateAuthorities(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

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
	defer os.RemoveAll(tempDir)

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
	defer os.RemoveAll(tempDir)

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
	defer os.RemoveAll(tempDir)

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	ctx := context.Background()

	_, err = manager.GetTLSConfig(ctx, "non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPKIManager_RotateCertificate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

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
	defer os.RemoveAll(tempDir)

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	ctx := context.Background()

	_, err = manager.RotateCertificate(ctx, "non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPKIManager_ConcurrentAccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager, err := NewPKIManager(tempDir)
	assert.NoError(t, err)

	ctx := context.Background()

	// Test concurrent access
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(i int) {
			defer func() { done <- true }()

			// Create certificate request
			request := &CertificateRequest{
				ID:           "concurrent-server-" + string(rune(i)),
				Type:         CertificateTypeServer,
				Subject:      "CN=concurrent-server-" + string(rune(i)),
				KeySize:      2048,
				Algorithm:    "RSA",
				ValidityDays: 365,
				RequestedBy:  "admin",
			}

			err := manager.CreateCertificateRequest(ctx, request)
			assert.NoError(t, err)

			// Approve the request
			cert, err := manager.ApproveCertificateRequest(ctx, request.ID, "approver")
			assert.NoError(t, err)
			assert.NotNil(t, cert)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify all operations completed
	certificates := manager.ListCertificates()
	assert.Len(t, certificates, 5) // 5 new certificates (root CA may not be in list)
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
