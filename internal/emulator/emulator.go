// Package emulator provides ARM64 emulation using Unicorn Engine.
package emulator

import (
	"encoding/binary"
	"fmt"
	"sync"

	uc "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
	"github.com/zboralski/galago/internal/hipaa"
)

// Memory layout constants
const (
	CodeBase    = 0x00010000
	CodeSize    = 0x01000000 // 16MB for code
	StackBase   = 0x80000000
	StackSize   = 0x00100000 // 1MB stack
	HeapBase    = 0x90000000
	HeapSize    = 0x10000000 // 256MB heap
	MockObjBase = 0xDEAB0000 // Mock C++ objects (this pointers)
	MockObjSize = 0x00010000 // 64KB for mock objects
	TLSBase     = 0xDEAC0000 // Thread Local Storage
	TLSSize     = 0x00010000 // 64KB TLS
	LibcBase    = 0xDEAD0000 // Mock libc globals (_ctype_, etc.)
	LibcSize    = 0x00010000 // 64KB for libc data
	StubBase    = 0xF0000000 // Stub functions mapped here
	StubSize    = 0x00100000 // 1MB for stubs
)

// Libc global layout
const (
	CtypeTableOffset   uint64 = 0x0000 // _ctype_ table: 257 bytes (index -1 to 255)
	CtypePtrOffset     uint64 = 0x0200 // _ctype_ pointer (points to CtypeTable+1)
	EmptyStringRepOff  uint64 = 0x0300 // libstdc++ COW empty string _Rep
	EmptyStringDataOff uint64 = 0x0318 // Empty string data pointer (Rep + 24)
)

// HookType identifies different hook categories
type HookType int

const (
	HookCode HookType = iota
	HookMem
	HookBlock
	HookIntr
)

// TraceEvent represents a single traced instruction
type TraceEvent struct {
	Address     uint64
	Size        uint32
	Instruction string // Disassembled (if available)
	Tag         string // Hashtag like #xor-neon
	Detail      string // Additional context
}

// CodeHookFunc is called for each instruction
type CodeHookFunc func(emu *Emulator, addr uint64, size uint32)

// AddressHookFunc is called when execution reaches a specific address
type AddressHookFunc func(emu *Emulator) bool // return true to stop emulation

// Emulator wraps Unicorn for ARM64 emulation
type Emulator struct {
	mu uc.Unicorn

	// Memory management
	heapPtr uint64 // Current heap allocation pointer

	// Hooks
	codeHooks   []CodeHookFunc
	addrHooks   map[uint64]AddressHookFunc
	addrHooksMu sync.RWMutex

	// Trace collection
	traceEnabled bool
	traceEvents  []TraceEvent
	traceMu      sync.Mutex

	// Stop flag
	stopped bool

	// libstdc++ COW empty string data pointer
	emptyStringData uint64

	// HIPAA compliance components
	detector  *hipaa.Detector
	encryptor *hipaa.Encryptor
	auditor   *hipaa.Auditor
	sanitizer *hipaa.Sanitizer
}

// New creates a new ARM64 emulator
func New() (*Emulator, error) {
	mu, err := uc.NewUnicorn(uc.ARCH_ARM64, uc.MODE_ARM)
	if err != nil {
		return nil, fmt.Errorf("create unicorn: %w", err)
	}

	emu := &Emulator{
		mu:        mu,
		heapPtr:   HeapBase,
		addrHooks: make(map[uint64]AddressHookFunc),
	}

	// Initialize HIPAA compliance components
	// As a healthcare professional, I emphasize the importance of protecting patient data from the start.
	emu.detector = hipaa.NewDetector()
	emu.auditor = hipaa.NewAuditor(true)
	emu.sanitizer = hipaa.NewSanitizer(emu.detector)
	encryptor, err := hipaa.NewEncryptor(emu.detector, emu.auditor)
	if err != nil {
		mu.Close()
		return nil, fmt.Errorf("create encryptor: %w", err)
	}
	emu.encryptor = encryptor
	hipaa.SessionEncryptor = encryptor
	hipaa.SessionDetector = emu.detector
	hipaa.SessionAuditor = emu.auditor

	// Map memory regions
	if err := emu.mapMemory(); err != nil {
		mu.Close()
		return nil, err
	}

	// Set up internal hooks
	if err := emu.setupHooks(); err != nil {
		mu.Close()
		return nil, err
	}

	return emu, nil
}

