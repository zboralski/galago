// Package stubs provides C++ ABI stub implementations.
// This file implements stubs for libc++ (NDK) std::string with SSO (Short String Optimization)
// and basic exception handling.
package stubs

import (
	"github.com/zboralski/galago/internal/emulator"
)

// CxxAbiStubs provides stub implementations for C++ ABI functions.
// Supports libc++ (NDK) std::string with Short String Optimization (SSO).
type CxxAbiStubs struct {
	emu *emulator.Emulator

	// Trace callback for logging stub calls
	OnCall func(name string, detail string)

	// Callback for captured strings (from constructors, assign)
	OnStringCapture func(addr uint64, value string)

	// String pool for tracking constructed strings
	strings map[uint64]string
}

// NewCxxAbiStubs creates C++ ABI stubs for an emulator.
func NewCxxAbiStubs(emu *emulator.Emulator) *CxxAbiStubs {
	return &CxxAbiStubs{
		emu:     emu,
		strings: make(map[uint64]string),
	}
}

// Install registers all C++ ABI stub hooks based on imported symbols.
func (s *CxxAbiStubs) Install(imports map[string]uint64) {
	for name, addr := range imports {
		if addr == 0 {
			continue
		}

		// Match std::string constructors
		// _ZNSt6__ndk112basic_stringIcNS_11char_traitsIcEENS_9allocatorIcEEEC2IDnEEPKc
		if containsAll(name, "basic_string", "char_traits", "C2") {
			s.installStub(name, addr, s.stubStringCtor)
			continue
		}

		// Match std::string::assign(const char*)
		// _ZNSt6__ndk112basic_stringIcNS_11char_traitsIcEENS_9allocatorIcEEE6assignEPKc
		if containsAll(name, "basic_string", "6assignEPKc") {
			s.installStub(name, addr, s.stubStringAssign)
			continue
		}

		// Match std::string::c_str() or data()
		if containsAll(name, "basic_string", "c_str") || containsAll(name, "basic_string", "4data") {
			s.installStub(name, addr, s.stubStringCStr)
			continue
		}

		// Match std::string::size() or length()
		if containsAll(name, "basic_string", "4size") || containsAll(name, "basic_string", "6length") {
			s.installStub(name, addr, s.stubStringSize)
			continue
		}

		// Match __cxa_throw (exception throwing)
		if name == "__cxa_throw" {
			s.installStub(name, addr, s.stubCxaThrow)
			continue
		}

		// Match __cxa_begin_catch
		if name == "__cxa_begin_catch" {
			s.installStub(name, addr, s.stubCxaBeginCatch)
			continue
		}

		// Match __cxa_end_catch
		if name == "__cxa_end_catch" {
			s.installStub(name, addr, s.stubCxaEndCatch)
			continue
		}

		// Match __cxa_allocate_exception
		if name == "__cxa_allocate_exception" {
			s.installStub(name, addr, s.stubCxaAllocateException)
			continue
		}

		// Match __cxa_free_exception
		if name == "__cxa_free_exception" {
			s.installStub(name, addr, s.stubCxaFreeException)
			continue
		}

		// Match __cxa_guard_acquire/release (static initialization guards)
		if name == "__cxa_guard_acquire" {
			s.installStub(name, addr, s.stubCxaGuardAcquire)
			continue
		}
		if name == "__cxa_guard_release" {
			s.installStub(name, addr, s.stubCxaGuardRelease)
			continue
		}

		// Match __cxa_atexit
		if name == "__cxa_atexit" {
			s.installStub(name, addr, s.stubCxaAtexit)
			continue
		}
	}
}

// InstallAt installs a stub at a specific address.
func (s *CxxAbiStubs) InstallAt(name string, addr uint64, stub func()) {
	s.installStub(name, addr, stub)
}

func (s *CxxAbiStubs) installStub(name string, addr uint64, stub func()) {
	s.emu.HookAddress(addr, func(e *emulator.Emulator) bool {
		stub()
		return false
	})
}

func (s *CxxAbiStubs) log(name, detail string) {
	if s.OnCall != nil {
		s.OnCall(name, detail)
	}
}

func (s *CxxAbiStubs) returnFromStub() {
	lr := s.emu.LR()
	s.emu.SetPC(lr)
}

