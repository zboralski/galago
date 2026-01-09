// Package emulator provides vtable resolution for C++ virtual function calls.
// This implements Itanium C++ ABI vtable resolution using ELF relocations.
package emulator

import (
	"debug/elf"
	"encoding/binary"
	"sort"
	"strings"
)

// VTable represents a C++ virtual function table
type VTable struct {
	Name       string              // Mangled vtable symbol name (e.g., _ZTVN7cocos2d8LuaStackE)
	ClassName  string              // Demangled class name (e.g., cocos2d::LuaStack)
	Start      uint64              // Relocated vtable base address
	Size       uint64              // Vtable size in bytes
	Slots      map[uint64]SlotInfo // Byte offset from Start -> resolved function info
}

// SlotInfo contains information about a resolved vtable slot
type SlotInfo struct {
	Target    uint64 // Resolved function address
	SymName   string // Symbol name if known
	RelocType uint32 // Relocation type that populated this slot
	SlotIndex int    // Logical slot index (0 = first virtual function after RTTI)
}

// VTableMap maps vtable base addresses to VTable info
type VTableMap struct {
	Tables    map[uint64]*VTable   // vtable base address -> VTable
	ByClass   map[string]*VTable   // class name -> VTable (for convenience)
	SlotIndex map[uint64][]SlotInfo // slot byte offset -> all candidates (for matching)
}

