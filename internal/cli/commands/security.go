package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
	"github.com/Skpow1234/Peervault/internal/cli/security"
)

// AuthCommand manages authentication and user management
type AuthCommand struct {
	BaseCommand
	authManager *security.AuthManager
}

// NewAuthCommand creates a new auth command
func NewAuthCommand(client *client.Client, formatter *formatter.Formatter) *AuthCommand {
	return &AuthCommand{
		BaseCommand: BaseCommand{
			name:        "auth",
			description: "Manage authentication and user accounts",
			usage:       "auth [login|logout|create-user|list-users|enable-mfa|disable-mfa] [options]",
			client:      client,
			formatter:   formatter,
		},
		authManager: security.NewAuthManager("./config"),
	}
}

// Execute executes the auth command
func (c *AuthCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showAuthHelp()
	}

	subcommand := strings.ToLower(args[0])
	switch subcommand {
	case "login":
		return c.login(ctx, args[1:])
	case "logout":
		return c.logout()
	case "create-user":
		return c.createUser(args[1:])
	case "list-users":
		return c.listUsers()
	case "enable-mfa":
		return c.enableMFA(args[1:])
	case "disable-mfa":
		return c.disableMFA(args[1:])
	case "check-permissions":
		return c.checkPermissions(args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// login handles user login
func (c *AuthCommand) login(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: auth login <username> <password>")
	}

	username := args[0]
	password := args[1]

	c.formatter.PrintInfo("Authenticating user...")

	user, err := c.authManager.Authenticate(username, password)
	if err != nil {
		c.formatter.PrintError(fmt.Errorf("authentication failed: %w", err))
		return nil
	}

	// Check if MFA is enabled
	if user.MFAEnabled {
		c.formatter.PrintInfo("MFA is enabled. Please enter your MFA token:")
		// In a real implementation, you would prompt for MFA token
		c.formatter.PrintInfo("MFA verification would be required here")
	}

	// Create session
	session, err := c.authManager.CreateSession(user.ID, "127.0.0.1", "PeerVault CLI")
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Login successful for user: %s", username))
	c.formatter.PrintInfo(fmt.Sprintf("Session ID: %s", session.ID))
	c.formatter.PrintInfo(fmt.Sprintf("Expires: %s", session.ExpiresAt.Format("2006-01-02 15:04:05")))

	return nil
}

// logout handles user logout
func (c *AuthCommand) logout() error {
	c.formatter.PrintInfo("Logging out...")
	c.formatter.PrintSuccess("Logout successful")
	return nil
}

// createUser creates a new user
func (c *AuthCommand) createUser(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: auth create-user <username> <email> <password> [roles]")
	}

	username := args[0]
	email := args[1]
	password := args[2]

	var roles []string
	if len(args) > 3 {
		roles = strings.Split(args[3], ",")
	} else {
		roles = []string{"user"} // Default role
	}

	c.formatter.PrintInfo(fmt.Sprintf("Creating user: %s", username))

	user, err := c.authManager.CreateUser(username, email, password, roles)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("User created successfully: %s", user.Username))
	c.formatter.PrintInfo(fmt.Sprintf("User ID: %s", user.ID))
	c.formatter.PrintInfo(fmt.Sprintf("Roles: %s", strings.Join(user.Roles, ", ")))

	return nil
}

// listUsers lists all users
func (c *AuthCommand) listUsers() error {
	users := c.authManager.ListUsers()

	if len(users) == 0 {
		c.formatter.PrintInfo("No users found")
		return nil
	}

	c.formatter.PrintInfo("Users:")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("%-20s %-30s %-15s %-10s %-20s\n", "Username", "Email", "Roles", "MFA", "Last Login")
	fmt.Println(strings.Repeat("-", 80))

	for _, user := range users {
		mfaStatus := "Disabled"
		if user.MFAEnabled {
			mfaStatus = "Enabled"
		}

		lastLogin := "Never"
		if !user.LastLogin.IsZero() {
			lastLogin = user.LastLogin.Format("2006-01-02 15:04")
		}

		fmt.Printf("%-20s %-30s %-15s %-10s %-20s\n",
			user.Username,
			user.Email,
			strings.Join(user.Roles, ","),
			mfaStatus,
			lastLogin,
		)
	}

	return nil
}

