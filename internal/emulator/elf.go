package emulator

import (
	"debug/elf"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// ARM64 relocation types
const (
	R_AARCH64_ABS64     = 257  // Absolute 64-bit symbol reference
	R_AARCH64_GLOB_DAT  = 1025 // GOT entry for global data symbol
	R_AARCH64_JUMP_SLOT = 1026 // PLT GOT entry for function call
	R_AARCH64_RELATIVE  = 1027 // Position-independent data reference
)

// ELFInfo contains parsed ELF metadata
type ELFInfo struct {
	Path     string
	Machine  elf.Machine
	Entry    uint64
	Symbols  map[string]uint64 // symbol name -> virtual address (all symbols)
	Imports  map[string]uint64 // symbol name -> PLT stub address (external imports only)
	Segments []Segment
	BaseAddr uint64     // Load base address
	EndAddr  uint64     // End of loaded memory
	VTables  *VTableMap // Resolved C++ vtables (slot -> function mapping)
}

// Segment represents a loadable ELF segment
type Segment struct {
	VAddr  uint64
	PAddr  uint64
	Offset uint64
	Size   uint64 // File size
	MemSz  uint64 // Memory size (may be larger due to .bss)
	Flags  elf.ProgFlag
	Data   []byte
}

// LoadELFBase is the default base address for position-independent libraries.
// Android shared libraries typically load around 0x7xxxxxxxxx but we use a
// lower address for simpler emulation.
const LoadELFBase = 0x40000000 // 1GB

// LoadELF loads an ELF file and maps it into the emulator.
// Position-independent shared libraries (base addr 0) are relocated to LoadELFBase.
func (e *Emulator) LoadELF(path string) (*ELFInfo, error) {
	return e.LoadELFAt(path, 0) // 0 means auto-select base
}

// LoadELFAt loads an ELF file at a specific base address.
// If loadBase is 0, auto-selects based on file type:
// - Executables: use vaddr from file
// - Shared libraries (vaddr=0): relocate to LoadELFBase
func (e *Emulator) LoadELFAt(path string, loadBase uint64) (*ELFInfo, error) {
	f, err := elf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open ELF: %w", err)
	}
	defer f.Close()

	// Verify ARM64
	if f.Machine != elf.EM_AARCH64 {
		return nil, fmt.Errorf("expected ARM64 (EM_AARCH64), got %v", f.Machine)
	}

	// Find file base address (lowest PT_LOAD vaddr)
	fileBase := uint64(0xFFFFFFFFFFFFFFFF)
	fileEnd := uint64(0)

	for _, prog := range f.Progs {
		if prog.Type != elf.PT_LOAD {
			continue
		}
		if prog.Vaddr < fileBase {
			fileBase = prog.Vaddr
		}
		segEnd := prog.Vaddr + prog.Memsz
		if segEnd > fileEnd {
			fileEnd = segEnd
		}
	}

	if fileBase == 0xFFFFFFFFFFFFFFFF {
		return nil, fmt.Errorf("no PT_LOAD segments found")
	}

	// Determine relocation base
	// PIE/shared libraries have fileBase=0 or very low, need to relocate
	var relocOffset uint64
	if loadBase != 0 {
		// Explicit base requested
		relocOffset = loadBase - fileBase
	} else if fileBase < 0x10000 {
		// Position-independent, relocate to default base
		relocOffset = LoadELFBase - fileBase
	} else {
		// Use file's vaddr as-is
		relocOffset = 0
	}

	info := &ELFInfo{
		Path:     path,
		Machine:  f.Machine,
		Entry:    f.Entry + relocOffset,
		Symbols:  make(map[string]uint64),
		Imports:  make(map[string]uint64),
		BaseAddr: fileBase + relocOffset,
		EndAddr:  fileEnd + relocOffset,
	}

	// Load symbols from .dynsym and .symtab (with relocation)
	// Strip version suffixes (@@VERSION or @VERSION) for consistent lookup
	syms, err := f.DynamicSymbols()
	if err == nil {
		for _, sym := range syms {
			if sym.Value != 0 && sym.Name != "" {
				addr := sym.Value + relocOffset
				info.Symbols[sym.Name] = addr
				// Also store without version suffix for easier lookup
				if idx := strings.Index(sym.Name, "@@"); idx != -1 {
					info.Symbols[sym.Name[:idx]] = addr
				} else if idx := strings.Index(sym.Name, "@"); idx != -1 {
					info.Symbols[sym.Name[:idx]] = addr
				}
			}
		}
	}

	syms, err = f.Symbols()
	if err == nil {
		for _, sym := range syms {
			if sym.Value != 0 && sym.Name != "" {
				info.Symbols[sym.Name] = sym.Value + relocOffset
			}
		}
	}

	// Read file data for segments
	fileData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	// Load PT_LOAD segments
	for _, prog := range f.Progs {
		if prog.Type != elf.PT_LOAD {
			continue
		}

		// Apply relocation
		loadVAddr := prog.Vaddr + relocOffset

		seg := Segment{
			VAddr:  loadVAddr,
			PAddr:  prog.Paddr + relocOffset,
			Offset: prog.Off,
			Size:   prog.Filesz,
			MemSz:  prog.Memsz,
			Flags:  prog.Flags,
		}

		// Extract segment data
		if prog.Filesz > 0 && prog.Off+prog.Filesz <= uint64(len(fileData)) {
			seg.Data = fileData[prog.Off : prog.Off+prog.Filesz]
		}

		info.Segments = append(info.Segments, seg)

		// Map segment into emulator memory (aligned to page boundary)
		pageSize := uint64(0x1000)
		alignedAddr := loadVAddr & ^(pageSize - 1)
		alignedEnd := (loadVAddr + prog.Memsz + pageSize - 1) & ^(pageSize - 1)
		alignedSize := alignedEnd - alignedAddr

		// Map memory (ignore error if already mapped)
		_ = e.MapRegion(alignedAddr, alignedSize)

		// Write segment data
		if len(seg.Data) > 0 {
			if err := e.MemWrite(loadVAddr, seg.Data); err != nil {
				return nil, fmt.Errorf("write segment at 0x%x: %w", loadVAddr, err)
			}
		}

		// Zero out .bss portion (memory size > file size)
		if prog.Memsz > prog.Filesz {
			bssStart := loadVAddr + prog.Filesz
			bssSize := prog.Memsz - prog.Filesz
			zeros := make([]byte, bssSize)
			// Non-fatal if this fails
			_ = e.MemWrite(bssStart, zeros)
		}
	}

	// Build PLT stub address map FIRST (needed for relocation second pass)
	// PLT addresses go to Imports map (for stub installation) AND Symbols map (for lookups)
	addPLTSymbols(f, relocOffset, info.Symbols, info.Imports)

	// Apply relocations to fix GOT entries
	// First pass handles internal symbols, second pass resolves external symbols to PLT stubs
	if err := e.applyRelocations(f, relocOffset, info.Imports); err != nil {
		return nil, fmt.Errorf("apply relocations: %w", err)
	}

	// Build vtable map for C++ virtual function resolution
	// This parses ELF relocations to resolve vtable slot -> function address
	vtables, err := BuildVTableMap(f, relocOffset)
	if err == nil {
		info.VTables = vtables
	}

	// Initialize std::string globals that are NULL (point to empty string)
	// Many Cocos2d-x binaries have single-letter global std::strings (z, v, a, b, etc.)
	// that need to be initialized to point to the empty string representation
	e.initStringGlobals(info.Symbols)

	return info, nil
}

