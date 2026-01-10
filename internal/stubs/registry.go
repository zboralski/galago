// Package stubs provides a registry for self-registering hook implementations.
// Each stub package uses init() to register its hooks, enabling clean separation of concerns.
//
// Features:
//   - Self-registering stubs via init()
//   - Detectors that activate on signature matches (e.g., Cocos2d-x, Unity IL2CPP)
//   - Lazy initialization for complex subsystems (e.g., JNI vtables)
package stubs

import (
	"fmt"
	"strings"
	"sync"

	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/hipaa"
	glog "github.com/zboralski/galago/internal/log"
	"go.uber.org/zap"
)

// HookFunc is the signature for stub hook functions.
// Returns true to stop emulation, false to continue.
type HookFunc func(emu *emulator.Emulator) bool

// StubDef defines a stub with its symbol name and hook function.
type StubDef struct {
	Name     string   // Symbol name (e.g., "malloc", "pthread_create")
	Aliases  []string // Alternative symbol names
	Hook     HookFunc
	Category string // For logging: "libc", "pthread", "jni", etc.
}

// DetectorFunc is called when a detector's pattern matches.
// It receives the emulator, imports (PLT entries), and symbols (all symbols).
// Returns the number of hooks installed.
type DetectorFunc func(emu *emulator.Emulator, imports, symbols map[string]uint64) int

// Detector defines a pattern-based activation system.
// Detectors are triggered when certain symbols/patterns are found in symbols.
type Detector struct {
	Name        string       // Detector name (e.g., "cocos2dx", "unity-il2cpp", "jni")
	Patterns    []string     // Symbol patterns to match (any match triggers)
	Activate    DetectorFunc // Called when pattern matches
	Description string       // Human-readable description
}

// Registry holds all registered stub definitions.
type Registry struct {
	mu    sync.RWMutex
	stubs map[string]*StubDef // symbol name -> stub definition

	// Detectors
	detectorsMu sync.RWMutex
	detectors   []*Detector
	activated   map[string]bool // Track which detectors have been activated

	// Callbacks
	OnCall func(category, name, detail string)

	// Emulator reference (set during Install)
	emu *emulator.Emulator
}

// DefaultRegistry is the global registry used by init() functions.
var DefaultRegistry = NewRegistry()

// NewRegistry creates a new stub registry.
func NewRegistry() *Registry {
	return &Registry{
		stubs:     make(map[string]*StubDef),
		detectors: make([]*Detector, 0),
		activated: make(map[string]bool),
	}
}

// Register adds a stub definition to the registry.
// Called from init() functions in stub packages.
func (r *Registry) Register(def StubDef) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.stubs[def.Name] = &def
	for _, alias := range def.Aliases {
		r.stubs[alias] = &def
	}

	if Debug && glog.L != nil {
		glog.L.Debug("registered",
			zap.String("cat", def.Category),
			zap.String("fn", def.Name),
			zap.Strings("aliases", def.Aliases),
		)
	}
}

// RegisterFunc is a convenience method to register a simple stub.
func (r *Registry) RegisterFunc(category, name string, hook HookFunc, aliases ...string) {
	r.Register(StubDef{
		Name:     name,
		Aliases:  aliases,
		Hook:     hook,
		Category: category,
	})
}

// RegisterDetector adds a detector that activates on pattern match.
// Detectors are checked during Install() and activated if any pattern matches.
func (r *Registry) RegisterDetector(d Detector) {
	r.detectorsMu.Lock()
	defer r.detectorsMu.Unlock()
	r.detectors = append(r.detectors, &d)

	if Debug && glog.L != nil {
		glog.L.DetectorRegister(d.Name, d.Description, d.Patterns)
	}
}