// SSO String Layout (libc++ / NDK):
//
// The std::string object is 24 bytes on 64-bit systems.
//
// Short String (length < 23):
//   byte 0: (length << 1) | 0  (bit 0 = 0 indicates short string)
//   bytes 1-22: inline character data
//   byte 23: unused
//
// Long String (length >= 23):
//   bytes 0-7: capacity | 1   (bit 0 = 1 indicates long string)
//   bytes 8-15: length
//   bytes 16-23: pointer to heap-allocated data

const (
	ssoMaxLen  = 22 // Max inline string length for SSO
	ssoObjSize = 24 // Size of std::string object
)

// ReadSSOString reads a libc++ SSO std::string from memory.
func (s *CxxAbiStubs) ReadSSOString(addr uint64) (string, bool) {
	if addr == 0 || addr < 0x1000 {
		return "", false
	}

	data, err := s.emu.MemRead(addr, ssoObjSize)
	if err != nil || len(data) < ssoObjSize {
		return "", false
	}

	// Check if long or short string (bit 0 of first byte)
	isLong := (data[0] & 1) == 1

	if isLong {
		// Long string: length at offset 8, data pointer at offset 16
		length := uint64(data[8]) | uint64(data[9])<<8 | uint64(data[10])<<16 | uint64(data[11])<<24 |
			uint64(data[12])<<32 | uint64(data[13])<<40 | uint64(data[14])<<48 | uint64(data[15])<<56

		dataPtr := uint64(data[16]) | uint64(data[17])<<8 | uint64(data[18])<<16 | uint64(data[19])<<24 |
			uint64(data[20])<<32 | uint64(data[21])<<40 | uint64(data[22])<<48 | uint64(data[23])<<56

		if length > 4096 || dataPtr < 0x1000 {
			return "", false
		}

		strData, err := s.emu.MemRead(dataPtr, length)
		if err != nil {
			return "", false
		}
		return string(strData), true
	}

	// Short string: length is in bits 1-7 of first byte
	length := int(data[0] >> 1)
	if length > ssoMaxLen || length < 0 {
		return "", false
	}

	return string(data[1 : 1+length]), true
}

// WriteSSOString writes a string to memory in libc++ SSO format.
func (s *CxxAbiStubs) WriteSSOString(addr uint64, str string) error {
	strBytes := []byte(str)
	strLen := len(strBytes)

	if strLen <= ssoMaxLen {
		// Short string optimization
		ssoData := make([]byte, ssoObjSize)
		ssoData[0] = byte(strLen << 1) // Length in bits 1-7, bit 0 = 0 (short)
		copy(ssoData[1:], strBytes)
		if strLen < ssoMaxLen {
			ssoData[1+strLen] = 0 // Null terminator
		}
		return s.emu.MemWrite(addr, ssoData)
	}

	// Long string - allocate heap buffer
	bufSize := uint64(strLen + 1) // +1 for null terminator
	bufSize = (bufSize + 15) & ^uint64(15)
	dataPtr := s.emu.Malloc(bufSize)

	// Write string data to heap
	dataWithNull := append(strBytes, 0)
	if err := s.emu.MemWrite(dataPtr, dataWithNull); err != nil {
		return err
	}

	// Write std::string object
	ssoData := make([]byte, ssoObjSize)

	// Capacity with long bit set (bit 0 = 1)
	capacity := bufSize | 1
	ssoData[0] = byte(capacity)
	ssoData[1] = byte(capacity >> 8)
	ssoData[2] = byte(capacity >> 16)
	ssoData[3] = byte(capacity >> 24)
	ssoData[4] = byte(capacity >> 32)
	ssoData[5] = byte(capacity >> 40)
	ssoData[6] = byte(capacity >> 48)
	ssoData[7] = byte(capacity >> 56)

	// Length
	ssoData[8] = byte(strLen)
	ssoData[9] = byte(strLen >> 8)
	ssoData[10] = byte(strLen >> 16)
	ssoData[11] = byte(strLen >> 24)
	ssoData[12] = byte(strLen >> 32)
	ssoData[13] = byte(strLen >> 40)
	ssoData[14] = byte(strLen >> 48)
	ssoData[15] = byte(strLen >> 56)

	// Data pointer
	ssoData[16] = byte(dataPtr)
	ssoData[17] = byte(dataPtr >> 8)
	ssoData[18] = byte(dataPtr >> 16)
	ssoData[19] = byte(dataPtr >> 24)
	ssoData[20] = byte(dataPtr >> 32)
	ssoData[21] = byte(dataPtr >> 40)
	ssoData[22] = byte(dataPtr >> 48)
	ssoData[23] = byte(dataPtr >> 56)

	return s.emu.MemWrite(addr, ssoData)
}

