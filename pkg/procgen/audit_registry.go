package procgen

import "sync"

// AuditEntry describes a generator to be exercised by audit tests.
// Each generator package registers one entry via init() so the audit
// package can iterate over them without importing every generator directly.
type AuditEntry struct {
	Name      string
	Generator Generator
	Params    GenerationParams
}

var (
	auditMu      sync.Mutex
	auditEntries []AuditEntry
)

// RegisterAuditEntry adds an entry to the global audit registry.
// Call from an init() function in the generator's package.
func RegisterAuditEntry(e AuditEntry) {
	auditMu.Lock()
	auditEntries = append(auditEntries, e)
	auditMu.Unlock()
}

// GetAuditEntries returns a snapshot of all registered audit entries.
func GetAuditEntries() []AuditEntry {
	auditMu.Lock()
	cp := make([]AuditEntry, len(auditEntries))
	copy(cp, auditEntries)
	auditMu.Unlock()
	return cp
}