// checkDetectors runs pattern matching against symbols and activates matching detectors.
func (r *Registry) checkDetectors(emu *emulator.Emulator, imports, symbols map[string]uint64) int {
	r.detectorsMu.Lock()
	defer r.detectorsMu.Unlock()

	installed := 0

	for _, det := range r.detectors {
		// Skip already activated detectors
		if r.activated[det.Name] {
			continue
		}

		// Check if any pattern matches in symbols (includes both internal and imports)
		matched := false
		for symName := range symbols {
			for _, pattern := range det.Patterns {
				if matchPattern(symName, pattern) {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}

		if matched {
			if Debug && glog.L != nil {
				glog.L.DetectorActivate(det.Name, det.Description)
			}
			r.activated[det.Name] = true
			installed += det.Activate(emu, imports, symbols)
		}
	}

	return installed
}

// matchPattern checks if a symbol name matches a pattern.
// Patterns can use * for wildcard and can be substring matches.
func matchPattern(name, pattern string) bool {
	// Simple substring match for now
	if strings.Contains(pattern, "*") {
		// Convert glob to simple prefix/suffix matching
		if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
			// *foo* - contains
			return strings.Contains(name, pattern[1:len(pattern)-1])
		} else if strings.HasPrefix(pattern, "*") {
			// *foo - suffix
			return strings.HasSuffix(name, pattern[1:])
		} else if strings.HasSuffix(pattern, "*") {
			// foo* - prefix
			return strings.HasPrefix(name, pattern[:len(pattern)-1])
		}
	}
	// Exact match or substring
	return name == pattern || strings.Contains(name, pattern)
}

// Install hooks all registered stubs at their import addresses.
// Also runs pattern-based detectors to activate additional subsystems.
// When InstallFallbacks is true, also installs no-op stubs for unstubbed imports.
//
// Parameters:
//   - imports: PLT stub addresses for external symbols (fallbacks applied here)
//   - symbols: Optional additional symbols to search (internal functions, no fallbacks)
func (r *Registry) Install(emu *emulator.Emulator, imports map[string]uint64, symbols ...map[string]uint64) int {
	r.mu.Lock()
	// Reset activated detectors when installing to a new emulator
	// This allows running multiple entry points with fresh detector state
	if r.emu != emu {
		r.detectorsMu.Lock()
		r.activated = make(map[string]bool)
		r.detectorsMu.Unlock()
	}
	r.emu = emu
	r.mu.Unlock()

	installed := 0
	seen := make(map[uint64]bool) // Avoid double-hooking same address

	// Merge all symbol maps for detector pattern matching
	allSymbols := make(map[string]uint64)
	for k, v := range imports {
		allSymbols[k] = v
	}
	for _, syms := range symbols {
		for k, v := range syms {
			if _, exists := allSymbols[k]; !exists {
				allSymbols[k] = v
			}
		}
	}

	// First, check and activate detectors (pass both imports and merged symbols)
	installed += r.checkDetectors(emu, imports, allSymbols)

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Track which imports have stubs
	stubbed := make(map[uint64]bool)

	// Helper to install a stub at an address
	installStub := func(name string, def *StubDef, addr uint64, source string) {
		if seen[addr] {
			return
		}
		seen[addr] = true
		stubbed[addr] = true

		// Create closure to capture def
		stub := def
		emu.HookAddress(addr, func(e *emulator.Emulator) bool {
			return stub.Hook(e)
		})
		installed++

		if Debug && glog.L != nil {
			glog.L.StubInstall(def.Category, name, addr, source)
		}
	}

	// First pass: install stubs from imports (PLT entries)
	for name, def := range r.stubs {
		if addr, ok := imports[name]; ok && addr != 0 {
			installStub(name, def, addr, "import")
		}
	}

	// Second pass: install stubs from additional symbol maps (internal functions)
	for _, syms := range symbols {
		for name, def := range r.stubs {
			if addr, ok := syms[name]; ok && addr != 0 {
				installStub(name, def, addr, "internal")
			}
		}
	}

	// Install fallback stubs for unstubbed imports (return 0)
	if InstallFallbacks {
		for name, addr := range imports {
			if addr == 0 || stubbed[addr] || seen[addr] {
				continue
			}
			seen[addr] = true

			// Capture name for closure
			symName := name
			emu.HookAddress(addr, func(e *emulator.Emulator) bool {
				if Debug && glog.L != nil {
					glog.L.StubFallback(symName)
				}
				e.SetX(0, 0)
				ReturnFromStub(e)
				return false
			})
			installed++

			if Debug && glog.L != nil {
				glog.L.Debug("installed fallback",
					zap.String("fn", name),
					glog.Addr(addr),
				)
			}
		}
	}

	return installed
}

// GetEmulator returns the emulator reference.
func (r *Registry) GetEmulator() *emulator.Emulator {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.emu
}

// Log calls the OnCall callback and logs via zap.
// This is the primary method for stubs to report their activity.
func (r *Registry) Log(category, name, detail string) {
	r.mu.RLock()
	cb := r.OnCall
	emu := r.emu
	r.mu.RUnlock()

	// HIPAA compliance: Sanitize detail to prevent PHI leaks in logs
	// In medical practice, we redact sensitive information from records.
	sanitizedDetail := detail
	if hipaa.SessionDetector != nil && hipaa.SessionDetector.ContainsPHI(detail) {
		sanitizedDetail = hipaa.SessionDetector.SanitizePHI(detail)
		if hipaa.SessionAuditor != nil {
			hipaa.SessionAuditor.LogPHIDetected("Stub log", detail)
		}
	}

	// Get PC from emulator if available
	var pc uint64
	if emu != nil {
		pc = emu.LR() // Return address of stub call
	}

	// Call trace callback (for trace event collection)
	if cb != nil {
		cb(category, name, sanitizedDetail)
	}

	// Log via zap at debug level
	if glog.L != nil {
		glog.L.Trace(pc, category, name, sanitizedDetail)
	}
}

// Count returns the number of registered stubs.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.stubs)
}

