# Venture Web/WASM Directory

This directory previously contained WebAssembly build artifacts. The WASM build system has been moved to `build/wasm/` for better organization.

## New Location

All WebAssembly build artifacts are now in: **`build/wasm/`**

## Building WebAssembly

To build the WebAssembly version:

```bash
make build-wasm
```

This will create:
- `build/wasm/venture.wasm` - The compiled game binary
- `build/wasm/wasm_exec.js` - Go WebAssembly runtime

## Testing Locally

To test the WASM build locally:

```bash
make serve-wasm
```

Then open http://localhost:8080 in your browser.

## GitHub Pages Deployment

The GitHub Actions workflow automatically deploys the WebAssembly version to GitHub Pages on every push to the `main` branch. See `.github/workflows/pages.yml` for details.

The static HTML files (`index.html` and `game.html`) are checked into version control at `build/wasm/`, while the generated files (`venture.wasm` and `wasm_exec.js`) are built during CI/CD.

## Documentation

For more information about the WebAssembly build system, see:
- `build/wasm/README.md` - WASM build documentation
- `docs/GITHUB_PAGES.md` - GitHub Pages deployment guide