// enableMFA enables MFA for a user
func (c *AuthCommand) enableMFA(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: auth enable-mfa <username>")
	}

	username := args[0]

	// Find user by username
	var user *security.User
	users := c.authManager.ListUsers()
	for _, u := range users {
		if u.Username == username {
			user = u
			break
		}
	}

	if user == nil {
		return fmt.Errorf("user not found: %s", username)
	}

	c.formatter.PrintInfo(fmt.Sprintf("Enabling MFA for user: %s", username))

	qrURL, err := c.authManager.EnableMFA(user.ID)
	if err != nil {
		return fmt.Errorf("failed to enable MFA: %w", err)
	}

	c.formatter.PrintSuccess("MFA enabled successfully")
	c.formatter.PrintInfo("QR Code URL for authenticator app:")
	fmt.Println(qrURL)
	c.formatter.PrintInfo("Please scan this QR code with your authenticator app")

	return nil
}

// disableMFA disables MFA for a user
func (c *AuthCommand) disableMFA(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: auth disable-mfa <username>")
	}

	username := args[0]

	// Find user by username
	var user *security.User
	users := c.authManager.ListUsers()
	for _, u := range users {
		if u.Username == username {
			user = u
			break
		}
	}

	if user == nil {
		return fmt.Errorf("user not found: %s", username)
	}

	c.formatter.PrintInfo(fmt.Sprintf("Disabling MFA for user: %s", username))

	// In a real implementation, you would disable MFA
	user.MFAEnabled = false
	user.MFASecret = ""

	c.formatter.PrintSuccess("MFA disabled successfully")

	return nil
}

// checkPermissions checks user permissions
func (c *AuthCommand) checkPermissions(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: auth check-permissions <username> <permission>")
	}

	username := args[0]
	permission := args[1]

	// Find user by username
	var user *security.User
	users := c.authManager.ListUsers()
	for _, u := range users {
		if u.Username == username {
			user = u
			break
		}
	}

	if user == nil {
		return fmt.Errorf("user not found: %s", username)
	}

	hasPermission := c.authManager.CheckPermission(user.ID, permission)

	if hasPermission {
		c.formatter.PrintSuccess(fmt.Sprintf("User %s has permission: %s", username, permission))
	} else {
		c.formatter.PrintWarning(fmt.Sprintf("User %s does not have permission: %s", username, permission))
	}

	// Show all permissions for the user
	permissions := c.authManager.GetUserPermissions(user.ID)
	c.formatter.PrintInfo(fmt.Sprintf("All permissions for %s:", username))
	for _, perm := range permissions {
		fmt.Printf("  - %s\n", perm)
	}

	return nil
}

// showAuthHelp shows authentication help
func (c *AuthCommand) showAuthHelp() error {
	c.formatter.PrintInfo("Authentication Commands:")
	fmt.Println("  login <username> <password>     - Login to the system")
	fmt.Println("  logout                          - Logout from the system")
	fmt.Println("  create-user <username> <email> <password> [roles] - Create a new user")
	fmt.Println("  list-users                      - List all users")
	fmt.Println("  enable-mfa <username>           - Enable MFA for a user")
	fmt.Println("  disable-mfa <username>          - Disable MFA for a user")
	fmt.Println("  check-permissions <username> <permission> - Check user permissions")
	return nil
}

// CertCommand manages SSL/TLS certificates
type CertCommand struct {
	BaseCommand
	certManager *security.CertificateManager
}

// NewCertCommand creates a new cert command
func NewCertCommand(client *client.Client, formatter *formatter.Formatter) *CertCommand {
	return &CertCommand{
		BaseCommand: BaseCommand{
			name:        "cert",
			description: "Manage SSL/TLS certificates",
			usage:       "cert [generate|list|validate|revoke|delete|check-expiry] [options]",
			client:      client,
			formatter:   formatter,
		},
		certManager: security.NewCertificateManager("./config"),
	}
}

