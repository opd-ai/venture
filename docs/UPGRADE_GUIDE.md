# Upgrade Guide

This guide covers upgrading Venture between versions, including save file migrations, breaking changes, and rollback procedures.

## Table of Contents

1. [Version Compatibility](#version-compatibility)
2. [Upgrading to v1.0.0](#upgrading-to-v100)
3. [Save File Migration](#save-file-migration)
4. [Breaking Changes](#breaking-changes)
5. [Configuration Changes](#configuration-changes)
6. [Rollback Procedures](#rollback-procedures)
7. [Troubleshooting](#troubleshooting)

---

## Version Compatibility

### Semantic Versioning

Venture follows [Semantic Versioning](https://semver.org/):

| Version Change | Compatibility |
|----------------|---------------|
| PATCH (1.0.0 → 1.0.1) | Fully backward-compatible |
| MINOR (1.0.x → 1.1.x) | Backward-compatible with new features |
| MAJOR (1.x.x → 2.x.x) | May include breaking changes |

### Client-Server Compatibility

| Client Version | Server Version | Status |
|----------------|----------------|--------|
| 1.0.x | 1.0.x | ✅ Compatible |
| 1.0.x | 1.1.x | ✅ Compatible (server newer) |
| 1.1.x | 1.0.x | ⚠️ May lack new features |
| 1.x.x | 2.x.x | ❌ Incompatible |

---

## Upgrading to v1.0.0

### From Development Versions (v8.0.0 - v10.0.0)

These were internal development versions. Upgrading to v1.0.0 is straightforward:

1. **Stop the server gracefully:**
   ```bash
   # If running as systemd service
   sudo systemctl stop venture-server
   
   # If running manually, use Ctrl+C for graceful shutdown
   ```

2. **Backup your data:**
   ```bash
   # Backup saves directory
   cp -r ~/.venture/saves ~/.venture/saves.backup
   
   # Backup configuration
   cp ~/.venture/config.json ~/.venture/config.json.backup
   ```

3. **Install the new version:**
   ```bash
   # Using Homebrew (macOS)
   brew upgrade venture
   
   # Using apt (Debian/Ubuntu)
   sudo apt update && sudo apt upgrade venture
   
   # Using dnf (Fedora/RHEL)
   sudo dnf upgrade venture
   
   # Manual binary replacement
   wget https://github.com/opd-ai/venture/releases/download/v1.0.0/venture-linux-amd64
   chmod +x venture-linux-amd64
   sudo mv venture-linux-amd64 /usr/local/bin/venture
   ```

4. **Verify the installation:**
   ```bash
   venture --version
   # Expected output: Venture 1.0.0 Production (linux/amd64, go1.24.5)
   ```

5. **Start the server:**
   ```bash
   sudo systemctl start venture-server
   # or
   venture-server --port 7777
   ```

### From Pre-1.0 Save Files (v0.9.x)

Save files from v0.9.x versions are automatically migrated:

```bash
# The game automatically detects and migrates older saves
venture-client --load my-save.json

# To explicitly migrate without loading:
venture-client --migrate my-save.json
```

Supported migration paths:
- v0.9.0 → v1.0.0
- v0.9.1 → v1.0.0
- v0.9.2 → v1.0.0
- v0.9.3 → v1.0.0

---

## Save File Migration

### Automatic Migration

Venture automatically migrates save files when loading:

1. Detects save file version from `version` field
2. Applies migration hooks for each version step
3. Updates `version` field to current (1.0.0)
4. Creates backup before overwriting

### Manual Migration

```bash
# Migrate a save file to the current version
venture-client --migrate path/to/save.json

# Migrate with verbose output
venture-client --migrate path/to/save.json --log-level debug
```

### Migration Safety

Before any migration:
- Original save is backed up to `save.json.bak`
- Checksum is validated (SHA256)
- Migration is atomic (write to temp file, then rename)

If migration fails:
- Original save is preserved
- Error is logged with details
- Backup can be restored manually

### Verifying Save Files

```bash
# Check save file integrity
venture-client --verify path/to/save.json

# Output includes:
# - Version: 1.0.0
# - Checksum: valid/invalid
# - Player data: valid/corrupted
# - World data: valid/corrupted
```

---

## Breaking Changes

### v1.0.0 Breaking Changes

**None** - v1.0.0 is the first stable release. All previous versions were development releases.

### Future Version Changes

Breaking changes in future MAJOR versions will be documented here with:
- Description of the change
- Required migration steps
- Compatibility period for deprecated features

---

## Configuration Changes

### v1.0.0 Configuration

New configuration options in v1.0.0:

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--high-latency` | bool | false | Enable high-latency mode (200-5000ms) |
| `--metrics-port` | int | 9090 | Port for Prometheus metrics endpoint |
| `--health-port` | int | 8080 | Port for health check endpoints |

### Deprecated Options

None in v1.0.0. Future deprecations will follow this process:

1. Deprecation warning logged for 2 MINOR versions
2. Documentation updated with migration path
3. Option removed in next MAJOR version

### Environment Variables

```bash
# Logging configuration
export LOG_LEVEL=info          # debug, info, warn, error
export LOG_FORMAT=json         # json, text

# Performance tuning
export VENTURE_SPRITE_CACHE_MB=400   # Sprite cache size in MB
export VENTURE_MAX_ENTITIES=5000     # Maximum entity count
```

---

## Rollback Procedures

### Rolling Back Binary

1. **Stop the current server:**
   ```bash
   sudo systemctl stop venture-server
   ```

2. **Restore previous binary:**
   ```bash
   # If you kept the old binary
   sudo cp /usr/local/bin/venture.old /usr/local/bin/venture
   
   # Or download the previous version
   wget https://github.com/opd-ai/venture/releases/download/v0.9.3/venture-linux-amd64
   sudo mv venture-linux-amd64 /usr/local/bin/venture
   ```

3. **Restore save files (if needed):**
   ```bash
   # Restore from backup created during upgrade
   cp ~/.venture/saves.backup/* ~/.venture/saves/
   ```

4. **Start the server:**
   ```bash
   sudo systemctl start venture-server
   ```

### Rolling Back Save Files

If a save file was migrated but you need the old version:

```bash
# Restore from automatic backup
cp ~/.venture/saves/my-save.json.bak ~/.venture/saves/my-save.json

# Verify restoration
venture-client --verify ~/.venture/saves/my-save.json
```

### Rolling Back Configuration

```bash
# Restore configuration backup
cp ~/.venture/config.json.backup ~/.venture/config.json

# Verify configuration
venture-server --config ~/.venture/config.json --validate
```

---

## Troubleshooting

### Common Upgrade Issues

#### "Unsupported save file version"

**Cause:** Save file version is too old for automatic migration.

**Solution:**
1. Check the save file version: `head -1 save.json`
2. If version < 0.9.0, manual migration is required
3. Contact support for pre-0.9.0 save files

#### "Checksum mismatch"

**Cause:** Save file was corrupted during upgrade.

**Solution:**
```bash
# Attempt automatic recovery
venture-client --load save.json --recover

# If recovery fails, restore from backup
cp save.json.bak save.json
```

#### "Protocol version mismatch"

**Cause:** Client and server versions are incompatible.

**Solution:**
1. Check both versions: `venture-client --version` and `venture-server --version`
2. Upgrade client to match server MAJOR version
3. Or downgrade server if client cannot be updated

#### "Configuration option not recognized"

**Cause:** Using deprecated or renamed configuration option.

**Solution:**
1. Run with `--help` to see current options
2. Check CHANGELOG.md for renamed options
3. Update configuration file to use new option names

### Getting Help

- **GitHub Issues:** [github.com/opd-ai/venture/issues](https://github.com/opd-ai/venture/issues)
- **Documentation:** See `docs/` directory for detailed guides
- **Runbooks:** See `docs/runbooks/` for operational procedures

### Reporting Upgrade Issues

When reporting upgrade issues, include:

1. Previous version: `venture --version` (before upgrade)
2. New version: `venture --version` (after upgrade)
3. Operating system: `uname -a`
4. Error message (full text)
5. Relevant log output: `tail -100 ~/.venture/logs/venture.log`

---

## Quick Reference

### Pre-Upgrade Checklist

- [ ] Read CHANGELOG.md for the target version
- [ ] Backup saves directory
- [ ] Backup configuration files
- [ ] Note current version number
- [ ] Verify sufficient disk space for backups
- [ ] Schedule maintenance window (for servers)

### Post-Upgrade Checklist

- [ ] Verify new version installed correctly
- [ ] Check save file migration success
- [ ] Verify configuration loaded correctly
- [ ] Test basic functionality (connect, play)
- [ ] Monitor logs for errors
- [ ] Verify metrics/health endpoints (if used)

### Emergency Contacts

For critical production issues:
- Security issues: See SECURITY.md for responsible disclosure
- Critical bugs: Open GitHub issue with "critical" label
