# Runbook: Save File Corruption Recovery

**Severity:** P2 (Data Loss Risk)  
**Symptoms:** "Failed to load save", corrupted data errors, server crashes on save/load  
**Owner:** Infrastructure Team  
**Last Updated:** 2026-01-07

---

## Overview

This runbook helps recover from save file corruption in Venture game servers. Save corruption can result from crashes during save operations, disk errors, or software bugs. The game includes built-in protection (checksums, backups) to minimize data loss.

**Recovery Time:** 5-30 minutes depending on backup availability  
**Data Loss:** Minimal if backups available, up to last save interval otherwise

---

## Initial Assessment (2 minutes)

### 1. Verify Save File Corruption

Check server logs for corruption errors:

```bash
# Check for save/load errors
grep -i "save\|load\|corrupt\|checksum" /var/log/venture-server.log | tail -20

# Common error messages:
# - "checksum mismatch" → File corrupted
# - "failed to deserialize" → Invalid JSON/binary data
# - "save file not found" → File deleted or permissions issue
# - "backup restored" → Automatic recovery attempted
```

### 2. Identify Affected Save Files

Save files are stored in `/var/lib/venture/saves/` (or configured location):

```bash
# List save files
ls -lh /var/lib/venture/saves/

# Expected files:
# - world.save          (primary save file)
# - world.save.sha256   (checksum for verification)
# - world.save.bak      (most recent backup)
# - autosave_*.save     (automatic backups)
```

### 3. Verify Checksum

The game automatically validates checksums on load. Manual verification:

```bash
cd /var/lib/venture/saves/

# Calculate current checksum
sha256sum world.save > world.save.sha256.new

# Compare with stored checksum
diff world.save.sha256 world.save.sha256.new

# If different: File is corrupted
# If same: File integrity OK (issue might be format/version)
```

---

## Diagnosis (5 minutes)

### Step 1: Determine Corruption Type

**Scenario A: Checksum Mismatch**
- File modified after save
- Disk error corrupted data
- Incomplete write (crash during save)

**Scenario B: Deserialize Error**
- Save format incompatible (version mismatch)
- JSON syntax error
- Binary corruption in specific field

**Scenario C: File Not Found**
- Accidentally deleted
- Permissions issue
- Disk full during save (file partially written)

### Step 2: Check Backup Availability

```bash
cd /var/lib/venture/saves/

# Check for automatic backup
ls -lh world.save.bak

# Check for autosave backups
ls -lh autosave_*.save | tail -10

# Verify backup integrity
sha256sum world.save.bak > world.save.bak.sha256.calc
# Compare with world.save.bak.sha256 if exists
```

If no backups exist: **Data loss likely**, proceed to Scenario 4 (manual repair).

### Step 3: Identify Last Known Good State

```bash
# List backups by date
ls -lt /var/lib/venture/saves/autosave_*.save

# Most recent is best recovery candidate
# Each autosave has timestamp: autosave_YYYYMMDD_HHMMSS.save
```

### Step 4: Check Disk Space and Permissions

```bash
# Check disk space
df -h /var/lib/venture/saves/

# Expected: >1GB free for save operations
# Alert if: <100MB free

# Check permissions
ls -la /var/lib/venture/saves/

# Expected: venture-server user has read/write access
# Fix permissions if needed:
sudo chown -R venture-server:venture-server /var/lib/venture/saves/
sudo chmod 755 /var/lib/venture/saves/
sudo chmod 644 /var/lib/venture/saves/*.save*
```

---

## Recovery Procedures

### Scenario 1: Automatic Recovery (Checksum Mismatch)

The game has built-in automatic recovery. Try restarting the server:

```bash
# Restart server (will attempt automatic recovery)
systemctl restart venture-server

# Monitor logs for recovery attempt
tail -f /var/log/venture-server.log | grep -i "recovery\|backup\|restore"

# Expected log messages:
# "Checksum mismatch detected for world.save"
# "Attempting recovery from backup: world.save.bak"
# "Backup restored successfully"
# "Server started with recovered world state"
```

If automatic recovery succeeds:
- Verify server started: `systemctl status venture-server`
- Check health: `curl http://localhost:8081/health`
- Verify world loaded: `curl http://localhost:8081/status | jq .game_state`

**Data Loss:** Up to last save interval (default: 5 minutes if autosave enabled)

### Scenario 2: Manual Backup Restore

If automatic recovery fails or you want to restore specific backup:

```bash
# Stop server first
systemctl stop venture-server

cd /var/lib/venture/saves/

# Option A: Restore most recent backup
cp world.save.bak world.save
cp world.save.bak.sha256 world.save.sha256

# Option B: Restore specific autosave
cp autosave_20260107_143022.save world.save

# Generate checksum for restored file
sha256sum world.save > world.save.sha256

# Verify integrity
venture-server -verify-save world.save

# If verification passes, start server
systemctl start venture-server

# Monitor startup
tail -f /var/log/venture-server.log
```

**Verification Commands:**
```bash
# After restore, verify world state
curl http://localhost:8081/status | jq '.game_state | {entities: .entity_count, quests: .active_quests}'

# Compare with expected values (if known)
# Log in as admin and verify player data, structures, etc.
```

### Scenario 3: List and Select from Multiple Backups

If multiple backups available, select best one:

```bash
cd /var/lib/venture/saves/

# List all backups with details
ls -lht autosave_*.save

# For each backup, check if it's valid
for backup in autosave_*.save; do
  echo "Testing $backup..."
  venture-server -verify-save $backup && echo "  ✓ Valid" || echo "  ✗ Corrupted"
done

# Select most recent valid backup
# Example: autosave_20260107_143022.save is valid

# Restore selected backup
systemctl stop venture-server
cp autosave_20260107_143022.save world.save
sha256sum world.save > world.save.sha256
systemctl start venture-server
```

### Scenario 4: Manual Repair (Advanced)

If all backups are corrupted or missing, attempt manual repair:

**WARNING:** This is advanced and risky. Only attempt if:
- You have backups of the corrupted file
- No automatic recovery succeeded
- You understand JSON/save format

```bash
cd /var/lib/venture/saves/

# Make safety copy
cp world.save world.save.corrupted

# Open in text editor (saves are JSON)
# For binary saves, use hex editor
vim world.save

# Common repairs:
# 1. Fix JSON syntax errors (missing commas, brackets)
# 2. Remove corrupted entity entries
# 3. Fix invalid field values (NaN, null where not allowed)
# 4. Restore missing sections from older backup
```

**JSON Save Structure:**
```json
{
  "version": "1.0.0",
  "seed": 12345,
  "entities": [...],
  "world_state": {...},
  "player_data": [...]
}
```

**Common Corruption Patterns:**
- Missing closing brace `}` → Add at end of file
- `"value": NaN` → Change to valid number or remove
- Truncated file → Copy missing sections from backup
- Encoding issues → Re-save as UTF-8

After manual repair:

```bash
# Validate JSON syntax (if JSON save)
jq . world.save > /dev/null && echo "Valid JSON" || echo "Invalid JSON"

# Generate checksum
sha256sum world.save > world.save.sha256

# Test load
venture-server -verify-save world.save

# If verification passes, start server
systemctl start venture-server
```

### Scenario 5: Start Fresh (Last Resort)

If all recovery attempts fail and data loss is acceptable:

```bash
# Stop server
systemctl stop venture-server

cd /var/lib/venture/saves/

# Archive corrupted files for investigation
mkdir corrupted-$(date +%Y%m%d-%H%M%S)
mv world.save* autosave_* corrupted-*/

# Start server (will generate new world)
systemctl start venture-server

# Verify new world created
tail -f /var/log/venture-server.log | grep -i "world\|save"
# Expected: "Generating new world with seed X"
```

**Post-Recovery:**
- Notify players of world reset
- Restore player inventories from backups if possible (manual process)
- Consider enabling cloud backups to prevent future loss

---

## Prevention

### Enable Automatic Backups

Ensure automatic backups are configured:

```bash
# Edit server config
vim /etc/venture/server.conf

# Enable autosave
autosave_enabled=true
autosave_interval_minutes=5  # Save every 5 minutes
autosave_keep_count=20       # Keep last 20 autosaves

# Enable backup creation before save
backup_before_save=true

# Restart server to apply
systemctl restart venture-server
```

### Configure Cloud Backup (Optional)

For critical servers, enable cloud backup:

```bash
# Example: Backup to S3 every hour
crontab -e

# Add line:
0 * * * * aws s3 sync /var/lib/venture/saves/ s3://venture-backups/$(hostname)/saves/ --exclude "*.tmp"

# Or use rsync to backup server:
0 * * * * rsync -avz /var/lib/venture/saves/ backup-server:/backups/venture/
```