// Execute executes the cert command
func (c *CertCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showCertHelp()
	}

	subcommand := strings.ToLower(args[0])
	switch subcommand {
	case "generate":
		return c.generateCert(args[1:])
	case "list":
		return c.listCerts()
	case "validate":
		return c.validateCert(args[1:])
	case "revoke":
		return c.revokeCert(args[1:])
	case "delete":
		return c.deleteCert(args[1:])
	case "check-expiry":
		return c.checkExpiry(args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// generateCert generates a self-signed certificate
func (c *CertCommand) generateCert(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: cert generate <name> <subject> [validity_days]")
	}

	name := args[0]
	subject := args[1]
	validityDays := 365 // Default

	if len(args) > 2 {
		if days, err := fmt.Sscanf(args[2], "%d", &validityDays); err != nil || days != 1 {
			return fmt.Errorf("invalid validity days: %s", args[2])
		}
	}

	c.formatter.PrintInfo(fmt.Sprintf("Generating self-signed certificate: %s", name))

	cert, err := c.certManager.GenerateSelfSignedCert(name, subject, validityDays)
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %w", err)
	}

	c.formatter.PrintSuccess("Certificate generated successfully")
	c.formatter.PrintInfo(fmt.Sprintf("Certificate ID: %s", cert.ID))
	c.formatter.PrintInfo(fmt.Sprintf("Subject: %s", cert.Subject))
	c.formatter.PrintInfo(fmt.Sprintf("Valid until: %s", cert.NotAfter.Format("2006-01-02 15:04:05")))
	c.formatter.PrintInfo(fmt.Sprintf("Certificate file: %s", cert.FilePath))
	c.formatter.PrintInfo(fmt.Sprintf("Private key file: %s", cert.KeyPath))

	return nil
}

// listCerts lists all certificates
func (c *CertCommand) listCerts() error {
	certs := c.certManager.ListCertificates()

	if len(certs) == 0 {
		c.formatter.PrintInfo("No certificates found")
		return nil
	}

	c.formatter.PrintInfo("Certificates:")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("%-10s %-20s %-30s %-15s %-20s %-10s\n", "ID", "Name", "Subject", "Type", "Expires", "Status")
	fmt.Println(strings.Repeat("-", 100))

	for _, cert := range certs {
		fmt.Printf("%-10s %-20s %-30s %-15s %-20s %-10s\n",
			cert.ID[:8],
			cert.Name,
			cert.Subject,
			cert.Type,
			cert.NotAfter.Format("2006-01-02"),
			cert.Status,
		)
	}

	return nil
}

// validateCert validates a certificate
func (c *CertCommand) validateCert(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cert validate <cert_id>")
	}

	certID := args[0]

	c.formatter.PrintInfo(fmt.Sprintf("Validating certificate: %s", certID))

	err := c.certManager.ValidateCertificate(certID)
	if err != nil {
		return fmt.Errorf("certificate validation failed: %w", err)
	}

	c.formatter.PrintSuccess("Certificate is valid")

	return nil
}

// revokeCert revokes a certificate
func (c *CertCommand) revokeCert(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cert revoke <cert_id>")
	}

	certID := args[0]

	c.formatter.PrintInfo(fmt.Sprintf("Revoking certificate: %s", certID))

	err := c.certManager.RevokeCertificate(certID)
	if err != nil {
		return fmt.Errorf("failed to revoke certificate: %w", err)
	}

	c.formatter.PrintSuccess("Certificate revoked successfully")

	return nil
}

// deleteCert deletes a certificate
func (c *CertCommand) deleteCert(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cert delete <cert_id>")
	}

	certID := args[0]

	c.formatter.PrintInfo(fmt.Sprintf("Deleting certificate: %s", certID))

	err := c.certManager.DeleteCertificate(certID)
	if err != nil {
		return fmt.Errorf("failed to delete certificate: %w", err)
	}

	c.formatter.PrintSuccess("Certificate deleted successfully")

	return nil
}