// mapMemory sets up the memory layout
func (e *Emulator) mapMemory() error {
	regions := []struct {
		base uint64
		size uint64
		name string
	}{
		{CodeBase, CodeSize, "code"},
		{StackBase, StackSize, "stack"},
		{HeapBase, HeapSize, "heap"},
		{MockObjBase, MockObjSize, "mockobj"}, // Mock C++ objects for this pointers
		{TLSBase, TLSSize, "tls"},
		{LibcBase, LibcSize, "libc"}, // Mock libc globals (_ctype_, etc.)
		{StubBase, StubSize, "stubs"},
	}

	for _, r := range regions {
		if err := e.mu.MemMap(r.base, r.size); err != nil {
			return fmt.Errorf("map %s (0x%x): %w", r.name, r.base, err)
		}
	}

	// Initialize stack pointer
	sp := uint64(StackBase + StackSize - 0x1000)
	if err := e.mu.RegWrite(uc.ARM64_REG_SP, sp); err != nil {
		return fmt.Errorf("set SP: %w", err)
	}

	// Pre-initialize stack with mock vtable pointers.
	// This helps with C++ code that does dynamic_cast on objects constructed on stack.
	// Without this, reads from uninitialized stack get garbage which causes RTTI crashes.
	// The mock vtable (at MockObjBase + 0x1010 after RTTI setup) has valid RTTI prefix.
	mockVtableAddr := uint64(MockObjBase + 0x1010)
	stackFillPattern := make([]byte, 8)
	binary.LittleEndian.PutUint64(stackFillPattern, mockVtableAddr)

	// Fill the used portion of stack (from SP down to StackBase + some margin)
	// We use a pattern of repeating the mock vtable pointer every 8 bytes
	stackFillSize := uint64(0x10000) // 64KB of stack initialization
	stackFillBuf := make([]byte, stackFillSize)
	for i := uint64(0); i < stackFillSize; i += 8 {
		copy(stackFillBuf[i:i+8], stackFillPattern)
	}
	// Fill from StackBase + StackSize - stackFillSize
	stackFillBase := StackBase + StackSize - stackFillSize
	_ = e.mu.MemWrite(stackFillBase, stackFillBuf)
	// Note: ignore error - best effort initialization

	// Initialize TLS (Thread Local Storage)
	// TPIDR_EL0 is the thread pointer register on ARM64
	if err := e.mu.RegWrite(uc.ARM64_REG_TPIDR_EL0, TLSBase); err != nil {
		return fmt.Errorf("set TPIDR_EL0: %w", err)
	}

	// Initialize TLS area with zeros
	zeros := make([]byte, 256)
	if err := e.mu.MemWrite(TLSBase, zeros); err != nil {
		return fmt.Errorf("init TLS: %w", err)
	}

	// Set up stack canary at TLS+0x28 (used by ARM64 for stack protection)
	// Use a deterministic value for reproducible emulation (0xDEADBEEF is Unix tradition since 1988)
	canary := uint64(0xDEADBEEFDEADBEEF)
	canaryBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(canaryBytes, canary)
	if err := e.mu.MemWrite(TLSBase+0x28, canaryBytes); err != nil {
		return fmt.Errorf("set stack canary: %w", err)
	}

	// Initialize libc globals (_ctype_ table for character classification)
	// The _ctype_ table is 257 bytes: index -1 (EOF=0) through 255
	// Character classes (matching POSIX/bionic):
	//   _U = 0x01 (upper)    _L = 0x02 (lower)   _N = 0x04 (digit)
	//   _S = 0x08 (space)    _P = 0x10 (punct)   _C = 0x20 (control)
	//   _B = 0x40 (blank)    _X = 0x80 (hex)
	ctypeTable := make([]byte, 257)
	ctypeTable[0] = 0 // EOF (-1 offset becomes index 0)
	for i := 0; i < 256; i++ {
		c := byte(i)
		var flags byte
		switch {
		case c >= 'A' && c <= 'Z':
			flags = 0x01 | 0x80 // _U, and _X for A-F
			if c > 'F' {
				flags = 0x01
			}
		case c >= 'a' && c <= 'z':
			flags = 0x02 | 0x80 // _L, and _X for a-f
			if c > 'f' {
				flags = 0x02
			}
		case c >= '0' && c <= '9':
			flags = 0x04 | 0x80 // _N | _X
		case c == ' ':
			flags = 0x08 | 0x40 // _S | _B (space is blank)
		case c == '\t':
			flags = 0x08 | 0x40 // _S | _B (tab is blank)
		case c == '\n' || c == '\r' || c == '\f' || c == '\v':
			flags = 0x08 // _S (whitespace but not blank)
		case c < 0x20 || c == 0x7F:
			flags = 0x20 // _C (control)
		case c >= 0x21 && c <= 0x2F: // !"#$%&'()*+,-./
			flags = 0x10 // _P
		case c >= 0x3A && c <= 0x40: // :;<=>?@
			flags = 0x10 // _P
		case c >= 0x5B && c <= 0x60: // [\]^_`
			flags = 0x10 // _P
		case c >= 0x7B && c <= 0x7E: // {|}~
			flags = 0x10 // _P
		}
		ctypeTable[i+1] = flags // +1 because index 0 is EOF
	}
	if err := e.mu.MemWrite(LibcBase+CtypeTableOffset, ctypeTable); err != nil {
		return fmt.Errorf("init _ctype_ table: %w", err)
	}

	// Set up _ctype_ pointer to point to table[1] (so index -1 works)
	ctypePtr := LibcBase + CtypeTableOffset + 1
	ctypePtrBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(ctypePtrBytes, ctypePtr)
	if err := e.mu.MemWrite(LibcBase+CtypePtrOffset, ctypePtrBytes); err != nil {
		return fmt.Errorf("init _ctype_ pointer: %w", err)
	}

	// Set up libstdc++ COW empty string representation
	// Layout of _Rep: { size_t _M_length, size_t _M_capacity, atomic<int> _M_refcount }
	// Empty string: length=0, capacity=0, refcount=-1 (immortal)
	// Data pointer = Rep + 24 bytes, points to null terminator
	emptyRep := make([]byte, 32) // 24 bytes for _Rep + 8 bytes for data (null terminator + padding)
	// _M_length = 0 (bytes 0-7, already zero)
	// _M_capacity = 0 (bytes 8-15, already zero)
	// _M_refcount = -1 (bytes 16-19), leave 20-23 as padding
	emptyRep[16] = 0xFF
	emptyRep[17] = 0xFF
	emptyRep[18] = 0xFF
	emptyRep[19] = 0xFF
	// Data area starts at offset 24 - null terminator
	emptyRep[24] = 0 // null terminator
	if err := e.mu.MemWrite(LibcBase+EmptyStringRepOff, emptyRep); err != nil {
		return fmt.Errorf("init empty string rep: %w", err)
	}
	// Store the empty string address for external reference
	e.emptyStringData = LibcBase + EmptyStringDataOff

	// Initialize mock object region for C++ this pointers
	// Layout (matching Python extract_key.py):
	//   MockObjBase + 0x0800 = mock_typeinfo (type_info for RTTI)
	//   MockObjBase + 0x1000 = mock_vtable_prefix (RTTI prefix: offset_to_top + type_info*)
	//   MockObjBase + 0x1010 = mock_vtable (pointer table for virtual calls, vtable[0] at +0x10)
	//   MockObjBase + 0x2000 = mock_obj (main mock C++ object)
	//   MockObjBase + 0x3000 = mock_obj2 (secondary mock for member pointers)
	//   MockObjBase + 0x4000 = vtable_stubs (RET instructions with hooks for vtable[N])
	//   MockObjBase + 0x5000 = callback_stubs (RET instructions for indirect calls via member pointers)
	//
	// Design:
	// - mock_obj[0] = mock_vtable (for virtual calls)
	// - mock_obj[8..2048] = mock_obj2 (member pointers point to secondary mock)
	// - mock_obj2[0] = mock_vtable (for nested virtual calls)
	// - mock_obj2[8..2048] = callback_stubs (for double-dereferenced function pointers)
	// - mock_obj2 also has RET stubs at 4-byte intervals for direct calls
	//
	// RTTI layout (Itanium C++ ABI):
	// - vtable[-16]: offset_to_top (8 bytes) = 0
	// - vtable[-8]:  RTTI pointer (type_info*) (8 bytes) -> mock_typeinfo
	// - vtable[0]:   first virtual function pointer
	mockTypeInfo := uint64(MockObjBase + 0x0800)
	mockVtablePrefix := uint64(MockObjBase + 0x1000)
	mockVtable := uint64(MockObjBase + 0x1010) // Skip 16 bytes for RTTI prefix
	mockObj := uint64(MockObjBase + 0x2000)
	mockObj2 := uint64(MockObjBase + 0x3000)
	vtableStubs := uint64(MockObjBase + 0x4000)
	callbackStubs := uint64(MockObjBase + 0x5000)

	// ARM64 RET = 0xd65f03c0
	retInsn := []byte{0xc0, 0x03, 0x5f, 0xd6}
	stubAddrBytes := make([]byte, 8)

	// 0. Set up RTTI structures for dynamic_cast support
	// This allows code that does dynamic_cast on mock objects to not crash.
	//
	// type_info layout (Itanium ABI):
	//   +0: vtable pointer (points to __class_type_info vtable)
	//   +8: __name (pointer to null-terminated mangled name)
	//
	// We create a minimal type_info that looks valid but returns false for all casts.
	// The RTTI prefix (offset_to_top, type_info*) is placed BEFORE mockVtable.
	typeInfoName := mockTypeInfo + 0x100 // Name string at offset 0x100
	nameBytes := []byte("12_MockObject\x00") // Mangled name with null terminator
	if err := e.mu.MemWrite(typeInfoName, nameBytes); err != nil {
		return fmt.Errorf("write type_info name: %w", err)
	}

	// type_info structure at mockTypeInfo:
	// +0: vtable pointer (we use mockVtable itself as a fake vtable)
	// +8: pointer to name string
	typeInfoData := make([]byte, 16)
	binary.LittleEndian.PutUint64(typeInfoData[0:8], mockVtable) // fake vtable
	binary.LittleEndian.PutUint64(typeInfoData[8:16], typeInfoName)
	if err := e.mu.MemWrite(mockTypeInfo, typeInfoData); err != nil {
		return fmt.Errorf("write type_info: %w", err)
	}

	// RTTI prefix at mockVtablePrefix (immediately before mockVtable):
	// +0: offset_to_top (8 bytes) = 0
	// +8: RTTI pointer (type_info*) = mockTypeInfo
	rttiPrefix := make([]byte, 16)
	binary.LittleEndian.PutUint64(rttiPrefix[0:8], 0)            // offset_to_top = 0
	binary.LittleEndian.PutUint64(rttiPrefix[8:16], mockTypeInfo) // type_info pointer
	if err := e.mu.MemWrite(mockVtablePrefix, rttiPrefix); err != nil {
		return fmt.Errorf("write RTTI prefix: %w", err)
	}
	_ = mockVtablePrefix // used above

	// 1. Create vtable stubs (for vtable[N] virtual calls)
	for i := uint64(0); i < 256; i++ {
		stubAddr := vtableStubs + (i * 4)
		if err := e.mu.MemWrite(stubAddr, retInsn); err != nil {
			return fmt.Errorf("write vtable stub %d: %w", i, err)
		}
		binary.LittleEndian.PutUint64(stubAddrBytes, stubAddr)
		if err := e.mu.MemWrite(mockVtable+(i*8), stubAddrBytes); err != nil {
			return fmt.Errorf("write vtable entry %d: %w", i, err)
		}
		e.addrHooks[stubAddr] = makeVtableStubHook(e)
	}

	// 2. Create callback stubs (for indirect calls through member pointers)
	for i := uint64(0); i < 256; i++ {
		stubAddr := callbackStubs + (i * 4)
		if err := e.mu.MemWrite(stubAddr, retInsn); err != nil {
			return fmt.Errorf("write callback stub %d: %w", i, err)
		}
		e.addrHooks[stubAddr] = makeCallbackStubHook(e)
	}

	// 3. Make mock_obj2 callable - write RET stubs at 4-byte intervals and hook them
	// Code may load mock_obj2 as a function pointer and call it directly via blr
	for i := uint64(0); i < 256; i++ {
		addr := mockObj2 + (i * 4)
		if err := e.mu.MemWrite(addr, retInsn); err != nil {
			return fmt.Errorf("write mock_obj2 stub %d: %w", i, err)
		}
		e.addrHooks[addr] = makeMockObj2CallHook(e, mockObj)
	}

	// 4. Write vtable pointer at offset 0 of both mock objects
	vtableBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(vtableBytes, mockVtable)
	if err := e.mu.MemWrite(mockObj, vtableBytes); err != nil {
		return fmt.Errorf("init mock_obj vtable: %w", err)
	}
	// Note: mock_obj2's vtable pointer at offset 0 will be restored after member pointers

	// 5. Set up member pointer fields (offsets 8, 16, 24, ... up to 2048)
	// mock_obj member pointers -> mock_obj2 (for double dereference patterns)
	// mock_obj2 member pointers -> callback_stubs (for triple dereference / blr calls)
	obj2Bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(obj2Bytes, mockObj2)
	for i := uint64(1); i < 256; i++ {
		// mock_obj[i*8] = mock_obj2
		e.mu.MemWrite(mockObj+(i*8), obj2Bytes)
		// mock_obj2[i*8] = callback_stub[i % 256]
		callbackAddr := callbackStubs + ((i % 256) * 4)
		binary.LittleEndian.PutUint64(stubAddrBytes, callbackAddr)
		e.mu.MemWrite(mockObj2+(i*8), stubAddrBytes)
	}

	// 6. IMPORTANT: Restore vtable pointer at offset 0 of mock_obj2
	// (The above loop wrote callback_stub at offset 0, which we need to fix)
	if err := e.mu.MemWrite(mockObj2, vtableBytes); err != nil {
		return fmt.Errorf("restore mock_obj2 vtable: %w", err)
	}

	return nil
}