// initStringGlobals initializes NULL std::string globals to point to empty string.
// This fixes crashes in libstdc++ COW string code that expects globals to be
// initialized to the shared empty string representation.
func (e *Emulator) initStringGlobals(symbols map[string]uint64) {
	emptyDataPtr := e.GetEmptyStringData()
	if emptyDataPtr == 0 {
		return
	}

	// Initialize well-known std::string globals that are typically uninitialized
	// These are common in obfuscated Cocos2d-x binaries
	stringGlobals := []string{
		// Single-letter names (common obfuscation pattern)
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
		"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
		"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
		"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
		// Two-letter combos often used
		"a1", "b1", "c1", "d1", "e1",
		"AS", "TR", "co", "CO",
	}

	ptrBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(ptrBuf, emptyDataPtr)

	for _, name := range stringGlobals {
		if addr, ok := symbols[name]; ok && addr != 0 {
			// Check if the global is currently NULL
			data, err := e.MemRead(addr, 8)
			if err != nil {
				continue
			}
			val := binary.LittleEndian.Uint64(data)
			if val == 0 {
				// Initialize to point to empty string data
				_ = e.MemWrite(addr, ptrBuf)
			}
		}
	}
}

// addPLTSymbols adds PLT stub addresses for external symbols.
// This allows stubs to hook external function calls via their PLT entry.
// Addresses are added to both symbols (for lookups) and imports (for stub installation).
func addPLTSymbols(f *elf.File, relocOffset uint64, symbols, imports map[string]uint64) {
	// Find .plt section
	pltSec := f.Section(".plt")
	if pltSec == nil {
		return
	}

	// Find .rela.plt section
	relaPlt := f.Section(".rela.plt")
	if relaPlt == nil {
		return
	}

	// Get dynamic symbols (note: Go skips STN_UNDEF at index 0)
	dynSyms, err := f.DynamicSymbols()
	if err != nil {
		return
	}

	// Read .rela.plt data
	relaData, err := relaPlt.Data()
	if err != nil {
		return
	}

	// ARM64 PLT structure:
	// - PLT header: 32 bytes
	// - Each PLT entry: 16 bytes
	pltBase := pltSec.Addr + relocOffset
	const pltHeaderSize = 32
	const pltEntrySize = 16

	// Each RELA entry is 24 bytes
	entryIdx := 0
	for i := 0; i+24 <= len(relaData); i += 24 {
		rInfo := binary.LittleEndian.Uint64(relaData[i+8:])
		symIdx := int(rInfo >> 32)

		// Adjust for Go skipping STN_UNDEF (symIdx is 1-based in ELF, but array is 0-based)
		arrayIdx := symIdx - 1
		if arrayIdx < 0 || arrayIdx >= len(dynSyms) {
			entryIdx++
			continue
		}

		sym := dynSyms[arrayIdx]
		if sym.Name == "" {
			entryIdx++
			continue
		}

		// Only add for external symbols (value == 0)
		if sym.Value == 0 {
			pltAddr := pltBase + pltHeaderSize + uint64(entryIdx)*pltEntrySize
			// Store PLT address in both maps
			symbols[sym.Name] = pltAddr
			imports[sym.Name] = pltAddr
			// Also store without version suffix
			if idx := strings.Index(sym.Name, "@@"); idx != -1 {
				symbols[sym.Name[:idx]] = pltAddr
				imports[sym.Name[:idx]] = pltAddr
			} else if idx := strings.Index(sym.Name, "@"); idx != -1 {
				symbols[sym.Name[:idx]] = pltAddr
				imports[sym.Name[:idx]] = pltAddr
			}
		}

		entryIdx++
	}
}

