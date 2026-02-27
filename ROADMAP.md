~~~~
# PRODUCTION READINESS ASSESSMENT: Venture

## CRITICAL ISSUES

### Application Security Concerns:

- **Single production panic** in `pkg/rendering/display/config.go:60` — `panic()` during
  default config construction could crash a server or client if the hardcoded resolution
  constant is ever mutated. Replace with an error return or `log.Fatal` at startup.
- **Defer-based shutdown only** — `cmd/server/main.go` relies on `defer` for cleanup with
  no explicit `signal.Notify(SIGINT, SIGTERM)` handler. A non-graceful kill signal could
  bypass cleanup, leaving resources (sockets, temp files) open.
- **Note:** Transport security (TLS/HTTPS) is outside scope — assumed handled by
  infrastructure (reverse proxy, load balancer, or container orchestration).

### Reliability Concerns:

- **No explicit OS-signal handler** in the server entry point. If the process receives
  SIGTERM (e.g., from Kubernetes or systemd) during a long-running operation, deferred
  cleanup may not execute before the hard kill timeout.
- **`context.Background()` in 31 call sites** — many are acceptable (main, init, tests),
  but a handful inside long-lived goroutines would benefit from a cancellable parent
  context wired to the server lifecycle.
- **VR package (`pkg/vr`) has zero test coverage.** While feature-gated, any future
  activation could introduce untested code paths.

### Performance Concerns:

- **Sprite cache default 16 MB** (`pkg/rendering/cache/sprite_cache.go`) — for large
  worlds or high-resolution displays, this may trigger frequent LRU evictions and
  re-generation overhead. The maximum (300 MB) is tunable but not documented as a
  runtime flag.
- **Rate-limiter cleanup runs on a fixed 10-minute interval**
  (`pkg/validation/ratelimit.go`). Under sudden connection surges, stale tracking
  entries may accumulate until the next sweep.

## IMPLEMENTATION ROADMAP

### Phase 1: Foundation (High Priority)
**Duration:** 2–4 weeks
**Focus:** Essential production safety and operational readiness

**Tasks:**

1. **Add OS-signal handling to the server**
   - Wire `signal.Notify` for SIGINT and SIGTERM in `cmd/server/main.go`.
   - On signal receipt, cancel a root `context.Context`, triggering orderly shutdown of
     all subsystems (metrics exporter, stability monitor, network server).
   - *Acceptance criteria:* `kill -TERM <pid>` results in logged graceful shutdown with
     zero resource leaks; integration test validates exit code 0.

2. **Replace panic with error handling in display config**
   - In `pkg/rendering/display/config.go`, change the `panic()` in
     `NewConfigDefault()` to return `(*Config, error)`.
   - Propagate the error to callers; log and exit at the application boundary if the
     default resolution is invalid.
   - *Acceptance criteria:* `go vet ./pkg/rendering/display/...` passes; no `panic()`
     calls remain in non-test production code.

3. **Establish structured logging standards**
   - The logging package (`pkg/logging`) and Logrus integration are already in place.
   - Audit call sites still using `fmt.Printf` or bare `log.*` and migrate them to the
     structured logger with standard fields (`system_name`, `entityID`, `seed`).
   - *Acceptance criteria:* `grep -rn 'fmt.Print\|log.Print' pkg/ cmd/` returns zero
     matches outside test helpers and example files.

4. **Harden input validation coverage**
   - Chat validation (`pkg/validation/chat.go`): add max-byte-length check (not just
     rune length) to prevent oversized UTF-8 payloads.
   - Trade validation (`pkg/validation/trade.go`): ensure negative-quantity and
     zero-value trades are rejected.
   - *Acceptance criteria:* Table-driven tests cover boundary values; `go test -race
     ./pkg/validation/...` passes.

