# Venture WebAssembly Build

This directory contains the static files for deploying Venture to GitHub Pages as a WebAssembly application.

## Files

- `index.html` - Landing page with game information and embedded iframe
- `game.html` - Game container that loads the WASM binary
- `venture.wasm` - The compiled game binary (generated during CI/CD)
- `wasm_exec.js` - Go WebAssembly runtime (generated during CI/CD)

## Building Locally

To build the WASM version locally:

```bash
make wasm-build
```

This will:
1. Compile the client to `build/wasm/venture.wasm`
2. Copy the Go WASM runtime to `build/wasm/wasm_exec.js`

## Testing Locally

To test the WASM build locally, you need to serve the files over HTTP:

```bash
make wasm-serve
```

Then open http://localhost:8080 in your browser.

Alternatively, use Python's built-in HTTP server:

```bash
cd build/wasm
python3 -m http.server 8080
```

## Deployment

The GitHub Actions workflow automatically:
1. Builds the WASM binary from `cmd/client`
2. Copies `wasm_exec.js` from the Go installation
3. Deploys all files from `build/wasm/` to GitHub Pages

The workflow is triggered on:
- Push to `main` branch
- Manual workflow dispatch

## Notes

- The `venture.wasm` and `wasm_exec.js` files are NOT committed to the repository
- These files are generated during the CI/CD build process
- The static HTML files (`index.html` and `game.html`) are version controlled
- The WASM binary size is approximately 10-15 MB due to the full game engine

## Browser Compatibility

Venture WebAssembly requires:
- WebAssembly support (all modern browsers since 2017)
- WebGL support for rendering
- Web Audio API for sound

Tested browsers:
- Chrome/Edge 90+
- Firefox 88+
- Safari 14+
