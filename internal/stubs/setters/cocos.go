// Package setters provides hooks for key setter functions.
// This package captures encryption keys, secrets, and other sensitive values
// passed to setter functions in game engines and frameworks.
package setters

import (
	"fmt"
	"strings"
	"sync"

	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

// CapturedKey represents an extracted encryption key or secret.
type CapturedKey struct {
	Value    string // The key value
	Source   string // Function name that set the key
	Address  uint64 // Address where the key was captured
	KeyType  string // Type of key: "xxtea", "aes", "des", "custom"
	RiskLevel string // "critical", "high", "medium", "low"
}

var (
	capturedKeys   []CapturedKey
	capturedKeysMu sync.Mutex

	// OnKeyCapture is called when a key is captured
	OnKeyCapture func(key CapturedKey)
)

func init() {
	// Register Cocos2d-x detector
	stubs.RegisterDetector(stubs.Detector{
		Name: "cocos2dx",
		Patterns: []string{
			"cocos2d",
			"setXXTeaKey",
			"ZipUtils",
			"ccDecrypt",
			"jsb_set",
		},
		Activate:    activateCocos2dx,
		Description: "Cocos2d-x XXTEA key extraction",
	})

	// Register Unity IL2CPP detector
	stubs.RegisterDetector(stubs.Detector{
		Name: "unity-il2cpp",
		Patterns: []string{
			"il2cpp",
			"Il2Cpp",
			"mono_",
		},
		Activate:    activateUnityIL2CPP,
		Description: "Unity IL2CPP key extraction",
	})
}

// GetCapturedKeys returns all captured keys.
func GetCapturedKeys() []CapturedKey {
	capturedKeysMu.Lock()
	defer capturedKeysMu.Unlock()
	result := make([]CapturedKey, len(capturedKeys))
	copy(result, capturedKeys)
	return result
}

// ClearCapturedKeys clears the captured keys list.
func ClearCapturedKeys() {
	capturedKeysMu.Lock()
	capturedKeys = nil
	capturedKeysMu.Unlock()
}

// captureKey adds a key to the captured list and calls the callback.
func captureKey(key CapturedKey) {
	capturedKeysMu.Lock()
	capturedKeys = append(capturedKeys, key)
	cb := OnKeyCapture
	capturedKeysMu.Unlock()

	stubs.DefaultRegistry.Log("setter", key.Source, key.Value)

	if cb != nil {
		cb(key)
	}
}

// CaptureKeyDirect is an exported function to capture a key directly from vtable hooks.
// This is used by the runTrace code in main.go to capture keys from vtable dispatch.
func CaptureKeyDirect(value, source string, address uint64) {
	// Detect key type from source
	keyType := "unknown"
	sourceLower := strings.ToLower(source)
	if strings.Contains(sourceLower, "xxtea") || strings.Contains(sourceLower, "xtea") {
		keyType = "xxtea"
	} else if strings.Contains(sourceLower, "signature") {
		keyType = "signature"
	} else if strings.Contains(sourceLower, "crypto") || strings.Contains(sourceLower, "aes") {
		keyType = "crypto"
	}

	captureKey(CapturedKey{
		Value:     value,
		Source:    source,
		Address:   address,
		RiskLevel: "critical",
		KeyType:   keyType,
	})
}

// isPrintableASCII checks if all characters in the string are printable ASCII.
func isPrintableASCII(s string) bool {
	for _, c := range s {
		if c < 32 || c > 126 {
			return false
		}
	}
	return len(s) > 0
}

// activateCocos2dx installs Cocos2d-x key setter hooks.
func activateCocos2dx(emu *emulator.Emulator, imports, symbols map[string]uint64) int {
	installed := 0

	// Search in symbols (includes both internal functions and PLT imports)
	for name, addr := range symbols {
		if addr == 0 {
			continue
		}

		// setXXTeaKey patterns (various capitalizations) - case insensitive search
		lower := strings.ToLower(name)
		if strings.Contains(lower, "setxxteakey") ||
			strings.Contains(lower, "set_xxtea_key") {
			if stubs.Debug {
				stubs.DefaultRegistry.Log("setter", "xxtea-hook",
					fmt.Sprintf("%s @ 0x%x", name, addr))
			}
			emu.HookAddress(addr, makeXXTeaKeyHook(name))
			installed++
			continue
		}

		// jsb::setXXTeaKey
		if strings.Contains(name, "jsb") && strings.Contains(name, "XTea") {
			emu.HookAddress(addr, makeXXTeaKeyHook(name))
			installed++
			continue
		}

		// ZipUtils encryption key
		if strings.Contains(name, "ZipUtils") && strings.Contains(name, "Key") {
			emu.HookAddress(addr, makeXXTeaKeyHook(name))
			installed++
			continue
		}

		// cc::Application::setXXTeaKey
		if strings.Contains(name, "Application") && strings.Contains(name, "XTea") {
			emu.HookAddress(addr, makeXXTeaKeyHook(name))
			installed++
			continue
		}

		// Generic crypto key setters (setCryptoKey matches setCryptoKeyAndSign)
		if strings.Contains(name, "setCryptoKey") ||
			strings.Contains(name, "CryptoKeyAndSign") ||
			strings.Contains(name, "setEncryptKey") ||
			strings.Contains(name, "setDecryptKey") {
			if stubs.Debug {
				stubs.DefaultRegistry.Log("setter", "crypto-hook",
					fmt.Sprintf("%s @ 0x%x", name, addr))
			}
			emu.HookAddress(addr, makeStdStringSetterHook(name))
			installed++
			continue
		}

		// AES key setters
		if strings.Contains(name, "setAESKey") ||
			strings.Contains(name, "AES_set_key") ||
			strings.Contains(name, "aes_key") {
			emu.HookAddress(addr, makeGenericKeyHook(name, "aes"))
			installed++
			continue
		}
	}

	if installed > 0 {
		stubs.DefaultRegistry.Log("setter", "cocos2dx", "key setters installed")
	}
	return installed
}

// activateUnityIL2CPP installs Unity IL2CPP hooks for key extraction.
func activateUnityIL2CPP(emu *emulator.Emulator, imports, symbols map[string]uint64) int {
	installed := 0

	// Search in symbols (includes both internal functions and PLT imports)
	for name, addr := range symbols {
		if addr == 0 {
			continue
		}

		// Skip Cocos2d-x patterns - those are handled by cocos2dx detector
		if strings.Contains(name, "setCryptoKey") || strings.Contains(name, "setXXTeaKey") ||
			strings.Contains(name, "cocos2d") {
			continue
		}

		// Common Unity encryption patterns
		if strings.Contains(name, "Encrypt") || strings.Contains(name, "Decrypt") {
			if strings.Contains(name, "Key") || strings.Contains(name, "key") {
				emu.HookAddress(addr, makeGenericKeyHook(name, "custom"))
				installed++
				continue
			}
		}

		// Unity Crypto utilities
		if strings.Contains(name, "Crypto") && strings.Contains(name, "Key") {
			emu.HookAddress(addr, makeGenericKeyHook(name, "custom"))
			installed++
			continue
		}
	}

	if installed > 0 {
		stubs.DefaultRegistry.Log("setter", "unity-il2cpp", "key setters installed")
	}
	return installed
}

// readStdString reads a std::string from memory.
// Supports both libc++ (SSO) and libstdc++ (COW) layouts.
// Returns the string value and true on success.
func readStdString(emu *emulator.Emulator, addr uint64) (string, bool) {
	if addr == 0 {
		return "", false
	}

	// Read first 24 bytes (std::string object)
	data, err := emu.MemRead(addr, 24)
	if err != nil || len(data) < 24 {
		if stubs.Debug {
			stubs.DefaultRegistry.Log("setter-debug", "readStdString",
				fmt.Sprintf("read failed addr=%x err=%v len=%d", addr, err, len(data)))
		}
		return "", false
	}

	// Read first qword as potential pointer
	ptr := uint64(data[0]) | uint64(data[1])<<8 | uint64(data[2])<<16 | uint64(data[3])<<24 |
		uint64(data[4])<<32 | uint64(data[5])<<40 | uint64(data[6])<<48 | uint64(data[7])<<56

	if stubs.Debug {
		stubs.DefaultRegistry.Log("setter-debug", "readStdString",
			fmt.Sprintf("addr=%x ptr=%x data[0:8]=%x data[8:16]=%x data[16:24]=%x",
				addr, ptr, data[0:8], data[8:16], data[16:24]))
	}

	// Try libstdc++ layout first: pointer at offset 0, length stored before data
	// In libstdc++, the string data pointer points to a _Rep struct followed by the char data
	// But for our emulator's malloc, we just have the pointer to char data directly
	if ptr >= 0x90000000 && ptr < 0xA0000000 {
		// Looks like a heap pointer from our emulator
		str, _ := emu.MemReadString(ptr, 256)
		if len(str) > 0 && isPrintable(str) {
			if stubs.Debug {
				stubs.DefaultRegistry.Log("setter-debug", "readStdString",
					fmt.Sprintf("libstdc++ layout: ptr=%x str=%q", ptr, str))
			}
			return str, true
		}
	}

	// Try libc++ SSO layout: bit 0 of first byte indicates long/short
	if data[0]&1 == 0 {
		// Short string: length = data[0] >> 1, chars at offset 1
		length := int(data[0] >> 1)
		if length > 0 && length <= 22 {
			result := string(data[1 : 1+length])
			if isPrintable(result) {
				if stubs.Debug {
					stubs.DefaultRegistry.Log("setter-debug", "readStdString",
						fmt.Sprintf("libc++ SSO: len=%d str=%q", length, result))
				}
				return result, true
			}
		}
	} else {
		// Long string: capacity at 0-7 (with bit 0 set), length at 8-15, pointer at 16-23
		length := uint64(data[8]) | uint64(data[9])<<8 | uint64(data[10])<<16 | uint64(data[11])<<24 |
			uint64(data[12])<<32 | uint64(data[13])<<40 | uint64(data[14])<<48 | uint64(data[15])<<56
		dataPtr := uint64(data[16]) | uint64(data[17])<<8 | uint64(data[18])<<16 | uint64(data[19])<<24 |
			uint64(data[20])<<32 | uint64(data[21])<<40 | uint64(data[22])<<48 | uint64(data[23])<<56

		if stubs.Debug {
			stubs.DefaultRegistry.Log("setter-debug", "readStdString",
				fmt.Sprintf("libc++ long: len=%d ptr=%x", length, dataPtr))
		}

		if length > 0 && length <= 256 && dataPtr != 0 {
			strData, err := emu.MemRead(dataPtr, length)
			if err == nil && isPrintable(string(strData)) {
				return string(strData), true
			}
		}
	}

	return "", false
}

// isLuaSetterSymbol checks if the symbol is a Lua-style setter (const char* params, not std::string)
// Lua setters: ResourcesDecode::setXXTeaKey, LuaStack::setXXTEAKeyAndSign
func isLuaSetterSymbol(name string) bool {
	lower := strings.ToLower(name)
	return (strings.Contains(lower, "setxxteakey") || strings.Contains(lower, "setxxtea")) &&
		!strings.Contains(name, "basic_string") // Exclude std::string versions
}

// makeXXTeaKeyHook creates a hook for XXTEA key setter functions.
// Supports:
// - std::string const& parameters (jsb_set_xxtea_key): X0 = std::string ref
// - Lua-style const char* member methods (ResourcesDecode::setXXTeaKey): X0=this, X1=key_ptr, X2=key_len
// - Lua-style const char* static methods: X0=key_ptr, X1=key_len
func makeXXTeaKeyHook(funcName string) func(*emulator.Emulator) bool {
	return func(emu *emulator.Emulator) bool {
		var key string

		x0 := emu.X(0)
		x1 := emu.X(1)
		x2 := emu.X(2)
		x3 := emu.X(3)
		x4 := emu.X(4)

		if stubs.Debug {
			stubs.DefaultRegistry.Log("setter-debug", funcName,
				fmt.Sprintf("X0=%x X1=%x X2=%x X3=%x X4=%x isLua=%v",
					x0, x1, x2, x3, x4, isLuaSetterSymbol(funcName)))
		}

		// For std::string const& parameters (jsb_set_xxtea_key), X0 points to std::string
		if strings.Contains(funcName, "basic_string") || strings.Contains(funcName, "jsb_set") {
			if str, ok := readStdString(emu, x0); ok && len(str) > 0 && isPrintable(str) {
				key = str
			}
		}

		// Lua-style: ResourcesDecode::setXXTeaKey(const char* key, int keyLen, ...)
		// ResourcesDecode::setXXTeaKey(const char* key1, int len1, const char* key2, int len2)
		// Two calling conventions:
		// 1. Member method: X0=this, X1=key_ptr, X2=key_len, X3=key2_ptr, X4=key2_len
		// 2. Static method: X0=key_ptr, X1=key_len
		if key == "" && isLuaSetterSymbol(funcName) {
			// Heuristic: If X1 is a small number (< 256), likely a length
			// meaning X0 is key_ptr (static method convention)
			if x1 < 256 && x1 > 0 {
				// Static: X0=key_ptr, X1=key_len
				// Read null-terminated string (the length may not include all chars)
				if str, _ := emu.MemReadString(x0, 128); len(str) > 0 && isPrintable(str) {
					key = str
				}
			} else {
				// Member: X0=this, X1=key_ptr
				// Read as null-terminated string - the lengths passed in X2/X4 may be
				// for separate key1/key2 parts, but they're often stored contiguously
				// as one null-terminated string
				if str, _ := emu.MemReadString(x1, 128); len(str) > 0 && isPrintable(str) {
					key = str
				}
			}
		}

		// Fallback: Try as const char* pointer in X0, X1, X2
		if key == "" {
			for _, reg := range []int{0, 1, 2} {
				ptr := emu.X(reg)
				if ptr == 0 || ptr < 0x1000 || ptr > 0x7000000000000000 {
					continue
				}
				str, _ := emu.MemReadString(ptr, 256)
				if len(str) > 0 && isPrintable(str) {
					key = str
					break
				}
			}
		}

		if key != "" {
			captureKey(CapturedKey{
				Value:     key,
				Source:    funcName,
				Address:   emu.PC(),
				KeyType:   "xxtea",
				RiskLevel: "critical",
			})

			// Try to capture signature from X3 (member method: X3=sign_ptr, X4=sign_len)
			if x3 > 0x1000 && x4 > 0 && x4 < 256 {
				if sign, _ := emu.MemReadString(x3, 128); len(sign) > 0 && isPrintable(sign) {
					captureKey(CapturedKey{
						Value:     sign,
						Source:    funcName + "[signature]",
						Address:   emu.PC(),
						KeyType:   "signature",
						RiskLevel: "low",
					})
				}
			}
		}

		stubs.ReturnFromStub(emu)
		return false
	}
}

// makeGenericKeyHook creates a hook for generic key setter functions.
func makeGenericKeyHook(funcName, keyType string) func(*emulator.Emulator) bool {
	return func(emu *emulator.Emulator) bool {
		// Try to find the key in X0, X1, or X2
		for _, reg := range []int{0, 1, 2} {
			keyPtr := emu.X(reg)
			if keyPtr == 0 || keyPtr > 0x7000000000000000 {
				continue
			}

			key, _ := emu.MemReadString(keyPtr, 256)
			if len(key) > 0 && isPrintable(key) {
				captureKey(CapturedKey{
					Value:     key,
					Source:    funcName,
					Address:   emu.PC(),
					KeyType:   keyType,
					RiskLevel: "high",
				})
				break
			}
		}

		stubs.ReturnFromStub(emu)
		return false
	}
}

// makeStdStringSetterHook creates a hook for setter methods taking std::string const& parameters.
// Signature: setCryptoKeyAndSign(std::string const& key, std::string const& sign)
// ARM64 ABI: X0=this, X1=&key_string, X2=&sign_string
func makeStdStringSetterHook(funcName string) func(*emulator.Emulator) bool {
	return func(emu *emulator.Emulator) bool {
		x1 := emu.X(1) // &key
		x2 := emu.X(2) // &sign

		// Try to read key from X1 (std::string const&)
		var key string
		if str, ok := readStdString(emu, x1); ok && len(str) > 0 && isPrintable(str) {
			key = str
		}

		// Try to read signature from X2 (std::string const&)
		var sign string
		if str, ok := readStdString(emu, x2); ok && len(str) > 0 && isPrintable(str) {
			sign = str
		}

		// Fallback: try reading as pointers to const char*
		if key == "" {
			if x1 > 0x1000 && x1 < 0x7000000000000000 {
				if str, _ := emu.MemReadString(x1, 128); len(str) > 0 && isPrintable(str) {
					key = str
				}
			}
		}

		if key != "" {
			captureKey(CapturedKey{
				Value:     key,
				Source:    funcName,
				Address:   emu.PC(),
				KeyType:   "xxtea",
				RiskLevel: "critical",
			})
			if sign != "" {
				captureKey(CapturedKey{
					Value:     sign,
					Source:    funcName + "[signature]",
					Address:   emu.PC(),
					KeyType:   "signature",
					RiskLevel: "low",
				})
			}
		}

		stubs.ReturnFromStub(emu)
		return false
	}
}

// isPrintable checks if a string contains only printable ASCII characters.
func isPrintable(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < 0x20 || c > 0x7e {
			return false
		}
	}
	return true
}