5. **Document all runtime configuration flags**
   - Consolidate the 50+ flags from `cmd/server/main.go` and `cmd/client/util.go` into
     a single `docs/CONFIGURATION.md` with defaults, valid ranges, and environment
     variable equivalents.
   - *Acceptance criteria:* Every `flag.*` call in `cmd/` has a corresponding entry in
     the configuration document.

### Phase 2: Performance & Reliability (Medium Priority)
**Duration:** 3–5 weeks
**Focus:** Production resilience under load

**Tasks:**

1. **Wire cancellable contexts through long-lived goroutines**
   - Replace bare `context.Background()` in goroutines spawned by systems (world
     persistence auto-save, stability monitor loop, metrics exporter) with a context
     derived from the server's root context.
   - *Acceptance criteria:* Server shutdown completes within 5 seconds under normal load;
     `go vet -copylocks ./...` and `-race` pass.

2. **Tune sprite cache and expose configuration**
   - Add a `--sprite-cache-mb` flag (default 64, max 300) to the client.
   - Add cache-pressure metrics to the observability exporter: hit rate, eviction rate,
     current size.
   - *Acceptance criteria:* Under 2 000-entity benchmark, cache hit rate ≥ 90 %; metrics
     visible at `/metrics` endpoint.

3. **Enhance rate-limiter adaptive cleanup**
   - When tracked-client count exceeds a configurable high-water mark (e.g., 10 000),
     trigger an immediate sweep rather than waiting for the next 10-minute interval.
   - *Acceptance criteria:* Unit test simulates 20 000 clients; memory does not grow
     unbounded; cleanup fires before the scheduled interval.

4. **Add integration tests for federation resilience**
   - Test circuit-breaker state transitions (Closed → Open → HalfOpen → Closed) with
     the scenario runner in `pkg/network/resilience/`.
   - Test connection-pool stale-entry cleanup under simulated latency.
   - *Acceptance criteria:* `go test -tags integration ./pkg/network/federation/...`
     passes; scenarios cover all three circuit-breaker states.

5. **Expand test coverage for VR package**
   - Add unit tests for `pkg/vr` detection logic and rendering stubs.
   - *Acceptance criteria:* Coverage ≥ 30 % for `pkg/vr`.

### Phase 3: Operational Excellence (Lower Priority)
**Duration:** 4–6 weeks
**Focus:** Long-term maintainability and observability

**Tasks:**

1. **Centralize health-check endpoint**
   - Expose an HTTP `/healthz` (liveness) and `/readyz` (readiness) on the metrics port
     (default 9090).
   - Readiness checks: world loaded, network listener bound, at least one successful
     game tick completed.
   - *Acceptance criteria:* `curl localhost:9090/healthz` returns 200; Kubernetes
     liveness/readiness probes documented in `docs/PRODUCTION_DEPLOYMENT.md`.

2. **Add request-correlation propagation**
   - The correlation-ID infrastructure exists in `pkg/errors/correlation.go`.
   - Propagate correlation IDs through network packets and log entries so a single
     client action can be traced end-to-end (client → server → federation peer).
   - *Acceptance criteria:* A sample trace in structured logs shows the same correlation
     ID across client-send, server-receive, and federation-forward.

3. **Automate dependency vulnerability scanning**
   - Add `govulncheck` to the CI pipeline (`build.yml` or `test.yml`).
   - Gate merges on zero known-vulnerable dependencies.
   - *Acceptance criteria:* `govulncheck ./...` runs in CI; pipeline fails on
     HIGH/CRITICAL findings.

4. **Performance regression benchmarks in CI**
   - Add a CI step that runs `go test -bench=. -benchmem ./pkg/engine/...` and compares
     against stored baselines using `benchstat`.
   - *Acceptance criteria:* CI warns on ≥ 10 % regression in allocation count or
     ns/op for tagged benchmarks.

5. **Formalize rollback procedures**
   - Document rollback steps for server binary, save-file schema (leveraging
     `pkg/saveload/migrator.go`), and federation protocol version mismatches.
   - *Acceptance criteria:* Runbook in `docs/` with step-by-step rollback validated
     against the version-compatibility checks in `pkg/version/`.

