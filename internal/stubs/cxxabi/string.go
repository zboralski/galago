// Package cxxabi provides stub implementations for C++ ABI functions.
// This file implements stubs for libc++ (NDK) std::string with SSO (Short String Optimization).
package cxxabi

import (
	"strings"
	"sync"

	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

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
	SSOMaxLen  = 22 // Max inline string length for SSO
	SSOObjSize = 24 // Size of std::string object
)

var (
	// trackedStrings stores constructed strings by address
	trackedStrings   = make(map[uint64]string)
	trackedStringsMu sync.RWMutex

	// OnStringCapture is called when a string is constructed/assigned
	OnStringCapture func(addr uint64, value string)
)

func init() {
	// Register __cxa_demangle as a simple stub
	stubs.Register(stubs.StubDef{
		Name:     "__cxa_demangle",
		Category: "cxxabi",
		Hook:     stubCxaDemangle,
	})

	// Register std::string as a detector - activates on mangled basic_string symbols
	stubs.RegisterDetector(stubs.Detector{
		Name: "cxxabi-string",
		Patterns: []string{
			"basic_string",
			"_ZNSt",
			"__ndk1",
		},
		Activate:    activateStringHooks,
		Description: "C++ std::string SSO implementation",
	})
}

// activateStringHooks installs std::string hooks when C++ string symbols are detected.
func activateStringHooks(emu *emulator.Emulator, imports, symbols map[string]uint64) int {
	// Install from both imports and symbols
	installed := InstallStringHooks(emu, symbols)
	if installed > 0 {
		stubs.DefaultRegistry.Log("cxxabi", "activate", "std::string hooks installed")
	}
	return installed
}

// InstallStringHooks scans imports and installs std::string hooks based on mangled names.
// This must be called after loading the binary to match the actual mangled names.
func InstallStringHooks(emu *emulator.Emulator, imports map[string]uint64) int {
	installed := 0

	for name, addr := range imports {
		if addr == 0 {
			continue
		}

		// Match std::string constructors
		// _ZNSt6__ndk112basic_stringIcNS_11char_traitsIcEENS_9allocatorIcEEEC2IDnEEPKc
		if containsAll(name, "basic_string", "char_traits", "C2") && strings.Contains(name, "PKc") {
			emu.HookAddress(addr, stubStringCtor)
			installed++
			continue
		}

		// Match std::string::assign(const char*)
		// _ZNSt6__ndk112basic_stringIcNS_11char_traitsIcEENS_9allocatorIcEEE6assignEPKc
		if containsAll(name, "basic_string", "6assignEPKc") {
			emu.HookAddress(addr, stubStringAssign)
			installed++
			continue
		}

		// Match std::string::c_str() or data()
		if containsAll(name, "basic_string", "c_str") || containsAll(name, "basic_string", "4data") {
			emu.HookAddress(addr, stubStringCStr)
			installed++
			continue
		}

		// Match std::string::size() or length()
		if containsAll(name, "basic_string", "4size") || containsAll(name, "basic_string", "6length") {
			emu.HookAddress(addr, stubStringSize)
			installed++
			continue
		}

		// Match std::string::empty()
		if containsAll(name, "basic_string", "5empty") {
			emu.HookAddress(addr, stubStringEmpty)
			installed++
			continue
		}

		// Match std::string::clear()
		if containsAll(name, "basic_string", "5clear") {
			emu.HookAddress(addr, stubStringClear)
			installed++
			continue
		}

		// Match std::string::operator[]
		if containsAll(name, "basic_string", "ixEm") {
			emu.HookAddress(addr, stubStringIndex)
			installed++
			continue
		}
	}

	return installed
}

// ReadSSOString reads a libc++ SSO std::string from memory.
func ReadSSOString(emu *emulator.Emulator, addr uint64) (string, bool) {
	if addr == 0 || addr < 0x1000 {
		return "", false
	}

	data, err := emu.MemRead(addr, SSOObjSize)
	if err != nil || len(data) < SSOObjSize {
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

		strData, err := emu.MemRead(dataPtr, length)
		if err != nil {
			return "", false
		}
		return string(strData), true
	}

	// Short string: length is in bits 1-7 of first byte
	length := int(data[0] >> 1)
	if length > SSOMaxLen || length < 0 {
		return "", false
	}

	return string(data[1 : 1+length]), true
}

// WriteSSOString writes a string to memory in libc++ SSO format.
func WriteSSOString(emu *emulator.Emulator, addr uint64, str string) error {
	strBytes := []byte(str)
	strLen := len(strBytes)

	if strLen <= SSOMaxLen {
		// Short string optimization
		ssoData := make([]byte, SSOObjSize)
		ssoData[0] = byte(strLen << 1) // Length in bits 1-7, bit 0 = 0 (short)
		copy(ssoData[1:], strBytes)
		if strLen < SSOMaxLen {
			ssoData[1+strLen] = 0 // Null terminator
		}
		return emu.MemWrite(addr, ssoData)
	}

	// Long string - allocate heap buffer
	bufSize := uint64(strLen + 1) // +1 for null terminator
	bufSize = (bufSize + 15) & ^uint64(15)
	dataPtr := emu.Malloc(bufSize)

	// Write string data to heap
	dataWithNull := append(strBytes, 0)
	if err := emu.MemWrite(dataPtr, dataWithNull); err != nil {
		return err
	}

	// Write std::string object
	ssoData := make([]byte, SSOObjSize)

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

	return emu.MemWrite(addr, ssoData)
}

