package emulator

import (
	"os"
	"testing"
)

// TestELFLoader tests loading a real ARM64 ELF file
func TestELFLoader(t *testing.T) {
	// Try to find a test binary
	testPaths := []string{
		"/Volumes/tank4a - Data/Users/Shared/re/unityx/samples/com.ivar.mafia/libil2cpp.so",
		"/Users/az/re/samples/libil2cpp.so",
		os.ExpandEnv("$HOME/re/samples/libil2cpp.so"),
	}

	var testPath string
	for _, p := range testPaths {
		if _, err := os.Stat(p); err == nil {
			testPath = p
			break
		}
	}

	if testPath == "" {
		t.Skip("No test binary found, skipping ELF loader test")
	}

	t.Logf("Testing with: %s", testPath)

	emu, err := New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	info, err := emu.LoadELF(testPath)
	if err != nil {
		t.Fatalf("Failed to load ELF: %v", err)
	}

	t.Logf("ELF loaded successfully:")
	t.Logf("  Base address: 0x%x", info.BaseAddr)
	t.Logf("  End address:  0x%x", info.EndAddr)
	t.Logf("  Entry point:  0x%x", info.Entry)
	t.Logf("  Symbols:      %d", len(info.Symbols))
	t.Logf("  Segments:     %d", len(info.Segments))

	// Check for JNI_OnLoad
	jniOnLoad := info.FindJNIOnLoad()
	if jniOnLoad != 0 {
		t.Logf("  JNI_OnLoad:   0x%x", jniOnLoad)
	} else {
		t.Log("  JNI_OnLoad:   not found")
	}

	// Verify base address is reasonable
	if info.BaseAddr == 0 || info.BaseAddr > 0xFFFFFFFF {
		t.Errorf("Suspicious base address: 0x%x", info.BaseAddr)
	}

	// Verify we have segments
	if len(info.Segments) == 0 {
		t.Error("No segments loaded")
	}

	// Verify we can read memory at the base address
	data, err := emu.MemRead(info.BaseAddr, 4)
	if err != nil {
		t.Errorf("Failed to read memory at base: %v", err)
	}

	// ELF magic: 0x7f 'E' 'L' 'F'
	if len(data) >= 4 && data[0] == 0x7f && data[1] == 'E' && data[2] == 'L' && data[3] == 'F' {
		t.Log("  ELF magic verified at base address")
	} else {
		t.Logf("  Data at base: %x (may not be ELF header)", data)
	}

	// Check for some Unity/IL2CPP specific symbols
	il2cppSyms := info.FindSymbolsBySubstring("il2cpp")
	if len(il2cppSyms) > 0 {
		t.Logf("  IL2CPP symbols: %d found", len(il2cppSyms))
		// Show a few examples
		count := 0
		for name, addr := range il2cppSyms {
			if count < 5 {
				t.Logf("    %s @ 0x%x", name, addr)
				count++
			}
		}
	}
}

func TestFindEntryPoint(t *testing.T) {
	info := &ELFInfo{
		Entry: 0x1000,
		Symbols: map[string]uint64{
			"JNI_OnLoad":  0x2000,
			"il2cpp_init": 0x3000,
			"some_func":   0x4000,
		},
	}

	// Should prefer JNI_OnLoad
	entry := info.FindEntryPoint("")
	if entry != 0x2000 {
		t.Errorf("Expected JNI_OnLoad (0x2000), got 0x%x", entry)
	}

	// Should use preferred entry if specified
	entry = info.FindEntryPoint("il2cpp_init")
	if entry != 0x3000 {
		t.Errorf("Expected il2cpp_init (0x3000), got 0x%x", entry)
	}

	// Case-insensitive
	entry = info.FindEntryPoint("JNI_ONLOAD")
	if entry != 0x2000 {
		t.Errorf("Expected JNI_OnLoad (0x2000) case-insensitive, got 0x%x", entry)
	}

	// Test without JNI_OnLoad - should use priority-based selection
	info2 := &ELFInfo{
		Entry: 0x1000,
		Symbols: map[string]uint64{
			"il2cpp_init": 0x3000, // Not in priority list
		},
	}
	entry = info2.FindEntryPoint("")
	// il2cpp_init is NOT in the priority list, so fallback to ELF Entry
	if entry != 0x1000 {
		t.Errorf("Expected ELF entry (0x1000) as fallback, got 0x%x", entry)
	}

	// Test with priority-based symbols
	info3 := &ELFInfo{
		Entry: 0x1000,
		Symbols: map[string]uint64{
			"_ZN11AppDelegate30applicationDidFinishLaunchingEv": 0x5000, // Priority 1
			"JNI_OnLoad": 0x2000, // Priority 7
		},
	}
	entry = info3.FindEntryPoint("")
	if entry != 0x5000 {
		t.Errorf("Expected AppDelegate (0x5000) over JNI_OnLoad, got 0x%x", entry)
	}
}
