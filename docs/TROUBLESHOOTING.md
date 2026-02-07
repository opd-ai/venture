# Troubleshooting Guide

**Venture - Fully Procedural Multiplayer Action-RPG**  
**Version:** 1.0.0  
**Last Updated:** February 2026

This guide provides solutions to common technical issues, organized by symptom.

---

## Table of Contents

1. [Quick Diagnosis](#quick-diagnosis)
2. [Checking Logs](#checking-logs)
3. [Common Issues](#common-issues)
4. [Platform-Specific Issues](#platform-specific-issues)
5. [Advanced Troubleshooting](#advanced-troubleshooting)
6. [Reporting Bugs](#reporting-bugs)
7. [Getting Help](#getting-help)

---

## Quick Diagnosis

**Start here if unsure:** [Logs](#checking-logs) → [Common Issues](#common-issues) → [Platform-Specific](#platform-specific-issues) → [Report Bug](#reporting-bugs)

---

## Checking Logs

**Location:**
- **Client:** `logs/venture-client.log`
- **Server:** `logs/venture-server.log`
- **Web:** Browser Console (F12 → Console tab)

**Log Levels:**
- **DEBUG:** Verbose internal state (not errors)
- **INFO:** Normal operation events
- **WARN:** Recoverable issues (performance degradation, missing assets)
- **ERROR:** Operation failures (requires investigation)
- **FATAL:** Unrecoverable errors (crash)

**Tip:** Search for `ERROR` or `FATAL` keywords in logs for root cause.

---

## Common Issues

### Game Won't Launch

**Symptom:** Double-clicking executable does nothing, or crashes immediately.

**Solutions:**

1. **Missing Dependencies (Linux):**
   ```bash
   sudo apt install libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev \
   libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config \
   libx11-dev
   ```

2. **Permissions (Linux/macOS):**
   ```bash
   chmod +x venture-client
   ./venture-client
   ```

3. **Visual C++ Redistributables (Windows):**
   - Download: [vc_redist.x64.exe](https://aka.ms/vs/17/release/vc_redist.x64.exe)
   - Install and restart

4. **macOS Gatekeeper:**
   - System Preferences → Security & Privacy → "Open Anyway" button
   - Or: `xattr -d com.apple.quarantine venture-client`

5. **Antivirus False Positive:**
   - Whitelist `venture-client.exe` in antivirus settings
   - Known issue with Windows Defender (safe to allow)

---

### Low FPS / Stuttering

**Symptom:** Frame rate below 30 FPS, periodic freezing.

**Solutions:**

1. **Reduce Resolution:**
   - Settings → Video → Resolution: 1280×720 (from 1920×1080)

2. **Lower Graphics Quality:**
   - Settings → Graphics → Quality Preset: Low
   - Disable: Bloom, Ambient Occlusion, Motion Blur

3. **Reduce Particle Count:**
   - Settings → Graphics → Particle Multiplier: 0.5 (from 1.0)

4. **Clear Cache:**
   - Settings → Graphics → Clear Sprite Cache
   - Restart game

5. **Update GPU Drivers:**
   - NVIDIA: [GeForce Experience](https://www.nvidia.com/en-us/geforce/geforce-experience/)
   - AMD: [Radeon Software](https://www.amd.com/en/support)
   - Intel: [Download Center](https://downloadcenter.intel.com/)

6. **Dedicated GPU (Laptops):**
   - NVIDIA Control Panel → Manage 3D Settings → Program Settings → Add `venture-client.exe` → Preferred GPU: High-performance NVIDIA processor
   - Or AMD Radeon Settings equivalent

**Performance Mode:** `-quality=low` flag bypasses menu, applies all optimizations at launch.

---

### Multiplayer Connection Failed

**Symptom:** "Failed to connect to server" or timeout errors.

**Solutions:**

1. **Server Not Running:**
   - Verify server process: `ps aux | grep venture-server` (Linux/macOS) or Task Manager (Windows)
   - Start server: `./venture-server -port 8080`

2. **Incorrect IP/Port:**
   - Find server IP: `ifconfig` (Linux/macOS) or `ipconfig` (Windows)
   - Default port: 8080 (verify with `-port` flag)
   - Connect: `./venture-client -multiplayer -server 192.168.1.100:8080`

3. **Firewall Blocking:**
   ```bash
   # Linux (ufw)
   sudo ufw allow 8080/tcp
   sudo ufw allow 8080/udp
   
   # Windows Firewall
   # Control Panel → Windows Defender Firewall → Advanced Settings → Inbound Rules → New Rule → Port → TCP/UDP 8080
   ```

4. **Port Forwarding (Internet Play):**
   - Router admin panel (192.168.1.1 or 192.168.0.1)
   - Port Forward: External 8080 → Internal <server-local-IP>:8080
   - Protocol: Both TCP and UDP

5. **Timeout Issues:**
   - Increase timeout: `-timeout 60` (60 seconds, default: 10)
   - High-latency mode: `-high-latency` flag

---

### Audio Problems

**Crackling / Distortion:**
- Settings → Audio → Buffer Size: 2048 (from 512)
- Sample Rate: 44100 Hz (avoid 48000 Hz)

**No Sound:**
- Check OS volume mixer (ensure Venture not muted)
- Settings → Audio → Output Device: Select correct device
- Restart audio subsystem: Settings → Audio → Restart Audio

**Echo / Reverb:**
- Disable: Settings → Audio → Environmental Effects: Off

**WebAssembly Latency:**
- Expected: 50-100ms delay (browser limitation)
- No fix available, use desktop client for best audio

---

### Save File Corruption

**Symptom:** "Failed to load save" or corrupted data errors.

**Recovery Steps:**

1. **Restore Backup:**
   ```bash
   cd saves/backups/
   ls -lt  # Find most recent autosave
   cp autosave_20251225_143022.save ../current.save
   ```

2. **Manual Repair (Advanced):**
   - Save files are JSON
   - Open in text editor, look for syntax errors (missing commas, brackets)
   - Fix manually or restore missing sections from older backup

3. **Cloud Backup (if enabled):**
   - Settings → Saves → Restore from Cloud

**Prevention:**
- Enable auto-backup: Settings → Saves → Auto-Backup: Every 10 minutes
- Cloud sync: Settings → Saves → Cloud Backup: Enable

---

### Screen Flickering / Visual Artifacts

**Solutions:**

1. **Update GPU Drivers** (critical!)

2. **Disable Post-Processing:**
   - Settings → Graphics → Bloom: Off
   - AO: Off, Motion Blur: Off, Vignette: Off

3. **Change Renderer:**
   - `-renderer opengl2` or `-renderer opengl3` flag (try both)

4. **Fullscreen Toggle:**
   - Press F11 twice (switch windowed → fullscreen → windowed)

5. **Monitor Refresh Rate:**
   - Ensure native refresh rate (60Hz, 120Hz, 144Hz)
   - Avoid non-standard rates (75Hz)

6. **V-Sync:**
   - Settings → Graphics → V-Sync: On (prevents tearing)

---

### Crashes / Freezes

**Symptom:** Game exits unexpectedly, "Not Responding" dialog.

**Immediate Actions:**
1. Check log file for `FATAL` entries
2. Note exact activity when crash occurred (combat, menu, multiplayer join)

**Common Causes:**

**Out of Memory:**
- Reduce cache size: Settings → Graphics → Cache Size: 50MB (from 150MB)
- Close background apps (browsers, editors)

**Invalid Save Data:**
- Load previous backup (see [Save File Corruption](#save-file-corruption))

**GPU Driver Crash:**
- Update drivers (see [Low FPS](#low-fps--stuttering) solutions)

**Mod Conflict:**
- Disable mods: Remove all `.venturemod` files from `mods/` folder
- Test without mods, re-enable one-by-one to isolate culprit

**Submit Crash Report:**
- Include `logs/venture-client.log` + steps to reproduce
- [GitHub Issues](https://github.com/opd-ai/venture/issues/new?labels=crash)

---

## Platform-Specific Issues

### Linux

**Black Screen on Launch:**
- Missing OpenGL: `sudo apt install mesa-utils`
- Verify GL support: `glxinfo | grep "OpenGL version"`
- Required: OpenGL 2.1+

**X11 Errors:**
- Install X11 dev libraries (see [Missing Dependencies](#game-wont-launch))

**Wayland Compatibility:**
- Set `GDK_BACKEND=x11` environment variable
- Or run under XWayland

### macOS

**"damaged and can't be opened" Error:**
```bash
xattr -cr /path/to/Venture.app
```

**Apple Silicon Performance:**
- Prefer ARM64 build (`venture-v10.0-macos-arm64.tar.gz`)
- Rosetta 2 works but slower (x64 build)

**Sound Issues:**
- Grant microphone permission (Privacy & Security → Microphone)

### Windows

**msvcr120.dll Missing:**
- Install [Visual C++ 2013 Redistributable](https://www.microsoft.com/en-us/download/details.aspx?id=40784)

**SmartScreen Warning:**
- Click "More info" → "Run anyway"
- Binary is unsigned (open-source project)

**High DPI Scaling:**
- Right-click `venture-client.exe` → Properties → Compatibility → Override high DPI scaling: Application

### WebAssembly

**Blank Screen:**
- Check browser console (F12 → Console) for errors
- Required: WebGL 2.0 (test at [WebGL Report](https://webglreport.com/))

**Slow Loading:**
- First load downloads ~15MB wasm file (slow on mobile data)
- Subsequent loads use cache (instant)

**Keyboard Conflicts:**
- Alt key triggers browser menu: Use Ctrl instead
- F1-F12 may conflict: Rebind controls (Settings → Controls)

### Mobile

**Touch Unresponsive:**
- Reduce Background Apps: Free 500MB+ RAM
- Screen Protector: Replace with high-sensitivity protector
- Update OS: Latest iOS/Android recommended

**Battery Drain:**
- Reduce frame rate: Settings → Graphics → FPS Limit: 30
- Lower resolution: Settings → Video → Resolution: 720p

**Overheating:**
- Enable Low Power Mode (iOS Settings → Battery)
- Reduce graphics quality (Settings → Graphics → Low)

---

## Advanced Troubleshooting

### Enable Debug Logging

Add `-verbose` flag to client/server:
```bash
./venture-client -verbose > debug.log 2>&1
```

Logs all internal operations to `debug.log` (large file, use only when needed).

### Profile Performance

```bash
./venture-client -profile-cpu
```

Generates `cpu.prof` file. Analyze with:
```bash
go tool pprof cpu.prof
# (pprof) top20
# (pprof) list FunctionName
```

### Test Mode

```bash
./venture-client -test-mode
```

Skips main menu, loads test world (seed 0), enables debug overlay (F3).

### Network Diagnostics

**Latency Test:**
```bash
ping <server-ip>  # Check packet loss and RTT
```

**Port Open Test:**
```bash
telnet <server-ip> 8080  # Should connect
# Or: nc -zv <server-ip> 8080
```

**Bandwidth Test:**
- Use multiplayer for 5 minutes
- Check: Settings → Network → Statistics
- Expected: <100 KB/s send/receive

---

## Reporting Bugs

**Before Reporting:**
1. Reproduce bug (can you trigger it consistently?)
2. Check existing issues: [GitHub Issues](https://github.com/opd-ai/venture/issues)
3. Collect data:
   - Log file (`logs/venture-client.log`)
   - OS/GPU info (`Settings → About`)
   - Steps to reproduce

**Create Issue:**
1. Go to [New Issue](https://github.com/opd-ai/venture/issues/new/choose)
2. Select template: Bug Report
3. Fill all sections (repro steps, expected vs. actual behavior)
4. Attach log file + screenshot (if visual bug)
5. Add labels: `bug`, `platform:windows`, etc.

**Priority Labels:**
- `critical`: Game unplayable (crashes, won't launch)
- `high`: Major feature broken (save file corruption, multiplayer desync)
- `medium`: Minor feature broken (UI glitch, audio distortion)
- `low`: Cosmetic issue (typo, visual artifact)

---

## Getting Help

**Community Support:**
- **Discord:** #help channel (fastest response, 100+ online users)
- **GitHub Discussions:** [Q&A Section](https://github.com/opd-ai/venture/discussions)
- **Reddit:** r/VentureRPG (community-run)

**Developer Support:**
- **GitHub Issues:** Technical bugs only
- **Email:** support@venture-rpg.com (response time: 48-72 hours)

**Documentation:**
- [FAQ](FAQ.md): 50+ common questions
- [User Manual](USER_MANUAL.md): Comprehensive feature guide
- [Getting Started](GETTING_STARTED.md): Setup and tutorials

---

**Can't find a solution?** Join our [Discord](https://discord.gg/venture) - the community is happy to help!

**Version:** 1.0.0 (February 2026)  
**Maintained By:** Venture Development Team & Community Contributors