// applyRelocations processes ELF relocations to fix GOT entries.
// The imports map provides PLT stub addresses for external symbols (needed for R_AARCH64_ABS64).
func (e *Emulator) applyRelocations(f *elf.File, relocOffset uint64, imports map[string]uint64) error {
	// Build symbol table for lookups
	// NOTE: Go's DynamicSymbols() skips the first entry (STN_UNDEF at index 0),
	// so symIdx from relocations needs to be decremented by 1 for lookup.
	dynSyms, _ := f.DynamicSymbols()
	symByIndex := make(map[int]elf.Symbol)
	for i, sym := range dynSyms {
		// Store at i+1 to match ELF symbol indices (which include STN_UNDEF at 0)
		symByIndex[i+1] = sym
	}

	// Process .rela.dyn section
	for _, sec := range f.Sections {
		if sec.Type != elf.SHT_RELA {
			continue
		}
		if sec.Name != ".rela.dyn" && sec.Name != ".rela.plt" {
			continue
		}

		data, err := sec.Data()
		if err != nil {
			continue
		}

		// Each RELA entry is 24 bytes: r_offset (8), r_info (8), r_addend (8)
		entrySize := 24
		for i := 0; i+entrySize <= len(data); i += entrySize {
			rOffset := binary.LittleEndian.Uint64(data[i:])
			rInfo := binary.LittleEndian.Uint64(data[i+8:])
			rAddend := int64(binary.LittleEndian.Uint64(data[i+16:]))

			relType := uint32(rInfo & 0xFFFFFFFF)
			symIdx := int(rInfo >> 32)

			targetAddr := rOffset + relocOffset

			switch relType {
			case R_AARCH64_RELATIVE:
				// *target = base + addend
				resolved := relocOffset + uint64(rAddend)
				buf := make([]byte, 8)
				binary.LittleEndian.PutUint64(buf, resolved)
				_ = e.MemWrite(targetAddr, buf)

			case R_AARCH64_GLOB_DAT, R_AARCH64_JUMP_SLOT:
				// *target = base + symbol.st_value
				// JUMP_SLOT is used for PLT GOT entries - resolve to actual function address
				if sym, ok := symByIndex[symIdx]; ok {
					if sym.Value != 0 {
						resolved := sym.Value + relocOffset
						buf := make([]byte, 8)
						binary.LittleEndian.PutUint64(buf, resolved)
						_ = e.MemWrite(targetAddr, buf)
					} else if sym.Name == "__stack_chk_guard" {
						// External symbol from libc - point to our TLS canary
						// The canary value is at TLS+0x28
						canaryAddr := uint64(TLSBase + 0x28)
						buf := make([]byte, 8)
						binary.LittleEndian.PutUint64(buf, canaryAddr)
						_ = e.MemWrite(targetAddr, buf)
					} else if sym.Name == "_ctype_" {
						// External symbol from libc - point to our mock ctype table
						// The table is at LibcBase + CtypeTableOffset + 1 (index -1 starts at offset 0)
						ctypeAddr := uint64(LibcBase + CtypeTableOffset + 1)
						buf := make([]byte, 8)
						binary.LittleEndian.PutUint64(buf, ctypeAddr)
						_ = e.MemWrite(targetAddr, buf)
					}
				}

			case R_AARCH64_ABS64:
				// *target = base + symbol.st_value + addend
				// For internal symbols (st_value > 0): resolve directly
				// For external symbols (st_value == 0): resolve to PLT stub address
				if sym, ok := symByIndex[symIdx]; ok {
					if sym.Value != 0 {
						// Internal symbol - use symbol value
						resolved := sym.Value + relocOffset + uint64(rAddend)
						buf := make([]byte, 8)
						binary.LittleEndian.PutUint64(buf, resolved)
						_ = e.MemWrite(targetAddr, buf)
					} else if sym.Name != "" {
						// External symbol - resolve to PLT stub (Unity IL2CPP uses this for malloc, etc.)
						// Strip version suffix for lookup
						symName := sym.Name
						if idx := strings.Index(symName, "@@"); idx != -1 {
							symName = symName[:idx]
						} else if idx := strings.Index(symName, "@"); idx != -1 {
							symName = symName[:idx]
						}
						if stubAddr, ok := imports[symName]; ok {
							resolved := stubAddr + uint64(rAddend)
							buf := make([]byte, 8)
							binary.LittleEndian.PutUint64(buf, resolved)
							_ = e.MemWrite(targetAddr, buf)
						}
					}
				} else if rAddend > 0 {
					// No symbol, just base + addend
					resolved := relocOffset + uint64(rAddend)
					buf := make([]byte, 8)
					binary.LittleEndian.PutUint64(buf, resolved)
					_ = e.MemWrite(targetAddr, buf)
				}
			}
		}
	}

	return nil
}