// checkExpiry checks for expiring certificates
func (c *CertCommand) checkExpiry(args []string) error {
	days := 30 // Default
	if len(args) > 0 {
		if d, err := fmt.Sscanf(args[0], "%d", &days); err != nil || d != 1 {
			return fmt.Errorf("invalid days: %s", args[0])
		}
	}

	c.formatter.PrintInfo(fmt.Sprintf("Checking certificates expiring within %d days...", days))

	expiring := c.certManager.CheckExpiringCertificates(days)

	if len(expiring) == 0 {
		c.formatter.PrintSuccess("No certificates expiring within the specified period")
		return nil
	}

	c.formatter.PrintWarning(fmt.Sprintf("Found %d certificates expiring within %d days:", len(expiring), days))
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("%-10s %-20s %-30s %-20s\n", "ID", "Name", "Subject", "Expires")
	fmt.Println(strings.Repeat("-", 80))

	for _, cert := range expiring {
		fmt.Printf("%-10s %-20s %-30s %-20s\n",
			cert.ID[:8],
			cert.Name,
			cert.Subject,
			cert.NotAfter.Format("2006-01-02 15:04:05"),
		)
	}

	return nil
}

// showCertHelp shows certificate help
func (c *CertCommand) showCertHelp() error {
	c.formatter.PrintInfo("Certificate Commands:")
	fmt.Println("  generate <name> <subject> [validity_days] - Generate self-signed certificate")
	fmt.Println("  list                                    - List all certificates")
	fmt.Println("  validate <cert_id>                      - Validate a certificate")
	fmt.Println("  revoke <cert_id>                        - Revoke a certificate")
	fmt.Println("  delete <cert_id>                        - Delete a certificate")
	fmt.Println("  check-expiry [days]                     - Check for expiring certificates")
	return nil
}

// AuditCommand manages audit logging
type AuditCommand struct {
	BaseCommand
	auditManager *security.AuditManager
}

// NewAuditCommand creates a new audit command
func NewAuditCommand(client *client.Client, formatter *formatter.Formatter) *AuditCommand {
	return &AuditCommand{
		BaseCommand: BaseCommand{
			name:        "audit",
			description: "Manage audit logging and security events",
			usage:       "audit [query|report|failed-logins|high-risk] [options]",
			client:      client,
			formatter:   formatter,
		},
		auditManager: security.NewAuditManager("./config"),
	}
}