// GetSSODataPtr returns a pointer to the string data (for c_str()/data()).
func GetSSODataPtr(emu *emulator.Emulator, addr uint64) uint64 {
	if addr == 0 || addr < 0x1000 {
		return 0
	}

	data, err := emu.MemRead(addr, SSOObjSize)
	if err != nil || len(data) < SSOObjSize {
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
func GetSSOLength(emu *emulator.Emulator, addr uint64) uint64 {
	if addr == 0 || addr < 0x1000 {
		return 0
	}

	data, err := emu.MemRead(addr, SSOObjSize)
	if err != nil || len(data) < SSOObjSize {
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

// GetTrackedStrings returns all tracked strings.
func GetTrackedStrings() map[uint64]string {
	trackedStringsMu.RLock()
	defer trackedStringsMu.RUnlock()
	result := make(map[uint64]string, len(trackedStrings))
	for k, v := range trackedStrings {
		result[k] = v
	}
	return result
}

// ClearTrackedStrings clears the tracked strings map.
func ClearTrackedStrings() {
	trackedStringsMu.Lock()
	trackedStrings = make(map[uint64]string)
	trackedStringsMu.Unlock()
}

// trackString adds a string to the tracking map and calls the callback.
func trackString(addr uint64, value string) {
	trackedStringsMu.Lock()
	trackedStrings[addr] = value
	cb := OnStringCapture
	trackedStringsMu.Unlock()

	if cb != nil && len(value) > 0 {
		cb(addr, value)
	}
}

// stubStringCtor implements std::string constructor from const char*.
func stubStringCtor(emu *emulator.Emulator) bool {
	thisPtr := emu.X(0)
	srcPtr := emu.X(1)

	str, _ := emu.MemReadString(srcPtr, 4096)
	WriteSSOString(emu, thisPtr, str)
	trackString(thisPtr, str)

	truncated := str
	if len(truncated) > 30 {
		truncated = truncated[:30] + "..."
	}
	stubs.DefaultRegistry.Log("cxxabi", "string::ctor", "\""+truncated+"\"")

	emu.SetX(0, thisPtr)
	stubs.ReturnFromStub(emu)
	return false
}

// stubStringAssign implements std::string::assign(const char*).
func stubStringAssign(emu *emulator.Emulator) bool {
	thisPtr := emu.X(0)
	srcPtr := emu.X(1)

	str, _ := emu.MemReadString(srcPtr, 4096)
	WriteSSOString(emu, thisPtr, str)
	trackString(thisPtr, str)

	truncated := str
	if len(truncated) > 30 {
		truncated = truncated[:30] + "..."
	}
	stubs.DefaultRegistry.Log("cxxabi", "string::assign", "\""+truncated+"\"")

	emu.SetX(0, thisPtr)
	stubs.ReturnFromStub(emu)
	return false
}

// stubStringCStr implements std::string::c_str() and data().
func stubStringCStr(emu *emulator.Emulator) bool {
	thisPtr := emu.X(0)
	dataPtr := GetSSODataPtr(emu, thisPtr)

	emu.SetX(0, dataPtr)
	stubs.ReturnFromStub(emu)
	return false
}

// stubStringSize implements std::string::size() and length().
func stubStringSize(emu *emulator.Emulator) bool {
	thisPtr := emu.X(0)
	length := GetSSOLength(emu, thisPtr)

	emu.SetX(0, length)
	stubs.ReturnFromStub(emu)
	return false
}

// stubStringEmpty implements std::string::empty().
func stubStringEmpty(emu *emulator.Emulator) bool {
	thisPtr := emu.X(0)
	length := GetSSOLength(emu, thisPtr)

	if length == 0 {
		emu.SetX(0, 1) // true
	} else {
		emu.SetX(0, 0) // false
	}
	stubs.ReturnFromStub(emu)
	return false
}

// stubStringClear implements std::string::clear().
func stubStringClear(emu *emulator.Emulator) bool {
	thisPtr := emu.X(0)
	// Write empty string
	WriteSSOString(emu, thisPtr, "")
	stubs.ReturnFromStub(emu)
	return false
}

// stubStringIndex implements std::string::operator[].
func stubStringIndex(emu *emulator.Emulator) bool {
	thisPtr := emu.X(0)
	index := emu.X(1)

	dataPtr := GetSSODataPtr(emu, thisPtr)
	charPtr := dataPtr + index

	emu.SetX(0, charPtr)
	stubs.ReturnFromStub(emu)
	return false
}

// stubCxaDemangle implements __cxa_demangle.
func stubCxaDemangle(emu *emulator.Emulator) bool {
	mangledPtr := emu.X(0)
	// output := emu.X(1)
	// length := emu.X(2)
	statusPtr := emu.X(3)

	// Read mangled name
	mangled, _ := emu.MemReadString(mangledPtr, 512)
	stubs.DefaultRegistry.Log("cxxabi", "__cxa_demangle", mangled)

	// Just return the mangled name (no actual demangling)
	result := emu.Malloc(uint64(len(mangled) + 1))
	emu.MemWriteString(result, mangled)

	if statusPtr != 0 {
		emu.MemWriteU32(statusPtr, 0) // Success
	}

	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

// Helper functions

func containsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