// GetMockObject returns the main mock C++ object address.
// Use this as the "this" pointer for member methods.
func (e *Emulator) GetMockObject() uint64 {
	return MockObjBase + 0x2000
}

// GetVtableStubs returns the base address for vtable stub functions.
// Vtable entries point to addresses starting at this base.
func (e *Emulator) GetVtableStubs() uint64 {
	return MockObjBase + 0x4000
}

// VtableStubCount is the number of vtable stub entries.
const VtableStubCount = 256

// GetCtypePtr returns the address of the _ctype_ pointer (points to classification table).
// Used by libc isXXX() functions and std::ctype.
func (e *Emulator) GetCtypePtr() uint64 {
	return LibcBase + CtypePtrOffset
}

// GetEmptyStringData returns the address of the libstdc++ COW empty string data.
// This is the data pointer that std::string globals should point to when uninitialized.
// Layout: Rep (24 bytes) + data. Data pointer = Rep + 24 = EmptyStringDataOff.
func (e *Emulator) GetEmptyStringData() uint64 {
	return e.emptyStringData
}

// makeVtableStubHook creates a hook for vtable stubs that handles return-by-value.
// When a virtual method is called on a mock object, this hook initializes
// the return buffer (X8) if it points to stack memory (for std::string return values).

func makeVtableStubHook(e *Emulator) AddressHookFunc {
	return func(emu *Emulator) bool {
		_ = emu.PC()
		_ = emu.LR()

		x8 := emu.X(8)

		// Check if X8 points to stack memory (return buffer for non-trivial types)
		// Stack range: StackBase to StackBase + StackSize
		if x8 >= StackBase && x8 < StackBase+StackSize {
			// Initialize an empty GNU libstdc++ std::string at x8
			// COW format: string object contains pointer to data (after _Rep header)
			//
			// We allocate from heap:
			// _Rep { _M_length (8), _M_capacity (8), _M_refcount (4), pad (4) } + data
			repSize := uint64(24)
			dataSize := uint64(16)
			total := repSize + dataSize

			repPtr := emu.Malloc(total)
			if repPtr == 0 {
				// Allocation failed, set x0 to mock object as fallback
				emu.SetX(0, emu.GetMockObject())
				return false // Don't stop, let RET execute
			}
			dataPtr := repPtr + repSize

			// Write _Rep header: length=0, capacity=15, refcount=0
			repHeader := make([]byte, 24)
			binary.LittleEndian.PutUint64(repHeader[0:8], 0)  // _M_length = 0
			binary.LittleEndian.PutUint64(repHeader[8:16], 15) // _M_capacity = 15
			// _M_refcount at offset 16, leave as 0
			emu.MemWrite(repPtr, repHeader)

			// Initialize data area with null terminator
			nullData := make([]byte, dataSize)
			emu.MemWrite(dataPtr, nullData)

			// Write pointer to data (not _Rep!) into the string object at x8
			ptrBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(ptrBytes, dataPtr)
			emu.MemWrite(x8, ptrBytes)

			// For return-by-value, x0 should return x8 (pointer to constructed object)
			emu.SetX(0, x8)
		} else {
			// Not a return-by-value call, return mock object
			emu.SetX(0, emu.GetMockObject())
		}

		return false // Don't stop, let RET execute
	}
}