// BuildVTableMap parses ELF relocations and builds a map of vtable slot -> function.
// This implements Android/linker64 relocation resolution for vtable entries.
func BuildVTableMap(f *elf.File, relocOffset uint64) (*VTableMap, error) {
	vtm := &VTableMap{
		Tables:    make(map[uint64]*VTable),
		ByClass:   make(map[string]*VTable),
		SlotIndex: make(map[uint64][]SlotInfo),
	}

	// 1. Collect vtable symbols and their ranges
	var vtSyms []elf.Symbol
	addVt := func(s elf.Symbol) {
		if s.Value != 0 && strings.HasPrefix(s.Name, "_ZTV") {
			vtSyms = append(vtSyms, s)
		}
	}

	if syms, _ := f.DynamicSymbols(); syms != nil {
		for _, s := range syms {
			addVt(s)
		}
	}
	if syms, _ := f.Symbols(); syms != nil {
		for _, s := range syms {
			addVt(s)
		}
	}

	// Sort by address for range detection
	sort.Slice(vtSyms, func(i, j int) bool { return vtSyms[i].Value < vtSyms[j].Value })

	// Build vtable ranges
	type vtRange struct {
		name       string
		className  string
		start, end uint64
	}
	var ranges []vtRange
	for i, s := range vtSyms {
		start := s.Value + relocOffset
		end := start + s.Size
		// If size is 0, estimate end from next vtable or use fallback
		if s.Size == 0 && i+1 < len(vtSyms) {
			end = vtSyms[i+1].Value + relocOffset
		} else if s.Size == 0 {
			end = start + 0x400 // Fallback: 128 slots max
		}
		className := extractClassName(s.Name)
		ranges = append(ranges, vtRange{s.Name, className, start, end})
	}

	// Helper to find which vtable range contains an address
	findRange := func(addr uint64) *vtRange {
		for i := range ranges {
			if addr >= ranges[i].start && addr < ranges[i].end {
				return &ranges[i]
			}
		}
		return nil
	}

	// 2. Build symbol index for ABS64 lookups
	dynSyms, _ := f.DynamicSymbols()
	symByIdx := make(map[int]elf.Symbol)
	for i, s := range dynSyms {
		symByIdx[i+1] = s // ELF symbol indices are 1-based (0 is STN_UNDEF)
	}

	// 3. Build reverse symbol map: address -> symbol name (for resolving R_AARCH64_RELATIVE targets)
	addrToSym := make(map[uint64]string)
	addSymToMap := func(s elf.Symbol) {
		if s.Value != 0 && s.Name != "" && s.Info&0xf == 2 { // STT_FUNC
			addr := s.Value + relocOffset
			if _, exists := addrToSym[addr]; !exists {
				addrToSym[addr] = cleanSymbolName(s.Name)
			}
		}
	}
	for _, s := range dynSyms {
		addSymToMap(s)
	}
	if staticSyms, _ := f.Symbols(); staticSyms != nil {
		for _, s := range staticSyms {
			addSymToMap(s)
		}
	}

	// 3. Process relocations and populate vtable slots
	for _, sec := range f.Sections {
		if sec.Type != elf.SHT_RELA {
			continue
		}
		data, err := sec.Data()
		if err != nil {
			continue
		}

		// Parse RELA entries (24 bytes each on ARM64)
		for i := 0; i+24 <= len(data); i += 24 {
			rOffset := binary.LittleEndian.Uint64(data[i:])
			rInfo := binary.LittleEndian.Uint64(data[i+8:])
			rAddend := int64(binary.LittleEndian.Uint64(data[i+16:]))

			relType := uint32(rInfo & 0xffffffff)
			symIdx := int(rInfo >> 32)

			targetAddr := rOffset + relocOffset

			// Check if this relocation falls inside a vtable
			vt := findRange(targetAddr)
			if vt == nil {
				continue
			}

			// Resolve function pointer using linker logic
			var resolved uint64
			switch relType {
			case R_AARCH64_RELATIVE:
				// base + addend
				resolved = relocOffset + uint64(rAddend)
			case R_AARCH64_ABS64:
				// symbol + addend
				if sym, ok := symByIdx[symIdx]; ok && sym.Value != 0 {
					resolved = sym.Value + relocOffset + uint64(rAddend)
				} else {
					resolved = relocOffset + uint64(rAddend)
				}
			case R_AARCH64_GLOB_DAT, R_AARCH64_JUMP_SLOT:
				// symbol value
				if sym, ok := symByIdx[symIdx]; ok && sym.Value != 0 {
					resolved = sym.Value + relocOffset
				}
			default:
				continue
			}

			if resolved == 0 {
				continue
			}

			// Calculate slot offset within vtable
			slotOffset := targetAddr - vt.start

			// Itanium ABI: first 16 bytes are offset_to_top (8) + RTTI pointer (8)
			// Function pointers start at offset 16
			var slotIndex int
			if slotOffset >= 16 {
				slotIndex = int((slotOffset - 16) / 8)
			} else {
				slotIndex = -1 // RTTI/metadata area
			}

			// Get or create VTable entry
			tbl := vtm.Tables[vt.start]
			if tbl == nil {
				tbl = &VTable{
					Name:      vt.name,
					ClassName: vt.className,
					Start:     vt.start,
					Size:      vt.end - vt.start,
					Slots:     make(map[uint64]SlotInfo),
				}
				vtm.Tables[vt.start] = tbl
				if vt.className != "" {
					vtm.ByClass[vt.className] = tbl
				}
			}

			// Get symbol name: first try relocation symbol, then reverse lookup by target address
			symName := ""
			if sym, ok := symByIdx[symIdx]; ok && sym.Name != "" {
				symName = cleanSymbolName(sym.Name)
			}
			// For R_AARCH64_RELATIVE, lookup target address in symbol table
			if symName == "" {
				if name, ok := addrToSym[resolved]; ok {
					symName = name
				}
			}

			slotInfo := SlotInfo{
				Target:    resolved,
				SymName:   symName,
				RelocType: relType,
				SlotIndex: slotIndex,
			}
			tbl.Slots[slotOffset] = slotInfo

			// Also index by slot offset for quick lookup during emulation
			vtm.SlotIndex[slotOffset] = append(vtm.SlotIndex[slotOffset], slotInfo)
		}
	}

	return vtm, nil
}

// extractClassName extracts class name from mangled vtable symbol.
// _ZTVN7cocos2d8LuaStackE -> cocos2d::LuaStack
func extractClassName(mangledName string) string {
	if !strings.HasPrefix(mangledName, "_ZTV") {
		return ""
	}

	// Handle nested names: _ZTVN<parts>E
	if strings.HasPrefix(mangledName, "_ZTVN") {
		rest := mangledName[5:] // Skip _ZTVN
		return parseNestedName(rest)
	}

	// Handle simple names: _ZTV<length><name>
	rest := mangledName[4:] // Skip _ZTV
	if len(rest) > 0 && rest[0] >= '1' && rest[0] <= '9' {
		length, name := parseLengthPrefixedName(rest)
		if length > 0 {
			return name
		}
	}

	return ""
}

