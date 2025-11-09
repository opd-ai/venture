# CodeQL Security Scanning Setup

This directory contains the CodeQL security scanning workflow and supporting scripts for the Venture project.

## Overview

The CodeQL security scanning system performs automated security analysis on the Go codebase and publishes results to both GitHub's Security tab and as HTML reports on GitHub Pages.

## Files

### `.github/workflows/codeql.yml`
Main CodeQL workflow that:
- Runs on every push to `main`, pull requests, and weekly on Mondays
- Performs CodeQL analysis using `security-extended` and `security-and-quality` query packs
- Generates SARIF results uploaded to GitHub Security tab
- Converts SARIF to HTML for human-readable reports
- Uploads both SARIF and HTML reports as artifacts

### `.github/scripts/sarif-to-html.sh`
Bash script that converts CodeQL SARIF output to styled HTML reports:
- Uses `jq` for JSON parsing (with Python fallback)
- Generates interactive HTML with filtering by severity
- Matches the Venture dark theme styling
- Creates user-friendly visualizations of security findings

### `.github/workflows/pages.yml` (updated)
GitHub Pages deployment workflow enhanced to:
- Download latest CodeQL HTML reports from workflow artifacts
- Create a security index page at `/security/`
- Deploy reports alongside the game to GitHub Pages
- Maintain existing game deployment without conflicts

## Usage

### Viewing Reports

After workflows run, security reports are available at:
- **GitHub Security Tab**: https://github.com/opd-ai/venture/security/code-scanning
- **GitHub Pages**: https://opd-ai.github.io/venture/security/
- **Workflow Artifacts**: Download from Actions tab

### Running Manually

Trigger CodeQL analysis manually:
```bash
# Via GitHub UI: Go to Actions → CodeQL Security Analysis → Run workflow

# Or using GitHub CLI:
gh workflow run codeql.yml
```

### Local Testing

Test SARIF to HTML conversion locally:
```bash
# Install jq (optional but recommended)
sudo apt-get install jq

# Run conversion
.github/scripts/sarif-to-html.sh <sarif-dir> <output-dir>

# Example:
.github/scripts/sarif-to-html.sh ./sarif-results ./html-reports
```

## Configuration

### CodeQL Settings

The workflow ignores certain paths to reduce noise:
- `examples/**` - Example code not part of production
- `cmd/*test/**` - Test utilities

Query packs used:
- `security-extended` - Extended security queries
- `security-and-quality` - Security and code quality queries

### Customization

To modify CodeQL settings, edit `.github/workflows/codeql.yml`:
- Change `cron` schedule for different scan frequency
- Adjust `paths-ignore` to exclude additional directories
- Modify query packs in the `queries` field

## Architecture

```
┌─────────────────────┐
│ Push to main / PR   │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ CodeQL Workflow     │
│  - Build project    │
│  - Run analysis     │
│  - Generate SARIF   │
│  - Convert to HTML  │
└──────────┬──────────┘
           │
           ├─────────────────────┐
           │                     │
           ▼                     ▼
┌─────────────────────┐  ┌─────────────────────┐
│ GitHub Security Tab │  │ Workflow Artifacts  │
│ (SARIF upload)      │  │ (HTML + SARIF)      │
└─────────────────────┘  └──────────┬──────────┘
                                    │
                                    ▼
                         ┌─────────────────────┐
                         │ Pages Workflow      │
                         │  - Download HTML    │
                         │  - Deploy to Pages  │
                         └──────────┬──────────┘
                                    │
                                    ▼
                         ┌─────────────────────┐
                         │ GitHub Pages        │
                         │ /security/          │
                         └─────────────────────┘
```

## Troubleshooting

### Workflow Fails

1. Check workflow logs in Actions tab
2. Verify Go version matches project requirements (1.24.5+)
3. Ensure build dependencies are installed correctly

### No HTML Reports on Pages

1. Verify CodeQL workflow completed successfully
2. Check that artifacts were uploaded (Actions → CodeQL workflow → Artifacts)
3. Ensure Pages workflow ran after CodeQL workflow
4. Check Pages deployment logs

### SARIF Conversion Fails

1. Verify SARIF files are in expected location
2. Check `jq` is available (or Python 3 as fallback)
3. Review script logs for detailed error messages

## Security Considerations

- CodeQL results may contain sensitive information about vulnerabilities
- Reports are public on GitHub Pages (same as the game)
- SARIF artifacts expire after 30 days
- Use GitHub Security tab for private vulnerability tracking

## Maintenance

### Regular Tasks

- Review weekly scan results
- Update CodeQL action versions when available
- Adjust query packs based on findings
- Archive critical findings for compliance

### Updates

When updating:
1. Test changes in a feature branch first
2. Verify both workflows complete successfully
3. Check HTML reports render correctly
4. Confirm no conflicts with existing Pages content

## Resources

- [CodeQL Documentation](https://codeql.github.com/docs/)
- [GitHub Code Scanning](https://docs.github.com/en/code-security/code-scanning)
- [SARIF Specification](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html)
- [GitHub Actions](https://docs.github.com/en/actions)