// GetSSODataPtr returns a pointer to the string data (for c_str()/data()).
func (s *CxxAbiStubs) GetSSODataPtr(addr uint64) uint64 {
	if addr == 0 || addr < 0x1000 {
		return 0
	}

	data, err := s.emu.MemRead(addr, ssoObjSize)
	if err != nil || len(data) < ssoObjSize {
		return 0
	}

	isLong := (data[0] & 1) == 1
	if isLong {
		// Long string: data pointer at offset 16
		return uint64(data[16]) | uint64(data[17])<<8 | uint64(data[18])<<16 | uint64(data[19])<<24 |
			uint64(data[20])<<32 | uint64(data[21])<<40 | uint64(data[22])<<48 | uint64(data[23])<<56
	}

	// Short string: data starts at offset 1 within the object
	return addr + 1
}

// GetSSOLength returns the length of an SSO string.
func (s *CxxAbiStubs) GetSSOLength(addr uint64) uint64 {
	if addr == 0 || addr < 0x1000 {
		return 0
	}

	data, err := s.emu.MemRead(addr, ssoObjSize)
	if err != nil || len(data) < ssoObjSize {
		return 0
	}

	isLong := (data[0] & 1) == 1
	if isLong {
		// Long string: length at offset 8
		return uint64(data[8]) | uint64(data[9])<<8 | uint64(data[10])<<16 | uint64(data[11])<<24 |
			uint64(data[12])<<32 | uint64(data[13])<<40 | uint64(data[14])<<48 | uint64(data[15])<<56
	}

	// Short string: length in bits 1-7 of first byte
	return uint64(data[0] >> 1)
}

// stubStringCtor implements std::string constructor from const char*.
// Prototype: basic_string(this, const char* s)
// X0 = this pointer, X1 = const char* source
func (s *CxxAbiStubs) stubStringCtor() {
	thisPtr := s.emu.X(0)
	srcPtr := s.emu.X(1)

	// Read source string
	str, _ := s.emu.MemReadString(srcPtr, 4096)

	// Write SSO string to this
	s.WriteSSOString(thisPtr, str)

	// Track the string
	s.strings[thisPtr] = str

	// Callback for string capture
	if s.OnStringCapture != nil && len(str) > 0 {
		s.OnStringCapture(thisPtr, str)
	}

	s.log("string::ctor", formatStringOp(thisPtr, str))
	s.emu.SetX(0, thisPtr) // Return this
	s.returnFromStub()
}

// stubStringAssign implements std::string::assign(const char*).
// Prototype: assign(this, const char* s)
// X0 = this pointer, X1 = const char* source
func (s *CxxAbiStubs) stubStringAssign() {
	thisPtr := s.emu.X(0)
	srcPtr := s.emu.X(1)

	// Read source string
	str, _ := s.emu.MemReadString(srcPtr, 4096)

	// Write SSO string to this
	s.WriteSSOString(thisPtr, str)

	// Track the string
	s.strings[thisPtr] = str

	// Callback for string capture
	if s.OnStringCapture != nil && len(str) > 0 {
		s.OnStringCapture(thisPtr, str)
	}

	s.log("string::assign", formatStringOp(thisPtr, str))
	s.emu.SetX(0, thisPtr) // Return this
	s.returnFromStub()
}

// stubStringCStr implements std::string::c_str() and data().
// Prototype: const char* c_str(this)
// X0 = this pointer
func (s *CxxAbiStubs) stubStringCStr() {
	thisPtr := s.emu.X(0)
	dataPtr := s.GetSSODataPtr(thisPtr)

	s.log("string::c_str", formatPtr("this", thisPtr, "data", dataPtr))
	s.emu.SetX(0, dataPtr)
	s.returnFromStub()
}

