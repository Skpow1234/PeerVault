package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// CertificateManager manages SSL/TLS certificates
type CertificateManager struct {
	configDir string
	certs     map[string]*Certificate
}

// Certificate represents a certificate with metadata
type Certificate struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // "server", "client", "ca"
	Subject   string    `json:"subject"`
	Issuer    string    `json:"issuer"`
	Serial    string    `json:"serial"`
	NotBefore time.Time `json:"not_before"`
	NotAfter  time.Time `json:"not_after"`
	KeySize   int       `json:"key_size"`
	Algorithm string    `json:"algorithm"`
	FilePath  string    `json:"file_path"`
	KeyPath   string    `json:"key_path"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"` // "valid", "expired", "revoked"
}

// NewCertificateManager creates a new certificate manager
func NewCertificateManager(configDir string) *CertificateManager {
	cm := &CertificateManager{
		configDir: configDir,
		certs:     make(map[string]*Certificate),
	}

	cm.loadCertificates()
	return cm
}

// GenerateSelfSignedCert generates a self-signed certificate
func (cm *CertificateManager) GenerateSelfSignedCert(name, subject string, validityDays int) (*Certificate, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"PeerVault"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    subject,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(0, 0, validityDays),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		DNSNames:    []string{"localhost", subject},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Generate certificate ID
	certID := cm.generateID()

	// Create certificate object
	cert := &Certificate{
		ID:        certID,
		Name:      name,
		Type:      "server",
		Subject:   subject,
		Issuer:    subject, // Self-signed
		Serial:    template.SerialNumber.String(),
		NotBefore: template.NotBefore,
		NotAfter:  template.NotAfter,
		KeySize:   2048,
		Algorithm: "RSA",
		FilePath:  filepath.Join(cm.configDir, "certs", certID+".crt"),
		KeyPath:   filepath.Join(cm.configDir, "certs", certID+".key"),
		CreatedAt: time.Now(),
		Status:    "valid",
	}

	// Save certificate and key to files
	if err := cm.saveCertificateFiles(cert, certDER, privateKey); err != nil {
		return nil, fmt.Errorf("failed to save certificate files: %w", err)
	}

	// Store certificate metadata
	cm.certs[certID] = cert
	cm.saveCertificates()

	return cert, nil
}

// GenerateCSR generates a certificate signing request
func (cm *CertificateManager) GenerateCSR(name, subject string) ([]byte, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create CSR template
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			Organization:  []string{"PeerVault"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    subject,
		},
		DNSNames: []string{subject},
	}

	// Create CSR
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSR: %w", err)
	}

	// Encode CSR to PEM
	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrDER,
	})

	return csrPEM, nil
}

// LoadCertificate loads a certificate from file
func (cm *CertificateManager) LoadCertificate(name, certPath, keyPath string) (*Certificate, error) {
	// Load certificate file
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Parse certificate
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Generate certificate ID
	certID := cm.generateID()

	// Create certificate object
	certObj := &Certificate{
		ID:        certID,
		Name:      name,
		Type:      "server",
		Subject:   cert.Subject.CommonName,
		Issuer:    cert.Issuer.CommonName,
		Serial:    cert.SerialNumber.String(),
		NotBefore: cert.NotBefore,
		NotAfter:  cert.NotAfter,
		KeySize:   2048, // Default, could be determined from key
		Algorithm: "RSA",
		FilePath:  certPath,
		KeyPath:   keyPath,
		CreatedAt: time.Now(),
		Status:    cm.getCertificateStatus(cert),
	}

	// Store certificate metadata
	cm.certs[certID] = certObj
	cm.saveCertificates()

	return certObj, nil
}

// GetCertificate returns a certificate by ID
func (cm *CertificateManager) GetCertificate(id string) (*Certificate, error) {
	cert, exists := cm.certs[id]
	if !exists {
		return nil, fmt.Errorf("certificate not found")
	}
	return cert, nil
}

// ListCertificates returns all certificates
func (cm *CertificateManager) ListCertificates() []*Certificate {
	var certs []*Certificate
	for _, cert := range cm.certs {
		certs = append(certs, cert)
	}
	return certs
}

