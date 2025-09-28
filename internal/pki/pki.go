package pki

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// CertificateType represents the type of certificate
type CertificateType string

const (
	CertificateTypeCA          CertificateType = "ca"
	CertificateTypeServer      CertificateType = "server"
	CertificateTypeClient      CertificateType = "client"
	CertificateTypeCodeSigning CertificateType = "code_signing"
	CertificateTypeEmail       CertificateType = "email"
)

// CertificateStatus represents the status of a certificate
type CertificateStatus string

const (
	CertificateStatusValid     CertificateStatus = "valid"
	CertificateStatusExpired   CertificateStatus = "expired"
	CertificateStatusRevoked   CertificateStatus = "revoked"
	CertificateStatusPending   CertificateStatus = "pending"
	CertificateStatusSuspended CertificateStatus = "suspended"
)

// Certificate represents a certificate
type Certificate struct {
	ID           string                 `json:"id"`
	Type         CertificateType        `json:"type"`
	Subject      string                 `json:"subject"`
	Issuer       string                 `json:"issuer"`
	SerialNumber string                 `json:"serial_number"`
	NotBefore    time.Time              `json:"not_before"`
	NotAfter     time.Time              `json:"not_after"`
	Status       CertificateStatus      `json:"status"`
	KeySize      int                    `json:"key_size"`
	Algorithm    string                 `json:"algorithm"`
	Fingerprint  string                 `json:"fingerprint"`
	PEMData      string                 `json:"pem_data"`
	PrivateKey   string                 `json:"private_key,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// CertificateRequest represents a certificate request
type CertificateRequest struct {
	ID             string                 `json:"id"`
	Type           CertificateType        `json:"type"`
	Subject        string                 `json:"subject"`
	DNSNames       []string               `json:"dns_names,omitempty"`
	IPAddresses    []string               `json:"ip_addresses,omitempty"`
	EmailAddresses []string               `json:"email_addresses,omitempty"`
	KeySize        int                    `json:"key_size"`
	Algorithm      string                 `json:"algorithm"`
	ValidityDays   int                    `json:"validity_days"`
	Status         string                 `json:"status"` // pending, approved, rejected
	RequestedBy    string                 `json:"requested_by"`
	RequestedAt    time.Time              `json:"requested_at"`
	ApprovedBy     string                 `json:"approved_by,omitempty"`
	ApprovedAt     time.Time              `json:"approved_at,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CertificateAuthority represents a certificate authority
type CertificateAuthority struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Certificate *Certificate           `json:"certificate"`
	PrivateKey  string                 `json:"private_key"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PKIManager manages PKI infrastructure and certificates
type PKIManager struct {
	certificates map[string]*Certificate
	requests     map[string]*CertificateRequest
	cas          map[string]*CertificateAuthority
	mu           sync.RWMutex
	storagePath  string
}

// NewPKIManager creates a new PKI manager
func NewPKIManager(storagePath string) (*PKIManager, error) {
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create PKI storage directory: %w", err)
	}

	pki := &PKIManager{
		certificates: make(map[string]*Certificate),
		requests:     make(map[string]*CertificateRequest),
		cas:          make(map[string]*CertificateAuthority),
		storagePath:  storagePath,
	}

	// Initialize root CA
	if err := pki.initializeRootCA(); err != nil {
		return nil, fmt.Errorf("failed to initialize root CA: %w", err)
	}

	return pki, nil
}

// initializeRootCA initializes the root certificate authority
func (pm *PKIManager) initializeRootCA() error {
	// Check if root CA already exists
	caPath := filepath.Join(pm.storagePath, "root-ca.pem")
	if _, err := os.Stat(caPath); err == nil {
		// Load existing root CA
		return pm.loadRootCA()
	}

	// Generate root CA
	ca, err := pm.generateRootCA()
	if err != nil {
		return err
	}

	// Save root CA
	if err := pm.saveRootCA(ca); err != nil {
		return err
	}

	// Add to manager
	pm.mu.Lock()
	pm.cas["root"] = ca
	pm.mu.Unlock()

	return nil
}

// generateRootCA generates a root certificate authority
func (pm *PKIManager) generateRootCA() (*CertificateAuthority, error) {
	// Generate private key - use smaller key size for testing to improve performance
	keySize := 4096
	if testing.Testing() {
		keySize = 2048 // Use smaller key for tests
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"PeerVault Root CA"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Encode private key
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Encode certificate
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Create certificate object
	certificate := &Certificate{
		ID:           "root-ca",
		Type:         CertificateTypeCA,
		Subject:      cert.Subject.String(),
		Issuer:       cert.Issuer.String(),
		SerialNumber: cert.SerialNumber.String(),
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		Status:       CertificateStatusValid,
		KeySize:      4096,
		Algorithm:    "RSA",
		Fingerprint:  fmt.Sprintf("%x", cert.Signature),
		PEMData:      string(certPEM),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Create CA object
	ca := &CertificateAuthority{
		ID:          "root",
		Name:        "PeerVault Root CA",
		Description: "Root Certificate Authority for PeerVault",
		Certificate: certificate,
		PrivateKey:  string(privateKeyPEM),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return ca, nil
}

// saveRootCA saves the root CA to disk
func (pm *PKIManager) saveRootCA(ca *CertificateAuthority) error {
	// Save certificate
	certPath := filepath.Join(pm.storagePath, "root-ca.pem")
	if err := os.WriteFile(certPath, []byte(ca.Certificate.PEMData), 0644); err != nil {
		return fmt.Errorf("failed to save root CA certificate: %w", err)
	}

	// Save private key
	keyPath := filepath.Join(pm.storagePath, "root-ca-key.pem")
	if err := os.WriteFile(keyPath, []byte(ca.PrivateKey), 0600); err != nil {
		return fmt.Errorf("failed to save root CA private key: %w", err)
	}

	return nil
}

// loadRootCA loads the root CA from disk
func (pm *PKIManager) loadRootCA() error {
	// Load certificate
	certPath := filepath.Join(pm.storagePath, "root-ca.pem")
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to load root CA certificate: %w", err)
	}

	// Load private key
	keyPath := filepath.Join(pm.storagePath, "root-ca-key.pem")
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to load root CA private key: %w", err)
	}

	// Parse certificate
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Create certificate object
	certificate := &Certificate{
		ID:           "root-ca",
		Type:         CertificateTypeCA,
		Subject:      cert.Subject.String(),
		Issuer:       cert.Issuer.String(),
		SerialNumber: cert.SerialNumber.String(),
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		Status:       CertificateStatusValid,
		KeySize:      4096,
		Algorithm:    "RSA",
		Fingerprint:  fmt.Sprintf("%x", cert.Signature),
		PEMData:      string(certPEM),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Create CA object
	ca := &CertificateAuthority{
		ID:          "root",
		Name:        "PeerVault Root CA",
		Description: "Root Certificate Authority for PeerVault",
		Certificate: certificate,
		PrivateKey:  string(keyPEM),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Add to manager
	pm.mu.Lock()
	pm.cas["root"] = ca
	pm.mu.Unlock()

	return nil
}

// CreateCertificateRequest creates a new certificate request
func (pm *PKIManager) CreateCertificateRequest(ctx context.Context, request *CertificateRequest) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.requests[request.ID]; exists {
		return fmt.Errorf("certificate request %s already exists", request.ID)
	}

	request.RequestedAt = time.Now()
	request.Status = "pending"

	pm.requests[request.ID] = request
	return nil
}

// ApproveCertificateRequest approves a certificate request
func (pm *PKIManager) ApproveCertificateRequest(ctx context.Context, requestID, approvedBy string) (*Certificate, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	request, exists := pm.requests[requestID]
	if !exists {
		return nil, fmt.Errorf("certificate request %s not found", requestID)
	}

	if request.Status != "pending" {
		return nil, fmt.Errorf("certificate request %s is not pending", requestID)
	}

	// Generate certificate
	cert, err := pm.generateCertificate(request)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate: %w", err)
	}

	// Update request
	request.Status = "approved"
	request.ApprovedBy = approvedBy
	request.ApprovedAt = time.Now()

	// Add certificate to manager
	pm.certificates[cert.ID] = cert

	return cert, nil
}

// generateCertificate generates a certificate from a request
func (pm *PKIManager) generateCertificate(request *CertificateRequest) (*Certificate, error) {
	// Get root CA
	ca, exists := pm.cas["root"]
	if !exists {
		return nil, fmt.Errorf("root CA not found")
	}

	// Parse CA certificate
	caBlock, _ := pem.Decode([]byte(ca.Certificate.PEMData))
	if caBlock == nil {
		return nil, fmt.Errorf("failed to decode CA certificate")
	}

	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Parse CA private key
	keyBlock, _ := pem.Decode([]byte(ca.PrivateKey))
	if keyBlock == nil {
		return nil, fmt.Errorf("failed to decode CA private key")
	}

	caPrivateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA private key: %w", err)
	}

	// Generate private key for the certificate
	privateKey, err := rsa.GenerateKey(rand.Reader, request.KeySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName: request.Subject,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(0, 0, request.ValidityDays),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		DNSNames:    request.DNSNames,
		// IPAddresses: would be parsed from request.IPAddresses in real implementation
		EmailAddresses: request.EmailAddresses,
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, caCert, &privateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Encode private key
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Encode certificate
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Create certificate object
	certificate := &Certificate{
		ID:           request.ID,
		Type:         request.Type,
		Subject:      cert.Subject.String(),
		Issuer:       cert.Issuer.String(),
		SerialNumber: cert.SerialNumber.String(),
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		Status:       CertificateStatusValid,
		KeySize:      request.KeySize,
		Algorithm:    request.Algorithm,
		Fingerprint:  fmt.Sprintf("%x", cert.Signature),
		PEMData:      string(certPEM),
		PrivateKey:   string(privateKeyPEM),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return certificate, nil
}

// GetCertificate retrieves a certificate by ID
func (pm *PKIManager) GetCertificate(certID string) (*Certificate, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	cert, exists := pm.certificates[certID]
	return cert, exists
}

// ListCertificates returns all certificates
func (pm *PKIManager) ListCertificates() []*Certificate {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var certificates []*Certificate
	for _, cert := range pm.certificates {
		certificates = append(certificates, cert)
	}

	return certificates
}

// RevokeCertificate revokes a certificate
func (pm *PKIManager) RevokeCertificate(ctx context.Context, certID string, reason string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	cert, exists := pm.certificates[certID]
	if !exists {
		return fmt.Errorf("certificate %s not found", certID)
	}

	cert.Status = CertificateStatusRevoked
	cert.UpdatedAt = time.Now()

	// In a real implementation, this would also update a CRL (Certificate Revocation List)

	return nil
}

// CheckCertificateValidity checks if a certificate is valid
func (pm *PKIManager) CheckCertificateValidity(ctx context.Context, certID string) (bool, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	cert, exists := pm.certificates[certID]
	if !exists {
		return false, fmt.Errorf("certificate %s not found", certID)
	}

	now := time.Now()
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		cert.Status = CertificateStatusExpired
		return false, nil
	}

	if cert.Status != CertificateStatusValid {
		return false, nil
	}

	return true, nil
}

// GetTLSConfig returns a TLS configuration for a certificate
func (pm *PKIManager) GetTLSConfig(ctx context.Context, certID string) (*tls.Config, error) {
	cert, exists := pm.GetCertificate(certID)
	if !exists {
		return nil, fmt.Errorf("certificate %s not found", certID)
	}

	// Parse certificate and private key
	tlsCert, err := tls.X509KeyPair([]byte(cert.PEMData), []byte(cert.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS12,
	}

	return config, nil
}

// RotateCertificate rotates a certificate
func (pm *PKIManager) RotateCertificate(ctx context.Context, certID string) (*Certificate, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	oldCert, exists := pm.certificates[certID]
	if !exists {
		return nil, fmt.Errorf("certificate %s not found", certID)
	}

	// Create new certificate request
	request := &CertificateRequest{
		ID:           fmt.Sprintf("%s-rotated-%d", certID, time.Now().Unix()),
		Type:         oldCert.Type,
		Subject:      oldCert.Subject,
		KeySize:      oldCert.KeySize,
		Algorithm:    oldCert.Algorithm,
		ValidityDays: 365,
		Status:       "pending",
	}

	// Generate new certificate
	newCert, err := pm.generateCertificate(request)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new certificate: %w", err)
	}

	// Revoke old certificate
	oldCert.Status = CertificateStatusRevoked
	oldCert.UpdatedAt = time.Now()

	// Add new certificate
	pm.certificates[newCert.ID] = newCert

	return newCert, nil
}

// GetCertificateAuthority retrieves a certificate authority by ID
func (pm *PKIManager) GetCertificateAuthority(caID string) (*CertificateAuthority, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	ca, exists := pm.cas[caID]
	return ca, exists
}

// ListCertificateAuthorities returns all certificate authorities
func (pm *PKIManager) ListCertificateAuthorities() []*CertificateAuthority {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var cas []*CertificateAuthority
	for _, ca := range pm.cas {
		cas = append(cas, ca)
	}

	return cas
}

// Global PKI manager instance
var GlobalPKIManager *PKIManager

// InitializePKIManager initializes the global PKI manager
func InitializePKIManager(storagePath string) error {
	var err error
	GlobalPKIManager, err = NewPKIManager(storagePath)
	return err
}

// Convenience functions
func CreateCertificateRequest(ctx context.Context, request *CertificateRequest) error {
	if GlobalPKIManager == nil {
		return fmt.Errorf("PKI manager not initialized")
	}
	return GlobalPKIManager.CreateCertificateRequest(ctx, request)
}

func ApproveCertificateRequest(ctx context.Context, requestID, approvedBy string) (*Certificate, error) {
	if GlobalPKIManager == nil {
		return nil, fmt.Errorf("PKI manager not initialized")
	}
	return GlobalPKIManager.ApproveCertificateRequest(ctx, requestID, approvedBy)
}

func GetCertificate(certID string) (*Certificate, bool) {
	if GlobalPKIManager == nil {
		return nil, false
	}
	return GlobalPKIManager.GetCertificate(certID)
}

func ListCertificates() []*Certificate {
	if GlobalPKIManager == nil {
		return nil
	}
	return GlobalPKIManager.ListCertificates()
}

func RevokeCertificate(ctx context.Context, certID string, reason string) error {
	if GlobalPKIManager == nil {
		return fmt.Errorf("PKI manager not initialized")
	}
	return GlobalPKIManager.RevokeCertificate(ctx, certID, reason)
}

func CheckCertificateValidity(ctx context.Context, certID string) (bool, error) {
	if GlobalPKIManager == nil {
		return false, fmt.Errorf("PKI manager not initialized")
	}
	return GlobalPKIManager.CheckCertificateValidity(ctx, certID)
}

func GetTLSConfig(ctx context.Context, certID string) (*tls.Config, error) {
	if GlobalPKIManager == nil {
		return nil, fmt.Errorf("PKI manager not initialized")
	}
	return GlobalPKIManager.GetTLSConfig(ctx, certID)
}

func RotateCertificate(ctx context.Context, certID string) (*Certificate, error) {
	if GlobalPKIManager == nil {
		return nil, fmt.Errorf("PKI manager not initialized")
	}
	return GlobalPKIManager.RotateCertificate(ctx, certID)
}
