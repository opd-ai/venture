# WebAssembly Quick Reference

## Building for WebAssembly

### Quick Build
```bash
make build-wasm
```

### Manual Build
```bash
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o web/venture.wasm ./cmd/client
cp $(go env GOROOT)/lib/wasm/wasm_exec.js web/
```

### Build and Test Locally
```bash
make serve-wasm
# Opens http://localhost:8080
```

## File Structure

```
web/
├── README.md           # This file (documentation)
├── .gitignore         # Ignores generated files
├── venture.wasm       # Generated: Compiled game (10-20 MB)
├── wasm_exec.js       # Generated: Go WASM runtime
├── game.html          # Generated: Game container
└── index.html         # Generated: Landing page
```

## GitHub Actions Deployment

### Automatic Deployment
- **Trigger:** Every push to `main` branch
- **Workflow:** `.github/workflows/pages.yml`
- **URL:** https://opd-ai.github.io/venture/

### Manual Deployment
1. Go to Actions tab on GitHub
2. Select "Deploy to GitHub Pages"
3. Click "Run workflow"

## Development Tips

### WASM Build Flags
- `-ldflags="-s -w"` - Strips debug info, reduces size by 20-30%
- `GOOS=js GOARCH=wasm` - Targets WebAssembly

### Testing Locally
Always test WASM builds locally before pushing:
```bash
# Option 1: Use make
make serve-wasm

# Option 2: Python
cd web && python3 -m http.server 8080

# Option 3: Any HTTP server
# Must serve from web/ directory
```

### Browser Console
Open developer tools (F12) to see:
- WASM loading status
- Runtime errors
- Performance metrics

### Common Issues

**WASM won't load:**
- Check file exists: `ls -lh web/venture.wasm`
- Verify file size is reasonable (5-30 MB)
- Check browser console for errors

**Blank screen:**
- WASM may still be loading (large file)
- Check console for JavaScript errors
- Verify wasm_exec.js is present

**Performance issues:**
- WebAssembly is slower than native
- Try Chrome/Edge for best performance
- Check if hardware acceleration is enabled

## Platform-Specific Code

### Conditional Compilation
Use build tags to exclude problematic code from WASM:

```go
//go:build !js && !wasm
// +build !js,!wasm

package mypackage

// This code won't be compiled for WASM
```

### WASM-Specific Code
```go
//go:build js && wasm
// +build js,wasm

package mypackage

// This code only compiles for WASM
```

## Resources

- [Ebiten WebAssembly Guide](https://ebitengine.org/en/documents/webassembly.html)
- [Go WebAssembly Wiki](https://github.com/golang/go/wiki/WebAssembly)
- [GitHub Pages Guide](GITHUB_PAGES.md) - Full deployment documentation