## RECOMMENDED LIBRARIES

| Need | Current | Recommendation | Justification |
|------|---------|----------------|---------------|
| Structured logging | Logrus v1.9.3 | **Keep Logrus** (or migrate to `log/slog` in Go 1.24) | Logrus is stable and widely adopted; `slog` is stdlib and zero-dependency. Migration is optional and low-risk. |
| Configuration | `flag` stdlib | **Keep `flag`** | Complexity does not yet warrant Viper; CLI flags are well-organized and validated via `pkg/config`. |
| Metrics export | Custom HTTP (`pkg/observability`) | **Keep custom** (consider OpenTelemetry long-term) | Current Prometheus-compatible exporter is lightweight; OpenTelemetry adds dependency weight only justified at scale. |
| Vulnerability scan | None | **`govulncheck`** (golang.org/x/vuln) | Official Go team tool; integrates with `go.sum`; zero configuration. |
| Benchmark comparison | None | **`benchstat`** (golang.org/x/perf) | Standard tool for statistically sound benchmark comparison in CI. |

**Transport Security Exclusion:** No TLS/SSL, HTTPS, or certificate-management libraries
are recommended. Transport encryption is assumed to be handled by reverse proxies, load
balancers, or container orchestration platforms as documented in
`docs/PRODUCTION_DEPLOYMENT.md`.

## SUCCESS CRITERIA

| Indicator | Target | Measurement |
|-----------|--------|-------------|
| Zero production panics | 0 `panic()` in non-test code | `grep -rn 'panic(' pkg/ cmd/ \| grep -v _test.go` |
| Graceful shutdown | Server exits 0 on SIGTERM within 5 s | Integration test + manual verification |
| Test coverage | ≥ 40 % per package (≥ 30 % for display-dependent) | `go test -cover ./pkg/...` |
| Benchmark stability | < 10 % regression gate in CI | `benchstat` comparison |
| Dependency safety | 0 HIGH/CRITICAL vulns | `govulncheck ./...` in CI |
| Structured logging | 100 % of production log calls use Logrus | Grep audit for `fmt.Print`/`log.Print` |
| Health endpoints | `/healthz` and `/readyz` return 200 | Automated probe in CI or smoke test |
| Configuration docs | 100 % flag coverage | Manual review against `flag.*` declarations |

## RISK ASSESSMENT

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Signal-handler change breaks hot-reload or debug workflows | Low | Medium | Feature-flag the new handler; test on all target platforms. |
| Panic removal in display config cascades type changes to many callers | Medium | Low | Use a staged migration: add error-returning variant first, deprecate panic variant. |
| `slog` migration (if pursued) causes subtle log-level behavior changes | Low | Low | Run parallel logging during migration; compare output in staging. |
| CI benchmark baselines become stale after hardware changes | Medium | Low | Store baselines per runner class; regenerate on infrastructure changes. |
| Federation circuit-breaker tuning too aggressive | Medium | Medium | Expose thresholds as server flags; monitor state-transition metrics in production. |

## SECURITY SCOPE CLARIFICATION

- Analysis focuses on **application-layer security only**.
- Transport encryption (TLS/HTTPS) is assumed to be handled by deployment infrastructure
  (reverse proxy, load balancer, service mesh, or Tor hidden-service layer).
- No recommendations for certificate management or SSL/TLS configuration are included.
- The codebase already passes all 30 automated security-audit checks
  (`pkg/security/audit.go`) covering federation, chat encryption, mod sandboxing, input
  validation, anti-cheat, and privacy domains.
- AES-256-GCM encryption with DH key exchange (`pkg/network/crypto.go`) is implemented
  correctly for application-level packet encryption.
- The mod system (`pkg/modding/sandbox.go`) is secure by design — JSON-only data mods
  with no executable code paths.
~~~~