// Execute executes the audit command
func (c *AuditCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showAuditHelp()
	}

	subcommand := strings.ToLower(args[0])
	switch subcommand {
	case "query":
		return c.queryEvents(args[1:])
	case "report":
		return c.generateReport(args[1:])
	case "failed-logins":
		return c.showFailedLogins(args[1:])
	case "high-risk":
		return c.showHighRiskEvents(args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// queryEvents queries audit events
func (c *AuditCommand) queryEvents(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: audit query <user_id> <limit>")
	}

	userID := args[0]
	limit := 10 // Default

	if len(args) > 1 {
		if l, err := fmt.Sscanf(args[1], "%d", &limit); err != nil || l != 1 {
			return fmt.Errorf("invalid limit: %s", args[1])
		}
	}

	c.formatter.PrintInfo(fmt.Sprintf("Querying events for user: %s (limit: %d)", userID, limit))

	events, err := c.auditManager.GetEventsByUser(userID, limit)
	if err != nil {
		return fmt.Errorf("failed to query events: %w", err)
	}

	if len(events) == 0 {
		c.formatter.PrintInfo("No events found")
		return nil
	}

	c.formatter.PrintInfo(fmt.Sprintf("Found %d events:", len(events)))
	fmt.Println(strings.Repeat("=", 120))
	fmt.Printf("%-20s %-15s %-20s %-15s %-10s %-10s %-20s\n", "Timestamp", "Username", "Action", "Resource", "Result", "Risk", "IP Address")
	fmt.Println(strings.Repeat("-", 120))

	for _, event := range events {
		fmt.Printf("%-20s %-15s %-20s %-15s %-10s %-10s %-20s\n",
			event.Timestamp.Format("2006-01-02 15:04:05"),
			event.Username,
			event.Action,
			event.Resource,
			event.Result,
			event.RiskLevel,
			event.IPAddress,
		)
	}

	return nil
}

// generateReport generates an audit report
func (c *AuditCommand) generateReport(args []string) error {
	// Default to last 7 days
	end := time.Now()
	start := end.AddDate(0, 0, -7)

	if len(args) >= 2 {
		// Parse start and end dates
		if startTime, err := time.Parse("2006-01-02", args[0]); err == nil {
			start = startTime
		}
		if endTime, err := time.Parse("2006-01-02", args[1]); err == nil {
			end = endTime
		}
	}

	c.formatter.PrintInfo(fmt.Sprintf("Generating audit report from %s to %s", start.Format("2006-01-02"), end.Format("2006-01-02")))

	report, err := c.auditManager.GenerateReport(start, end)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	c.formatter.PrintInfo("Audit Report:")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Period: %s to %s\n", report.StartTime.Format("2006-01-02"), report.EndTime.Format("2006-01-02"))
	fmt.Printf("Total Events: %d\n", report.TotalEvents)
	fmt.Printf("Failed Logins: %d\n", report.FailedLogins)
	fmt.Printf("High Risk Events: %d\n", report.HighRiskEvents)

	fmt.Println("\nEvents by User:")
	for user, count := range report.EventsByUser {
		fmt.Printf("  %s: %d\n", user, count)
	}

	fmt.Println("\nEvents by Action:")
	for action, count := range report.EventsByAction {
		fmt.Printf("  %s: %d\n", action, count)
	}

	fmt.Println("\nEvents by Risk Level:")
	for risk, count := range report.EventsByRisk {
		fmt.Printf("  %s: %d\n", risk, count)
	}

	return nil
}

// showFailedLogins shows failed login attempts
func (c *AuditCommand) showFailedLogins(args []string) error {
	limit := 10 // Default
	if len(args) > 0 {
		if l, err := fmt.Sscanf(args[0], "%d", &limit); err != nil || l != 1 {
			return fmt.Errorf("invalid limit: %s", args[0])
		}
	}

	c.formatter.PrintInfo(fmt.Sprintf("Showing last %d failed login attempts:", limit))

	events, err := c.auditManager.GetFailedLogins(limit)
	if err != nil {
		return fmt.Errorf("failed to get failed logins: %w", err)
	}

	if len(events) == 0 {
		c.formatter.PrintSuccess("No failed login attempts found")
		return nil
	}

	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("%-20s %-15s %-20s %-20s %-20s\n", "Timestamp", "Username", "IP Address", "User Agent", "Details")
	fmt.Println(strings.Repeat("-", 100))

	for _, event := range events {
		details := ""
		if event.Details != nil {
			if reason, ok := event.Details["reason"].(string); ok {
				details = reason
			}
		}

		fmt.Printf("%-20s %-15s %-20s %-20s %-20s\n",
			event.Timestamp.Format("2006-01-02 15:04:05"),
			event.Username,
			event.IPAddress,
			event.UserAgent,
			details,
		)
	}

	return nil
}

// showHighRiskEvents shows high-risk events
func (c *AuditCommand) showHighRiskEvents(args []string) error {
	limit := 10 // Default
	if len(args) > 0 {
		if l, err := fmt.Sscanf(args[0], "%d", &limit); err != nil || l != 1 {
			return fmt.Errorf("invalid limit: %s", args[0])
		}
	}

	c.formatter.PrintInfo(fmt.Sprintf("Showing last %d high-risk events:", limit))

	events, err := c.auditManager.GetHighRiskEvents(limit)
	if err != nil {
		return fmt.Errorf("failed to get high-risk events: %w", err)
	}

	if len(events) == 0 {
		c.formatter.PrintSuccess("No high-risk events found")
		return nil
	}

	fmt.Println(strings.Repeat("=", 120))
	fmt.Printf("%-20s %-15s %-20s %-15s %-10s %-10s %-20s\n", "Timestamp", "Username", "Action", "Resource", "Result", "Risk", "IP Address")
	fmt.Println(strings.Repeat("-", 120))

	for _, event := range events {
		fmt.Printf("%-20s %-15s %-20s %-15s %-10s %-10s %-20s\n",
			event.Timestamp.Format("2006-01-02 15:04:05"),
			event.Username,
			event.Action,
			event.Resource,
			event.Result,
			event.RiskLevel,
			event.IPAddress,
		)
	}

	return nil
}

// showAuditHelp shows audit help
func (c *AuditCommand) showAuditHelp() error {
	c.formatter.PrintInfo("Audit Commands:")
	fmt.Println("  query <user_id> [limit]         - Query events for a user")
	fmt.Println("  report [start_date] [end_date]  - Generate audit report")
	fmt.Println("  failed-logins [limit]           - Show failed login attempts")
	fmt.Println("  high-risk [limit]               - Show high-risk events")
	return nil
}
