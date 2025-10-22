# GitHub Actions CI/CD Pipeline

This repository uses GitHub Actions for automated testing, building, and releasing. The CI/CD pipeline is split into three separate workflows for modularity and clarity.

## Workflows Overview

### 1. Test Workflow (`test.yml`)
**Purpose:** Runs automated tests and code quality checks on pull requests and pushes to main.

**Triggers:**
- Pull requests targeting the `main` branch
- Direct pushes to the `main` branch

**What it does:**
- Tests on Go 1.23.x and 1.24.x
- Downloads dependencies
- Runs all tests in `./pkg/...` with test tag
- Runs `go vet` for code quality checks
- Checks code formatting with `gofmt`

**How to use:**
- Automatically runs when you open a PR or push to main
- Fix any test failures or formatting issues before merging

### 2. Build Workflow (`build.yml`)
**Purpose:** Builds cross-platform binaries for verification after code is merged to main.

**Triggers:**
- Pushes to the `main` branch

**What it does:**
- Builds the server binary for multiple platforms:
  - Linux (amd64, arm64)
  - macOS/Darwin (amd64, arm64)
  - Windows (amd64)
- Uploads build artifacts that are retained for 7 days
- Uses optimized build flags (`-ldflags="-s -w"`) to reduce binary size

**How to use:**
- Automatically runs after merging to main
- Download artifacts from the Actions tab to verify builds
- Artifacts are available for 7 days for testing

### 3. Release Workflow (`release.yml`)
**Purpose:** Creates releases with cross-platform binaries, supporting both nightly builds and versioned releases.

**Triggers:**
- **Nightly builds:** Scheduled daily at 00:00 UTC
- **Version releases:** Pushing a tag matching `v*.*.*` pattern (e.g., `v1.0.0`, `v2.1.3`)

**What it does:**
- Determines release type (nightly or version)
- For nightly builds:
  - Deletes existing nightly tag and release
  - Creates new nightly tag pointing to latest main
  - Marks as pre-release
- For version releases:
  - Creates stable release from the tag
  - Generates release notes from git commits
- Builds binaries for all platforms
- Creates compressed archives (`.tar.gz` for Unix, `.zip` for Windows)
- Attaches binaries to the GitHub release

## Usage Guide

### Running Tests Locally

Before pushing code, run tests locally to catch issues early:

```bash
# Run all tests
go test -tags test -v ./pkg/...

# Run code quality checks
go vet -tags test ./pkg/...

# Check formatting
gofmt -s -l .
```

### Creating a Nightly Build

Nightly builds run automatically every day at midnight UTC. No manual action is required.

To trigger a manual nightly build:
1. Go to the "Actions" tab in GitHub
2. Select the "Release" workflow
3. Click "Run workflow"
4. Select the main branch
5. Click "Run workflow"

**Note:** The nightly build will always use the latest commit on the main branch.

### Creating a Versioned Release

To create a versioned release (e.g., v1.0.0):

1. **Ensure main branch is ready for release:**
   ```bash
   git checkout main
   git pull origin main
   ```

2. **Create and push a version tag:**
   ```bash
   # Replace 1.0.0 with your version number
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

3. **The workflow will automatically:**
   - Detect the version tag
   - Build binaries for all platforms
   - Generate release notes from commits since the last tag
   - Create a GitHub release
   - Upload binaries as release assets

4. **Verify the release:**
   - Go to the "Releases" section of the repository
   - Check that the new release appears
   - Download and test the binaries

### Version Tag Pattern

The release workflow only triggers on tags matching the semantic versioning pattern:
- ✅ Valid: `v1.0.0`, `v2.1.3`, `v0.1.0`, `v10.5.2`
- ❌ Invalid: `1.0.0`, `release-1.0`, `v1.0`, `v1.0.0-beta`

### Downloading Release Artifacts

**From Releases:**
1. Go to the "Releases" section
2. Find your release (nightly or versioned)
3. Download the appropriate archive for your platform
4. Extract and run

**From Build Artifacts (temporary):**
1. Go to the "Actions" tab
2. Click on a completed "Build" workflow run
3. Scroll to "Artifacts"
4. Download the artifact for your platform
5. Artifacts expire after 7 days

## Platform-Specific Notes

### Linux
- Binaries: `venture-server-linux-amd64.tar.gz`, `venture-server-linux-arm64.tar.gz`
- Extract: `tar -xzf venture-server-linux-amd64.tar.gz`
- Run: `./venture-server-linux-amd64`

### macOS
- Binaries: `venture-server-darwin-amd64.tar.gz`, `venture-server-darwin-arm64.tar.gz`
- Extract: `tar -xzf venture-server-darwin-amd64.tar.gz`
- Run: `./venture-server-darwin-amd64`
- Note: You may need to grant execution permissions: `chmod +x venture-server-darwin-amd64`

### Windows
- Binary: `venture-server-windows-amd64.zip`
- Extract using Windows Explorer or `unzip`
- Run: `venture-server-windows-amd64.exe`

## Workflow Maintenance

### Updating Go Versions

To update the Go versions used in testing:
1. Edit `.github/workflows/test.yml`
2. Update the `go-version` array in the matrix strategy
3. Commit and push the changes

### Adding New Platforms

To add support for additional platforms:
1. Edit `.github/workflows/build.yml` and `.github/workflows/release.yml`
2. Add new entries to the matrix (build.yml) or platforms array (release.yml)
3. Use the format: `goos/goarch` (e.g., `freebsd/amd64`)

### Changing Release Schedule

To change the nightly build schedule:
1. Edit `.github/workflows/release.yml`
2. Update the `cron` expression in the schedule trigger
3. Use [crontab.guru](https://crontab.guru) to help with cron syntax

## Troubleshooting

### Test Workflow Fails
- Check the test output in the Actions tab
- Run tests locally: `go test -tags test -v ./pkg/...`
- Fix failing tests or code quality issues
- Push the fixes

### Build Workflow Fails
- Check if the code compiles locally for the failing platform
- Test cross-compilation: `GOOS=linux GOARCH=amd64 go build -tags test ./cmd/server`
- Ensure all dependencies are properly declared in `go.mod`

### Release Workflow Fails
- Check that you have proper permissions in the repository
- Verify the tag follows the correct pattern (`v*.*.*`)
- Check the workflow logs for specific errors
- Ensure the main branch is in a buildable state

### Nightly Release Not Created
- Verify the scheduled trigger is enabled (GitHub may disable scheduled workflows in inactive repos)
- Check that the main branch has recent commits
- Review workflow logs for any errors

## Security Notes

- All workflows use `GITHUB_TOKEN` which is automatically provided by GitHub
- No manual secret configuration is required
- The `GITHUB_TOKEN` has limited permissions scoped to the repository
- Workflows run in isolated environments with no persistent state

## Further Reading

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Documentation](https://golang.org/doc/)
- [Semantic Versioning](https://semver.org/)