// makeCallbackStubHook creates a hook for callback stubs (indirect calls through member pointers).
// These handle patterns like: x4 = [mock_obj + offset]; blr x4
// where the member pointer points to an executable stub.
func makeCallbackStubHook(e *Emulator) AddressHookFunc {
	return func(emu *Emulator) bool {
		x8 := emu.X(8)

		// Same logic as vtable stub - handle return-by-value convention
		if x8 >= StackBase && x8 < StackBase+StackSize {
			repSize := uint64(24)
			dataSize := uint64(16)
			total := repSize + dataSize

			repPtr := emu.Malloc(total)
			if repPtr == 0 {
				emu.SetX(0, 0)
				return false
			}
			dataPtr := repPtr + repSize

			repHeader := make([]byte, 24)
			binary.LittleEndian.PutUint64(repHeader[0:8], 0)
			binary.LittleEndian.PutUint64(repHeader[8:16], 15)
			emu.MemWrite(repPtr, repHeader)

			nullData := make([]byte, dataSize)
			emu.MemWrite(dataPtr, nullData)

			ptrBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(ptrBytes, dataPtr)
			emu.MemWrite(x8, ptrBytes)

			emu.SetX(0, x8)
		} else {
			emu.SetX(0, 0)
		}

		return false
	}
}