// List returns all registered stub names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.stubs))
	seen := make(map[string]bool)
	for name, def := range r.stubs {
		if seen[def.Name] {
			continue
		}
		seen[def.Name] = true
		names = append(names, name)
	}
	return names
}

// Debug enables verbose logging during installation.
var Debug = false

// InstallFallbacks enables fallback stubs for unstubbed imports.
// When true, all unknown imports get a stub that returns 0.
var InstallFallbacks = true

// Convenience functions for the default registry

// Register adds a stub to the default registry.
func Register(def StubDef) {
	DefaultRegistry.Register(def)
}

// RegisterFunc adds a simple stub to the default registry.
func RegisterFunc(category, name string, hook HookFunc, aliases ...string) {
	DefaultRegistry.RegisterFunc(category, name, hook, aliases...)
}

// Install hooks all stubs in the default registry.
func Install(emu *emulator.Emulator, imports map[string]uint64, symbols ...map[string]uint64) int {
	return DefaultRegistry.Install(emu, imports, symbols...)
}

// RegisterDetector adds a detector to the default registry.
func RegisterDetector(d Detector) {
	DefaultRegistry.RegisterDetector(d)
}

// Helper functions for stubs

// ReturnFromStub sets PC to LR to return from the current function.
func ReturnFromStub(emu *emulator.Emulator) {
	emu.SetPC(emu.LR())
}

// FormatHex formats a value as hex string.
func FormatHex(v uint64) string {
	if v == 0 {
		return "0"
	}
	return fmt.Sprintf("0x%x", v)
}

// FormatPtr formats name=value pairs.
func FormatPtr(name string, val uint64) string {
	return name + "=" + FormatHex(val)
}

// FormatPtrPair formats two name=value pairs.
func FormatPtrPair(name1 string, val1 uint64, name2 string, val2 uint64) string {
	if name2 == "" {
		return FormatPtr(name1, val1)
	}
	return FormatPtr(name1, val1) + " " + FormatPtr(name2, val2)
}
