# API Compatibility Policy

## Semantic Versioning

Venture follows [Semantic Versioning 2.0.0](https://semver.org/) with the format **MAJOR.MINOR.PATCH**:

- **MAJOR** (e.g., 1.x.x → 2.x.x): Incompatible API changes or breaking changes to:
  - Network protocol format
  - Save file format (not backward-compatible)
  - CLI flag behavior (removal or semantic change)
  - Public Go package APIs

- **MINOR** (e.g., 1.0.x → 1.1.x): New functionality added in a backward-compatible manner:
  - New CLI flags
  - New network message types (additive)
  - New save file fields (with defaults)
  - New game features

- **PATCH** (e.g., 1.0.0 → 1.0.1): Backward-compatible bug fixes:
  - Security patches
  - Performance improvements
  - Bug fixes that don't change expected behavior

## Public API Surface

The following are considered part of Venture's public API and are subject to compatibility guarantees:

### 1. CLI Flags

All command-line flags documented in `--help` output:

```
Client: venture-client --help
Server: venture-server --help
```

**Stability:** Flags are stable within a MAJOR version. Deprecated flags continue to work for 2 MINOR versions with deprecation warnings before removal.

### 2. Network Protocol

The client-server network protocol (message formats, handshake, state sync):

- Protocol version is embedded in handshake messages
- Clients and servers within the same MAJOR version are compatible
- Minor protocol extensions are backward-compatible

### 3. Save File Format

Save files created by one version can be loaded by:

- Same MAJOR version (guaranteed)
- Future MAJOR versions (via migration support, best effort)
- Previous MAJOR versions (not guaranteed)

Save format version is stored in `version` field of save files. The `pkg/saveload` package includes migrators for upgrading old save files.

### 4. Configuration Files

Configuration file formats (JSON/YAML):

- New fields are optional with sensible defaults
- Removed fields are ignored with deprecation warnings
- Field type changes are a MAJOR version change

### 5. Mod API

The modding interface (`pkg/modding`):

- Mod manifest format stable within MAJOR version
- Rule types and effects stable within MAJOR version
- Sandbox security guarantees maintained

## Deprecation Policy

1. **Announcement**: Deprecated features are documented in release notes and CHANGELOG.md
2. **Warning Period**: Deprecated features show runtime warnings for 2 MINOR versions
3. **Removal**: Deprecated features may be removed in the next MAJOR version

Example timeline:
- v1.2.0: Feature X deprecated, warning added
- v1.3.0: Warning continues
- v1.4.0: Warning continues (minimum 2 MINOR versions)
- v2.0.0: Feature X may be removed

## Breaking Changes

When a breaking change is necessary:

1. Document in CHANGELOG.md under "BREAKING CHANGES" section
2. Provide migration guide in docs/UPGRADE_GUIDE.md
3. Include automated migration tools where possible
4. Bump MAJOR version number

## Pre-Release Versions

Pre-release versions (Alpha, Beta, Development) may have unstable APIs:

- `1.0.0-alpha.1`: Early development, API may change frequently
- `1.0.0-beta.1`: Feature complete, API stabilizing
- `1.0.0-rc.1`: Release candidate, API frozen except for critical fixes
- `1.0.0`: Production release, API stability guarantees begin

## Version Checking

### CLI Version Flag

```bash
venture-client --version
venture-server --version
```

Output format:
```
Venture 1.0.0 Production (linux/amd64, go1.24.5)
```

### Programmatic Access

```go
import "github.com/opd-ai/venture/pkg/version"

// Get version string
v := version.Version      // "1.0.0"
full := version.FullVersion  // "1.0.0 Production"

// Get version components
major := version.Major    // 1
minor := version.Minor    // 0
patch := version.Patch    // 0

// Get build info including platform
info := version.BuildInfo() // "Venture 1.0.0 Production (linux/amd64, go1.24.5)"
```

## Compatibility Matrix

| Client Version | Server Version | Compatible? |
|----------------|----------------|-------------|
| 1.0.x          | 1.0.x          | ✅ Yes      |
| 1.0.x          | 1.1.x          | ✅ Yes      |
| 1.1.x          | 1.0.x          | ⚠️ Partial* |
| 1.x.x          | 2.x.x          | ❌ No       |

*Partial compatibility: Client may have features server doesn't support.

## Questions?

- See [CHANGELOG.md](CHANGELOG.md) for version history
- See [UPGRADE_GUIDE.md](UPGRADE_GUIDE.md) for migration instructions
- File issues on GitHub for compatibility concerns
