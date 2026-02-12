# Security Policy

For the complete security policy, including incident response procedures and security audit results, see [docs/SECURITY.md](docs/SECURITY.md).

## Reporting a Vulnerability

**DO NOT** open a public GitHub issue for security vulnerabilities.

### Preferred Method: GitHub Security Advisories

1. Go to the [Security Advisories](https://github.com/opd-ai/venture/security/advisories) page
2. Click "Report a vulnerability"
3. Fill in the vulnerability details
4. Submit for private review

### Alternative: Email

Email security@venture-rpg.com with:
- Description of the vulnerability
- Steps to reproduce
- Affected versions
- Potential impact

### Response Timeline

| Severity | Response Time | Fix Timeline |
|----------|---------------|--------------|
| Critical | 24 hours | 7 days |
| High | 48 hours | 30 days |
| Medium | 7 days | 60 days |
| Low | 14 days | 90 days |

## Supported Versions

| Version | Supported |
|---------|-----------|
| 1.x.x | ✅ Active |
| < 1.0 | ❌ End of life |

## Security Updates

- **Dependabot**: Enabled for automatic dependency updates
- **CodeQL**: Enabled for static analysis
- **gosec**: Run monthly and before releases

See [docs/SECURITY.md](docs/SECURITY.md) for complete details.