// FindSymbol looks up a symbol by name, returns 0 if not found
func (info *ELFInfo) FindSymbol(name string) uint64 {
	return info.Symbols[name]
}

// FindJNIOnLoad returns the address of JNI_OnLoad or 0
func (info *ELFInfo) FindJNIOnLoad() uint64 {
	// Try exact match first
	if addr := info.Symbols["JNI_OnLoad"]; addr != 0 {
		return addr
	}

	// Case-insensitive search
	for name, addr := range info.Symbols {
		if strings.EqualFold(name, "JNI_OnLoad") {
			return addr
		}
	}

	return 0
}

// FindEntryPoint finds a good entry point for emulation.
// Priority matches Python galago.py for Cocos2d-x key extraction:
// 1. Preferred entry (user-specified)
// 2. applicationDidFinishLaunching (best for Cocos2d-x - directly sets keys)
// 3. cocos_android_app_init (alternative entry point)
// 4. JNI_OnLoad (most common for Android .so, but may need chained calls)
// 5. cocos_main, Game::init (Cocos Creator 3.x)
// 6. Generic init patterns
func (info *ELFInfo) FindEntryPoint(preferredEntry string) uint64 {
	// Check preferred entry first
	if preferredEntry != "" {
		if addr := info.FindSymbol(preferredEntry); addr != 0 {
			return addr
		}
		// Case-insensitive search
		for name, addr := range info.Symbols {
			if strings.EqualFold(name, preferredEntry) {
				return addr
			}
		}
		// Substring search
		lower := strings.ToLower(preferredEntry)
		for name, addr := range info.Symbols {
			if strings.Contains(strings.ToLower(name), lower) {
				return addr
			}
		}
	}

	// Build list of candidate entry points with priorities
	type candidate struct {
		name     string
		addr     uint64
		priority int
	}
	var candidates []candidate

	for name, addr := range info.Symbols {
		if addr == 0 {
			continue
		}
		lower := strings.ToLower(name)

		// Exclude vtables, typeinfo, and other non-function symbols
		if strings.Contains(lower, "_ztv") || strings.Contains(lower, "_zti") ||
			strings.Contains(lower, "_zts") || strings.Contains(lower, "__func") ||
			strings.Contains(lower, "__clone") || strings.Contains(lower, "__target") ||
			strings.Contains(lower, "destroy") || strings.Contains(lower, "deallocate") {
			continue
		}

		// Priority 0: regist_lua (Lua games - direct key setup)
		if strings.Contains(lower, "regist_lua") {
			candidates = append(candidates, candidate{name, addr, 0})
			continue
		}
		// Priority 1: AppDelegate::applicationDidFinishLaunching (most reliable for key extraction)
		if strings.Contains(lower, "appdelegate") && strings.Contains(lower, "didfinish") {
			candidates = append(candidates, candidate{name, addr, 1})
			continue
		}
		// Priority 2: CCGameMain::applicationDidFinishLaunching (Lua games - less reliable)
		if strings.Contains(lower, "ccgamemain") && strings.Contains(lower, "didfinish") {
			candidates = append(candidates, candidate{name, addr, 2})
			continue
		}
		// Priority 3: Generic applicationDidFinishLaunching
		if strings.Contains(lower, "didfinishlaunching") {
			candidates = append(candidates, candidate{name, addr, 3})
			continue
		}
		// Priority 4: cocos_android_app_init
		if strings.Contains(lower, "cocos_android_app_init") {
			candidates = append(candidates, candidate{name, addr, 4})
			continue
		}
		// Priority 5: cocos_main (Cocos Creator 3.x)
		if strings.Contains(lower, "cocos_main") {
			candidates = append(candidates, candidate{name, addr, 5})
			continue
		}
		// Priority 6: Game::init (Cocos Creator 3.x) - _ZN4Game...initEv
		if strings.HasPrefix(lower, "_zn4game") && strings.Contains(lower, "initev") {
			candidates = append(candidates, candidate{name, addr, 6})
			continue
		}
		// Priority 7: JNI_OnLoad
		if strings.EqualFold(name, "JNI_OnLoad") {
			candidates = append(candidates, candidate{name, addr, 7})
			continue
		}
	}

	// Sort by priority and return best
	if len(candidates) > 0 {
		best := candidates[0]
		for _, c := range candidates[1:] {
			if c.priority < best.priority {
				best = c
			}
		}
		return best.addr
	}

	// Fallback to JNI_OnLoad
	if addr := info.FindJNIOnLoad(); addr != 0 {
		return addr
	}

	// Fallback to ELF entry point
	return info.Entry
}

// FindSymbolsMatching returns all symbols matching a predicate
func (info *ELFInfo) FindSymbolsMatching(predicate func(name string) bool) map[string]uint64 {
	result := make(map[string]uint64)
	for name, addr := range info.Symbols {
		if predicate(name) {
			result[name] = addr
		}
	}
	return result
}

// FindSymbolsBySubstring finds symbols containing the given substring
func (info *ELFInfo) FindSymbolsBySubstring(substr string) map[string]uint64 {
	return info.FindSymbolsMatching(func(name string) bool {
		return strings.Contains(strings.ToLower(name), strings.ToLower(substr))
	})
}

// IsExecutable returns true if the segment is executable
func (s *Segment) IsExecutable() bool {
	return s.Flags&elf.PF_X != 0
}

// IsWritable returns true if the segment is writable
func (s *Segment) IsWritable() bool {
	return s.Flags&elf.PF_W != 0
}

// IsReadable returns true if the segment is readable
func (s *Segment) IsReadable() bool {
	return s.Flags&elf.PF_R != 0
}
