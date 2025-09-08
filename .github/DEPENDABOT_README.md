# Dependabot Configuration

This repository uses Dependabot to automatically monitor and update dependencies, with a focus on security and comprehensive codebase analysis.

## 📁 Configuration Files

- **`.github/dependabot.yml`** - Main Dependabot configuration for regular updates
- **`.github/dependabot-security.yml`** - Security-focused Dependabot configuration
- **`.github/workflows/dependabot-security.yml`** - GitHub Actions workflow for security analysis
- **`.github/ISSUE_TEMPLATE/dependabot-security.md`** - Issue template for security alerts
- **`.github/security.yml`** - GitHub security features configuration

## 🔧 What Dependabot Monitors

### Package Ecosystems

- **Go modules** (main project, plugins, proto)
- **GitHub Actions**
- **Docker**
- **npm** (if present)
- **pip** (if present)

### Update Types

- **Regular updates** - Minor and patch versions
- **Security updates** - Critical security patches
- **Major updates** - Currently ignored to prevent breaking changes

## 📅 Schedule

- **Regular updates:** Weekly (Mondays at 09:00)
- **Security updates:** Daily (06:00)
- **Security analysis:** Daily (02:00)

## 🏷️ Labels and Organization

### Labels Used

- `dependencies` - All dependency updates
- `security` - Security-related updates
- `go` - Go-specific updates
- `github-actions` - GitHub Actions updates
- `docker` - Docker updates
- `plugins` - Plugin-specific updates
- `proto` - Protocol buffer updates
- `high-priority` - Critical updates
- `automated` - Automated processes

### Grouping

- **Go dependencies** - Grouped by minor/patch updates
- **GitHub Actions** - Grouped by minor/patch updates
- **Docker** - Grouped by minor/patch updates

## 🚨 Security Features

### Automatic Security Scanning

- **Go vulnerability check** (govulncheck)
- **Go security scan** (gosec)
- **Filesystem vulnerability scan** (Trivy)
- **Dependency analysis**

### Issue Creation

- Automatic issue creation for security vulnerabilities
- Detailed security alert templates
- Priority-based labeling
- Assignee assignment

### Monitoring

- Daily security analysis
- Weekly dependency reports
- Real-time vulnerability alerts
- GitHub Security tab integration

## 📊 Pull Request Management

### Limits

- **Main Go modules:** 15 PRs
- **Plugin Go modules:** 5 PRs
- **Proto Go modules:** 5 PRs
- **GitHub Actions:** 10 PRs
- **Docker:** 10 PRs
- **npm/pip:** 5 PRs each

### Reviewers and Assignees

- **Primary assignee:** Skpow1234
- **Reviewer:** Skpow1234
- **Auto-assignment** enabled

## 🔄 Workflow Integration

### GitHub Actions

- **Security analysis workflow** runs daily
- **Dependency monitoring** with status checks
- **Automated issue creation** for critical vulnerabilities
- **Security summary reports**

### Notifications

- **Email notifications** for security alerts
- **Issue creation** for critical vulnerabilities
- **Pull request reviews** for dependency updates

## 🛠️ Manual Operations

### Force Security Scan

```bash
# Trigger manual security analysis
gh workflow run dependabot-security.yml
```

### Check Dependabot Status

```bash
# View Dependabot alerts
gh api repos/:owner/:repo/dependabot/alerts

# View Dependabot secrets
gh api repos/:owner/:repo/dependabot/secrets
```

### Update Dependencies Manually

```bash
# Update Go dependencies
go get -u ./...

# Update specific dependency
go get -u github.com/package/name@latest
```

## 📈 Monitoring and Reporting

### Daily Reports

- Security vulnerability scan results
- Dependency status updates
- Outdated dependency lists

### Weekly Reports

- Comprehensive dependency analysis
- Update recommendations
- Security posture assessment

### Alerts and Notifications

- Critical vulnerability alerts
- High-priority security issues
- Dependency update notifications

## 🔒 Security Best Practices

### Dependency Management

- Regular security updates
- Vulnerability monitoring
- Automated testing after updates
- Rollback procedures

### Code Security

- Security scanning integration
- Vulnerability assessment
- Risk-based prioritization
- Compliance monitoring

## 📞 Support and Troubleshooting

### Common Issues

1. **Dependabot not running** - Check repository settings
2. **Security alerts not appearing** - Verify security features are enabled
3. **Update failures** - Review dependency compatibility
4. **False positives** - Configure ignore rules

### Configuration Updates

- Modify `.github/dependabot.yml` for regular updates
- Modify `.github/dependabot-security.yml` for security updates
- Update workflow files for automation changes

### Contact

- **Repository maintainer:** Skpow1234
- **Security issues:** Use the security alert template
- **General questions:** Create a regular issue

## 📚 Additional Resources

- [Dependabot Documentation](https://docs.github.com/en/code-security/dependabot)
- [GitHub Security Features](https://docs.github.com/en/code-security)
- [Go Security Best Practices](https://golang.org/doc/security)
- [Docker Security Scanning](https://docs.docker.com/engine/security/)

---

**Last Updated:** $(date)
**Configuration Version:** 2.0
**Maintainer:** Skpow1234
