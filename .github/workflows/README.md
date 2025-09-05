# GitHub Actions Workflows

This directory contains the CI/CD pipeline configurations for PeerVault.

## Workflows Overview

### 1. Main CI Pipeline (`ci.yml`)

The primary continuous integration pipeline that runs on every push and pull request.

**Features:**

- ✅ Lint and format checking
- ✅ Unit tests across multiple Go versions
- ✅ Integration tests
- ✅ Fuzz testing
- ✅ Basic security scanning
- ✅ Multi-platform builds (Linux, Windows, macOS)
- ✅ Docker image building and testing
- ✅ Performance benchmarks
- ✅ Code quality metrics
- ✅ Documentation validation

**Trigger:** Push to `main`/`develop`, Pull requests

### 2. Security Pipeline (`security.yml`)

Comprehensive security scanning and compliance checking pipeline.

**Features:**

- 🔒 Vulnerability scanning (govulncheck, gosec, semgrep)
- 🔒 Secrets detection
- 🔒 Compliance checking (SOC 2, GDPR, ISO 27001)
- 🔒 Security policy validation
- 🔒 Container security scanning (Trivy)
- 🔒 Security integration tests
- 🔒 Custom PeerVault security tools

**Trigger:** Push to `main`/`develop`, Pull requests, Daily schedule (2 AM UTC), Manual dispatch

### 3. Development Security Checks (`security-dev.yml`)

Quick security validation for development changes.

**Features:**

- 🔒 Security module compilation testing
- 🔒 Security unit tests
- 🔒 Security tools functionality testing
- 🔒 Security policy validation
- 🔒 Security documentation checks

**Trigger:** Pull requests affecting security files, Manual dispatch

## Security Features (Milestone 8)

The security pipeline implements comprehensive security features:

### 🔒 **Vulnerability Scanning**

- **govulncheck**: Go vulnerability database scanning
- **gosec**: Static analysis security scanner
- **semgrep**: Multi-language security scanner
- **detect-secrets**: Secrets and credentials detection

### 🔒 **Compliance Checking**

- **SOC 2 Type II**: Service organization controls
- **GDPR**: General Data Protection Regulation
- **ISO 27001**: Information security management
- **HIPAA**: Health Insurance Portability and Accountability Act
- **PCI DSS**: Payment Card Industry Data Security Standard

### 🔒 **Security Infrastructure**

- **RBAC**: Role-Based Access Control system
- **Audit Logging**: Comprehensive security event logging
- **Data Privacy**: Privacy controls and data protection
- **PKI**: Public Key Infrastructure and certificate management
- **Security Policies**: Access control and data classification policies

### 🔒 **Container Security**

- **Trivy**: Container vulnerability scanning
- **Docker Security**: Multi-stage build security
- **Image Scanning**: Base image vulnerability assessment

## Local Security Testing

### Using Bash Script (Linux/macOS)

```bash
# Run all security checks
./scripts/security-check.sh

# Run specific checks
./scripts/security-check.sh --vulnerability
./scripts/security-check.sh --compliance
./scripts/security-check.sh --test

# Install security tools
./scripts/security-check.sh --install-tools
```

### Using PowerShell Script (Windows)

```powershell
# Run all security checks
.\scripts\security-check.ps1

# Run specific checks
.\scripts\security-check.ps1 -Vulnerability
.\scripts\security-check.ps1 -Compliance
.\scripts\security-check.ps1 -Test

# Install security tools
.\scripts\security-check.ps1 -InstallTools
```

## Pipeline Status

### Critical Jobs (Must Pass)

- ✅ **Unit Tests**: Core functionality testing
- ✅ **Security**: Basic security scanning
- ✅ **Build**: Multi-platform binary building
- ✅ **Docker**: Container building and testing
- ✅ **Benchmarks**: Performance testing

### Non-Critical Jobs (Warnings Only)

- ⚠️ **Lint**: Code formatting and style (warnings only)
- ⚠️ **Integration Tests**: Application logic testing (warnings only)
- ⚠️ **Quality**: Code quality metrics (warnings only)
- ⚠️ **Docs**: Documentation validation (warnings only)

## Security Reports

The security pipeline generates comprehensive reports:

### Vulnerability Reports

- `security-reports/govulncheck-report.json`: Go vulnerability scan results
- `security-reports/gosec-report.json`: Static analysis security issues
- `security-reports/semgrep-report.json`: Multi-language security scan
- `security-reports/secrets-baseline.json`: Secrets detection results
- `security-reports/custom-security-report.json`: Custom PeerVault security scan

### Compliance Reports

- `compliance-reports/soc2-compliance.json`: SOC 2 compliance assessment
- `compliance-reports/gdpr-compliance.json`: GDPR compliance assessment
- `compliance-reports/iso27001-compliance.json`: ISO 27001 compliance assessment

### Container Security Reports

- `container-security-reports/trivy-node-report.json`: Node container vulnerabilities
- `container-security-reports/trivy-demo-report.json`: Demo container vulnerabilities

## Security Tools Integration

### Custom Security Tools

- **Security Scanner** (`security/audit/scanner.go`): Custom vulnerability scanning
- **Compliance Auditor** (`security/audit/compliance.go`): Compliance assessment
- **Security Scripts** (`security/tools/`): Automated security testing

### External Security Tools

- **govulncheck**: Go vulnerability database
- **gosec**: Go security scanner
- **semgrep**: Multi-language security scanner
- **Trivy**: Container security scanner
- **detect-secrets**: Secrets detection

## Configuration

### Environment Variables

- `GO_VERSION`: Go version for builds (default: 1.24.4)
- `CGO_ENABLED`: CGO support for tests (default: 1)

### Security Configuration

- Security policies: `security/policies/`
- Security tools: `security/tools/`
- Security documentation: `security/README.md`

## Troubleshooting

### Common Issues

1. **Security Scan Failures**
   - Check for high/critical vulnerabilities
   - Review security reports for details
   - Update dependencies if needed

2. **Compliance Check Failures**
   - Review compliance reports
   - Ensure security policies are in place
   - Check security documentation completeness

3. **Container Security Issues**
   - Update base images
   - Review container security reports
   - Check for known vulnerabilities

### Getting Help

1. **Check Pipeline Logs**: Review detailed logs in GitHub Actions
2. **Run Local Tests**: Use security check scripts locally
3. **Review Reports**: Examine generated security and compliance reports
4. **Security Documentation**: Check `security/README.md` for details

## Best Practices

### Development

1. **Run Local Security Checks**: Use security scripts before pushing
2. **Review Security Reports**: Check generated reports for issues
3. **Update Dependencies**: Keep dependencies updated for security
4. **Follow Security Policies**: Adhere to defined security policies

### CI/CD

1. **Monitor Pipeline Status**: Check pipeline status regularly
2. **Review Security Reports**: Examine security scan results
3. **Address Critical Issues**: Fix high/critical security issues immediately
4. **Maintain Compliance**: Ensure compliance requirements are met

## Security Milestone Status

**Milestone 8 — Security Hardening and Compliance (P7)** ✅ **COMPLETE**

- ✅ Security audit and penetration testing
- ✅ Access control and authorization (RBAC, ACLs)
- ✅ Data privacy and compliance features
- ✅ Certificate management and PKI
- ✅ Security policies and documentation
- ✅ Custom security tools and scripts
- ✅ CI/CD pipeline integration
- ✅ Local development tools

The PeerVault system now has enterprise-grade security, compliance, and CI/CD integration.