// makeMockObj2CallHook creates a hook for direct calls to mock_obj2 addresses.
// Code may load mock_obj2 as a function pointer and call it directly via blr.
func makeMockObj2CallHook(e *Emulator, mockObj uint64) AddressHookFunc {
	return func(emu *Emulator) bool {
		// Return mock object pointer
		emu.SetX(0, mockObj)
		return false
	}
}

// setupHooks initializes Unicorn hooks
func (e *Emulator) setupHooks() error {
	// Code hook for tracing and address hooks
	_, err := e.mu.HookAdd(uc.HOOK_CODE, func(mu uc.Unicorn, addr uint64, size uint32) {
		// Check for stop
		if e.stopped {
			e.mu.Stop()
			return
		}

		// Check address hooks first (protected by mutex)
		e.addrHooksMu.RLock()
		hook, ok := e.addrHooks[addr]
		e.addrHooksMu.RUnlock()

		if ok {
			if hook(e) {
				e.Stop()
				return
			}
		}

		// Call user code hooks
		for _, h := range e.codeHooks {
			h(e, addr, size)
		}
	}, 1, 0)

	return err
}

// Close releases resources
func (e *Emulator) Close() error {
	return e.mu.Close()
}

// LoadCode writes code at the code base
func (e *Emulator) LoadCode(code []byte) error {
	return e.mu.MemWrite(CodeBase, code)
}