// stubStringSize implements std::string::size() and length().
// Prototype: size_t size(this)
// X0 = this pointer
func (s *CxxAbiStubs) stubStringSize() {
	thisPtr := s.emu.X(0)
	length := s.GetSSOLength(thisPtr)

	s.log("string::size", formatPtr("this", thisPtr, "len", length))
	s.emu.SetX(0, length)
	s.returnFromStub()
}

// Exception handling stubs

// stubCxaThrow implements __cxa_throw.
// For emulation, we just stop execution since we don't support real exception unwinding.
func (s *CxxAbiStubs) stubCxaThrow() {
	excPtr := s.emu.X(0)
	// typeInfo := s.emu.X(1)
	// destructor := s.emu.X(2)

	s.log("__cxa_throw", formatPtr("exception", excPtr, "", 0))

	// Stop emulation on throw - we don't support real exception handling
	s.emu.Stop()
}

// stubCxaBeginCatch implements __cxa_begin_catch.
func (s *CxxAbiStubs) stubCxaBeginCatch() {
	excPtr := s.emu.X(0)

	s.log("__cxa_begin_catch", formatPtr("exception", excPtr, "", 0))
	s.emu.SetX(0, excPtr) // Return exception pointer
	s.returnFromStub()
}

// stubCxaEndCatch implements __cxa_end_catch.
func (s *CxxAbiStubs) stubCxaEndCatch() {
	s.log("__cxa_end_catch", "")
	s.returnFromStub()
}

// stubCxaAllocateException implements __cxa_allocate_exception.
func (s *CxxAbiStubs) stubCxaAllocateException() {
	size := s.emu.X(0)
	if size == 0 {
		size = 64
	}

	ptr := s.emu.Malloc(size + 128) // Extra for exception header

	s.log("__cxa_allocate_exception", formatPtr("size", size, "ptr", ptr))
	s.emu.SetX(0, ptr)
	s.returnFromStub()
}

// stubCxaFreeException implements __cxa_free_exception.
func (s *CxxAbiStubs) stubCxaFreeException() {
	s.log("__cxa_free_exception", "")
	s.returnFromStub()
}

// Static initialization guard stubs

var guardState = make(map[uint64]bool)

// stubCxaGuardAcquire implements __cxa_guard_acquire.
// Returns 1 if initialization should proceed, 0 if already initialized.
func (s *CxxAbiStubs) stubCxaGuardAcquire() {
	guardPtr := s.emu.X(0)

	// Check if already initialized
	if guardState[guardPtr] {
		s.log("__cxa_guard_acquire", formatPtr("guard", guardPtr, "result", 0))
		s.emu.SetX(0, 0) // Already initialized
	} else {
		s.log("__cxa_guard_acquire", formatPtr("guard", guardPtr, "result", 1))
		s.emu.SetX(0, 1) // Need to initialize
	}
	s.returnFromStub()
}

// stubCxaGuardRelease implements __cxa_guard_release.
// Marks the guard as initialized.
func (s *CxxAbiStubs) stubCxaGuardRelease() {
	guardPtr := s.emu.X(0)
	guardState[guardPtr] = true

	s.log("__cxa_guard_release", formatPtr("guard", guardPtr, "", 0))
	s.returnFromStub()
}

// stubCxaAtexit implements __cxa_atexit.
// Registers a destructor to be called at exit - we just ignore it.
func (s *CxxAbiStubs) stubCxaAtexit() {
	// func := s.emu.X(0)
	// arg := s.emu.X(1)
	// dso := s.emu.X(2)

	s.log("__cxa_atexit", "registered")
	s.emu.SetX(0, 0) // Return 0 for success
	s.returnFromStub()
}

// GetTrackedString returns a tracked string by address.
func (s *CxxAbiStubs) GetTrackedString(addr uint64) (string, bool) {
	str, ok := s.strings[addr]
	return str, ok
}

// GetAllTrackedStrings returns all tracked strings.
func (s *CxxAbiStubs) GetAllTrackedStrings() map[uint64]string {
	result := make(map[uint64]string)
	for k, v := range s.strings {
		result[k] = v
	}
	return result
}

// Helper functions

func containsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if !containsString(s, sub) {
			return false
		}
	}
	return true
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func formatStringOp(addr uint64, str string) string {
	if len(str) > 30 {
		str = str[:30] + "..."
	}
	return "addr=" + formatHex(addr) + " \"" + str + "\""
}