### Monitoring Save Operations

Set up alerts for save failures:

```bash
# Monitor logs for save errors
# Add to monitoring system (e.g., Logstash, Splunk)
grep -i "save.*error\|save.*failed" /var/log/venture-server.log

# Set up alert if save failures detected
```

### Regular Integrity Checks

Schedule daily integrity checks:

```bash
# Add daily cron job
crontab -e

# Add line to verify save integrity daily at 3 AM
0 3 * * * /usr/local/bin/verify-venture-saves.sh

# Create verification script
cat > /usr/local/bin/verify-venture-saves.sh << 'EOF'
#!/bin/bash
cd /var/lib/venture/saves/

# Verify primary save
if sha256sum -c world.save.sha256; then
  echo "$(date): world.save OK" >> /var/log/venture-saves.log
else
  echo "$(date): world.save CORRUPTED" >> /var/log/venture-saves.log
  # Send alert (email, Slack, PagerDuty, etc.)
  echo "ALERT: Venture save file corrupted" | mail -s "Venture Save Alert" admin@example.com
fi
EOF

chmod +x /usr/local/bin/verify-venture-saves.sh
```

### Disk Health Monitoring

Monitor disk health to prevent corruption from hardware issues:

```bash
# Check SMART status (if available)
sudo smartctl -a /dev/sda | grep -i "health\|error"

# Expected: PASSED, no errors

# Set up regular SMART monitoring
sudo smartd -c /etc/smartd.conf
```

### Backup Retention Policy

Configure backup retention to balance storage and recovery:

```bash
# Keep autosaves for 24 hours
autosave_keep_count=288  # 24 hours * 60 minutes / 5 minute interval

# Archive daily backups for 30 days
# Add daily cron job
crontab -e

# Add line:
0 2 * * * cp /var/lib/venture/saves/world.save /var/lib/venture/archives/world-$(date +\%Y\%m\%d).save

# Clean old archives (keep 30 days)
0 3 * * * find /var/lib/venture/archives/ -name "world-*.save" -mtime +30 -delete
```

---

## Post-Recovery Actions

After successful recovery:

1. **Verify Data Integrity:**
   ```bash
   # Log in as admin
   venture-client -admin
   
   # Check critical data:
   # - Player inventories and progress
   # - Guild data and territories
   # - Housing and structures
   # - Active quests
   ```

2. **Notify Players:**
   - Inform players of recovery
   - Report any data loss (e.g., "progress from last 10 minutes lost")
   - Provide compensation if applicable (in-game items, time extension)

3. **Investigate Root Cause:**
   ```bash
   # Check system logs for crashes
   journalctl -u venture-server -S today | grep -i "crash\|panic\|killed"
   
   # Check disk errors
   dmesg | grep -i "error\|fail" | grep -i "sd\|disk"
   
   # Check memory issues
   grep -i "oom\|out of memory" /var/log/syslog
   ```

4. **File Bug Report (if software issue):**
   - Include corrupted save file (if comfortable sharing)
   - Include recovery logs
   - Describe events leading to corruption
   - Server version, uptime, player count

5. **Update Backup Strategy:**
   - If recovery was difficult, increase backup frequency
   - Enable cloud backups if not already
   - Test restore procedure regularly

---

## Testing Recovery Procedure

Periodically test recovery to ensure backups are functional:

```bash
# Test on non-production server
# 1. Copy production save to test server
scp production-server:/var/lib/venture/saves/world.save /tmp/test-save.save

# 2. Corrupt test save (simulate corruption)
dd if=/dev/urandom of=/tmp/test-save.save bs=1 seek=1024 count=100 conv=notrunc

# 3. Attempt recovery
venture-server -verify-save /tmp/test-save.save  # Should fail

# 4. Restore from backup
cp production-server:/var/lib/venture/saves/world.save.bak /tmp/test-save.save

# 5. Verify recovery
venture-server -verify-save /tmp/test-save.save  # Should succeed

# 6. Document time to recovery and any issues
```

Recommended test frequency: Monthly

---

## Related Runbooks

- [Memory Leak Investigation](memory-leak.md) - If OOM crash caused corruption
- [High CPU Usage](high-cpu.md) - If server overload caused incomplete save

---

## Revision History

| Date       | Author           | Changes                          |
|------------|------------------|----------------------------------|
| 2026-01-07 | Infrastructure   | Initial version for v10.0        |