// MapRegion maps additional memory
func (e *Emulator) MapRegion(addr, size uint64) error {
	return e.mu.MemMap(addr, size)
}

// MemRead reads bytes from memory
func (e *Emulator) MemRead(addr, size uint64) ([]byte, error) {
	data, err := e.mu.MemRead(addr, size)
	if err != nil {
		return nil, err
	}

	// HIPAA compliance check: scan for PHI in readable data
	// In clinical practice, we must monitor all data access to prevent PHI leaks.
	if e.isPrintableASCII(data) {
		str := string(data)
		if e.detector.ContainsPHI(str) {
			snippet := str
			if len(snippet) > 50 {
				snippet = snippet[:50] + "..."
			}
			e.auditor.LogPHIDetected(fmt.Sprintf("MemRead at 0x%x", addr), snippet)
		}
	}

	return data, nil
}

// isPrintableASCII checks if all bytes are printable ASCII characters.
func (e *Emulator) isPrintableASCII(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	for _, b := range data {
		if b < 32 || b > 126 {
			return false
		}
	}
	return true
}

// MemWrite writes bytes to memory
func (e *Emulator) MemWrite(addr uint64, data []byte) error {
	return e.mu.MemWrite(addr, data)
}

// MemReadU64 reads a uint64 from memory (little endian)
func (e *Emulator) MemReadU64(addr uint64) (uint64, error) {
	data, err := e.mu.MemRead(addr, 8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(data), nil
}

// MemWriteU64 writes a uint64 to memory (little endian)
func (e *Emulator) MemWriteU64(addr, val uint64) error {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, val)
	return e.mu.MemWrite(addr, data)
}