// ValidateCertificate validates a certificate
func (cm *CertificateManager) ValidateCertificate(id string) error {
	cert, exists := cm.certs[id]
	if !exists {
		return fmt.Errorf("certificate not found")
	}

	// Check if certificate file exists
	if _, err := os.Stat(cert.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("certificate file not found")
	}

	// Check if key file exists
	if _, err := os.Stat(cert.KeyPath); os.IsNotExist(err) {
		return fmt.Errorf("private key file not found")
	}

	// Load and validate certificate
	_, err := tls.LoadX509KeyPair(cert.FilePath, cert.KeyPath)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %w", err)
	}

	// Update status
	cert.Status = "valid"
	cm.saveCertificates()

	return nil
}

// RevokeCertificate revokes a certificate
func (cm *CertificateManager) RevokeCertificate(id string) error {
	cert, exists := cm.certs[id]
	if !exists {
		return fmt.Errorf("certificate not found")
	}

	cert.Status = "revoked"
	cm.saveCertificates()

	return nil
}

// DeleteCertificate deletes a certificate
func (cm *CertificateManager) DeleteCertificate(id string) error {
	cert, exists := cm.certs[id]
	if !exists {
		return fmt.Errorf("certificate not found")
	}

	// Delete certificate file
	if err := os.Remove(cert.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete certificate file: %w", err)
	}

	// Delete key file
	if err := os.Remove(cert.KeyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete key file: %w", err)
	}

	// Remove from memory
	delete(cm.certs, id)
	cm.saveCertificates()

	return nil
}

// CheckExpiringCertificates returns certificates expiring within specified days
func (cm *CertificateManager) CheckExpiringCertificates(days int) []*Certificate {
	var expiring []*Certificate
	threshold := time.Now().AddDate(0, 0, days)

	for _, cert := range cm.certs {
		if cert.NotAfter.Before(threshold) && cert.Status == "valid" {
			expiring = append(expiring, cert)
		}
	}

	return expiring
}

// Utility functions
func (cm *CertificateManager) generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

func (cm *CertificateManager) getCertificateStatus(cert *x509.Certificate) string {
	now := time.Now()
	if now.After(cert.NotAfter) {
		return "expired"
	}
	if now.Before(cert.NotBefore) {
		return "not_yet_valid"
	}
	return "valid"
}

func (cm *CertificateManager) saveCertificateFiles(cert *Certificate, certDER []byte, privateKey *rsa.PrivateKey) error {
	// Create certs directory
	certsDir := filepath.Join(cm.configDir, "certs")
	if err := os.MkdirAll(certsDir, 0700); err != nil {
		return fmt.Errorf("failed to create certs directory: %w", err)
	}

	// Save certificate
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	if err := os.WriteFile(cert.FilePath, certPEM, 0644); err != nil {
		return fmt.Errorf("failed to write certificate file: %w", err)
	}

	// Save private key
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	if err := os.WriteFile(cert.KeyPath, keyPEM, 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}

func (cm *CertificateManager) loadCertificates() error {
	certsFile := filepath.Join(cm.configDir, "certificates.json")
	if _, err := os.Stat(certsFile); os.IsNotExist(err) {
		return nil // File doesn't exist, start with empty certificates
	}

	data, err := os.ReadFile(certsFile)
	if err != nil {
		return fmt.Errorf("failed to read certificates file: %w", err)
	}

	var certs []*Certificate
	if err := json.Unmarshal(data, &certs); err != nil {
		return fmt.Errorf("failed to unmarshal certificates: %w", err)
	}

	cm.certs = make(map[string]*Certificate)
	for _, cert := range certs {
		cm.certs[cert.ID] = cert
	}

	return nil
}

func (cm *CertificateManager) saveCertificates() error {
	certsFile := filepath.Join(cm.configDir, "certificates.json")

	var certs []*Certificate
	for _, cert := range cm.certs {
		certs = append(certs, cert)
	}

	data, err := json.MarshalIndent(certs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal certificates: %w", err)
	}

	return os.WriteFile(certsFile, data, 0600)
}