// parseNestedName parses Itanium nested name encoding
func parseNestedName(s string) string {
	parts := []string{}
	rest := s

	for len(rest) > 0 && rest[0] != 'E' {
		// Skip template arguments
		if rest[0] == 'I' {
			break
		}

		// Parse length-prefixed component
		length, name := parseLengthPrefixedName(rest)
		if length > 0 && name != "" {
			parts = append(parts, name)
			rest = rest[length:]
			// Skip the parsed name
			if len(rest) >= len(name) {
				// Already consumed by parseLengthPrefixedName
			}
		} else {
			break
		}

		// Advance past the name we just parsed
		if len(rest) > 0 {
			consumed := 0
			for consumed < len(rest) && rest[consumed] >= '0' && rest[consumed] <= '9' {
				consumed++
			}
			if consumed > 0 && consumed < len(rest) {
				numLen := 0
				for _, c := range rest[:consumed] {
					numLen = numLen*10 + int(c-'0')
				}
				if consumed+numLen <= len(rest) {
					rest = rest[consumed+numLen:]
				} else {
					break
				}
			} else {
				break
			}
		}
	}

	if len(parts) > 0 {
		return strings.Join(parts, "::")
	}
	return ""
}

// parseLengthPrefixedName parses a length-prefixed name like "7cocos2d"
// Returns (total bytes consumed, name)
func parseLengthPrefixedName(s string) (int, string) {
	if len(s) == 0 || s[0] < '1' || s[0] > '9' {
		return 0, ""
	}

	// Parse multi-digit length
	i := 0
	length := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		length = length*10 + int(s[i]-'0')
		i++
	}

	if i+length > len(s) {
		return 0, ""
	}

	return i + length, s[i : i+length]
}

// cleanSymbolName removes version suffixes from symbol names
func cleanSymbolName(name string) string {
	if idx := strings.Index(name, "@@"); idx != -1 {
		return name[:idx]
	}
	if idx := strings.Index(name, "@"); idx != -1 {
		return name[:idx]
	}
	return name
}

// ResolveVirtualCall resolves a virtual function call through the vtable.
// Given a vtable base and slot index, returns the resolved function address and symbol name.
func (vtm *VTableMap) ResolveVirtualCall(vtableBase uint64, slotIndex int) (uint64, string, bool) {
	tbl, ok := vtm.Tables[vtableBase]
	if !ok {
		return 0, "", false
	}

	// Calculate slot offset: skip RTTI header (16 bytes), then 8 bytes per slot
	slotOffset := uint64(16 + slotIndex*8)

	slot, ok := tbl.Slots[slotOffset]
	if !ok {
		return 0, "", false
	}

	return slot.Target, slot.SymName, true
}

// ResolveBySlotOffset resolves using byte offset from vtable base.
// This is useful when you know the raw offset from LDR instructions.
func (vtm *VTableMap) ResolveBySlotOffset(vtableBase, slotOffset uint64) (uint64, string, bool) {
	tbl, ok := vtm.Tables[vtableBase]
	if !ok {
		return 0, "", false
	}

	slot, ok := tbl.Slots[slotOffset]
	if !ok {
		return 0, "", false
	}

	return slot.Target, slot.SymName, true
}

// FindSetterSlots finds all vtable slots that resolve to setter functions.
// Returns a map of (vtable_base, slot_offset) -> SlotInfo for setter functions.
func (vtm *VTableMap) FindSetterSlots(setterPatterns []string) map[uint64]map[uint64]SlotInfo {
	result := make(map[uint64]map[uint64]SlotInfo)

	for vtBase, tbl := range vtm.Tables {
		for slotOff, slot := range tbl.Slots {
			if isSetterSymbol(slot.SymName, setterPatterns) {
				if result[vtBase] == nil {
					result[vtBase] = make(map[uint64]SlotInfo)
				}
				result[vtBase][slotOff] = slot
			}
		}
	}

	return result
}

// isSetterSymbol checks if a symbol name matches setter patterns
func isSetterSymbol(symName string, patterns []string) bool {
	if symName == "" {
		return false
	}
	lower := strings.ToLower(symName)
	for _, p := range patterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}

// GetAllSetterTargets returns all function addresses that are setter functions
// as resolved through vtables. This can be used to install direct hooks.
func (vtm *VTableMap) GetAllSetterTargets(setterPatterns []string) map[uint64]string {
	result := make(map[uint64]string)

	for _, tbl := range vtm.Tables {
		for _, slot := range tbl.Slots {
			if isSetterSymbol(slot.SymName, setterPatterns) {
				result[slot.Target] = slot.SymName
			}
		}
	}

	return result
}