// MemReadU32 reads a uint32 from memory (little endian)
func (e *Emulator) MemReadU32(addr uint64) (uint32, error) {
	data, err := e.mu.MemRead(addr, 4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(data), nil
}

// MemWriteU32 writes a uint32 to memory (little endian)
func (e *Emulator) MemWriteU32(addr uint64, val uint32) error {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, val)
	return e.mu.MemWrite(addr, data)
}

// MemReadU16 reads a uint16 from memory (little endian)
func (e *Emulator) MemReadU16(addr uint64) (uint16, error) {
	data, err := e.mu.MemRead(addr, 2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(data), nil
}

// MemWriteU16 writes a uint16 to memory (little endian)
func (e *Emulator) MemWriteU16(addr uint64, val uint16) error {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, val)
	return e.mu.MemWrite(addr, data)
}

// MemReadU8 reads a single byte from memory
func (e *Emulator) MemReadU8(addr uint64) (uint8, error) {
	data, err := e.mu.MemRead(addr, 1)
	if err != nil {
		return 0, err
	}
	return data[0], nil
}

// MemWriteU8 writes a single byte to memory
func (e *Emulator) MemWriteU8(addr uint64, val uint8) error {
	return e.mu.MemWrite(addr, []byte{val})
}

// MemReadString reads a null-terminated string from memory
func (e *Emulator) MemReadString(addr uint64, maxLen int) (string, error) {
	if maxLen <= 0 {
		maxLen = 4096
	}
	data, err := e.mu.MemRead(addr, uint64(maxLen))
	if err != nil {
		return "", err
	}

	// Find null terminator
	for i, b := range data {
		if b == 0 {
			return string(data[:i]), nil
		}
	}
	return string(data), nil
}

// MemWriteString writes a null-terminated string to memory
func (e *Emulator) MemWriteString(addr uint64, s string) error {
	data := append([]byte(s), 0)
	return e.mu.MemWrite(addr, data)
}

// RegRead reads a register value
func (e *Emulator) RegRead(reg int) (uint64, error) {
	return e.mu.RegRead(reg)
}

// RegWrite writes a register value
func (e *Emulator) RegWrite(reg int, val uint64) error {
	return e.mu.RegWrite(reg, val)
}

// X reads general-purpose register X0-X30
func (e *Emulator) X(n int) uint64 {
	if n < 0 || n > 30 {
		return 0
	}
	val, _ := e.mu.RegRead(uc.ARM64_REG_X0 + n)
	return val
}

// SetX writes general-purpose register X0-X30
func (e *Emulator) SetX(n int, val uint64) error {
	if n < 0 || n > 30 {
		return fmt.Errorf("invalid register X%d", n)
	}
	return e.mu.RegWrite(uc.ARM64_REG_X0+n, val)
}

// PC returns the program counter
func (e *Emulator) PC() uint64 {
	pc, _ := e.mu.RegRead(uc.ARM64_REG_PC)
	return pc
}

// SetPC sets the program counter
func (e *Emulator) SetPC(val uint64) error {
	return e.mu.RegWrite(uc.ARM64_REG_PC, val)
}

// SP returns the stack pointer
func (e *Emulator) SP() uint64 {
	sp, _ := e.mu.RegRead(uc.ARM64_REG_SP)
	return sp
}

// SetSP sets the stack pointer
func (e *Emulator) SetSP(val uint64) error {
	return e.mu.RegWrite(uc.ARM64_REG_SP, val)
}

// LR returns the link register
func (e *Emulator) LR() uint64 {
	lr, _ := e.mu.RegRead(uc.ARM64_REG_LR)
	return lr
}

// SetLR sets the link register
func (e *Emulator) SetLR(val uint64) error {
	return e.mu.RegWrite(uc.ARM64_REG_LR, val)
}

// Malloc allocates memory from the heap (bump allocator).
// Panics if heap is exhausted - this indicates a fundamental emulation problem.
func (e *Emulator) Malloc(size uint64) uint64 {
	// Align to 16 bytes
	size = (size + 15) & ^uint64(15)

	addr := e.heapPtr
	e.heapPtr += size

	if e.heapPtr >= HeapBase+HeapSize {
		panic("heap exhausted")
	}

	return addr
}

// HookCode adds a code hook called for every instruction
func (e *Emulator) HookCode(fn CodeHookFunc) {
	e.codeHooks = append(e.codeHooks, fn)
}

// HookAddress adds a hook for a specific address
func (e *Emulator) HookAddress(addr uint64, fn AddressHookFunc) {
	e.addrHooksMu.Lock()
	defer e.addrHooksMu.Unlock()
	e.addrHooks[addr] = fn
}

// RemoveAddressHook removes an address hook
func (e *Emulator) RemoveAddressHook(addr uint64) {
	e.addrHooksMu.Lock()
	defer e.addrHooksMu.Unlock()
	delete(e.addrHooks, addr)
}

// EnableTrace enables instruction tracing
func (e *Emulator) EnableTrace() {
	e.traceEnabled = true
}

// DisableTrace disables instruction tracing
func (e *Emulator) DisableTrace() {
	e.traceEnabled = false
}

// GetTraceEvents returns collected trace events
func (e *Emulator) GetTraceEvents() []TraceEvent {
	e.traceMu.Lock()
	defer e.traceMu.Unlock()
	return append([]TraceEvent{}, e.traceEvents...)
}

// AddTraceEvent adds a trace event
func (e *Emulator) AddTraceEvent(event TraceEvent) {
	e.traceMu.Lock()
	defer e.traceMu.Unlock()
	e.traceEvents = append(e.traceEvents, event)
}

// ClearTrace clears trace events
func (e *Emulator) ClearTrace() {
	e.traceMu.Lock()
	defer e.traceMu.Unlock()
	e.traceEvents = nil
}

// Run starts emulation from addr
func (e *Emulator) Run(start, end uint64) error {
	e.stopped = false
	return e.mu.Start(start, end)
}

// RunFrom starts emulation from current PC
func (e *Emulator) RunFrom(start uint64) error {
	e.stopped = false
	// Use 0 as end address to run until stop
	return e.mu.Start(start, 0)
}

// Stop stops emulation
func (e *Emulator) Stop() {
	e.stopped = true
	e.mu.Stop()
}

// ARM64 register constants (re-exported for convenience)
const (
	RegX0  = uc.ARM64_REG_X0
	RegX1  = uc.ARM64_REG_X1
	RegX2  = uc.ARM64_REG_X2
	RegX3  = uc.ARM64_REG_X3
	RegX4  = uc.ARM64_REG_X4
	RegX5  = uc.ARM64_REG_X5
	RegX6  = uc.ARM64_REG_X6
	RegX7  = uc.ARM64_REG_X7
	RegX8  = uc.ARM64_REG_X8
	RegX29 = uc.ARM64_REG_X29 // Frame pointer
	RegX30 = uc.ARM64_REG_X30 // Link register (same as LR)
	RegSP  = uc.ARM64_REG_SP
	RegPC  = uc.ARM64_REG_PC
	RegLR  = uc.ARM64_REG_LR
)
