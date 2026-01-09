package stubs

import (
	"testing"

	"github.com/zboralski/galago/internal/emulator"
)

func TestCxxAbiStubsInstantiation(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewCxxAbiStubs(emu)
	if stubs == nil {
		t.Fatal("NewCxxAbiStubs returned nil")
	}

	// Test empty install
	stubs.Install(nil)
	stubs.Install(map[string]uint64{})
}

func TestSSOShortString(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewCxxAbiStubs(emu)

	// Allocate memory for std::string object
	strObj := emu.Malloc(24)

	// Test short string (< 23 chars)
	testStr := "Hello, World!"
	err = stubs.WriteSSOString(strObj, testStr)
	if err != nil {
		t.Fatalf("WriteSSOString failed: %v", err)
	}

	// Read it back
	result, ok := stubs.ReadSSOString(strObj)
	if !ok {
		t.Fatal("ReadSSOString failed")
	}

	if result != testStr {
		t.Errorf("String mismatch: got %q, want %q", result, testStr)
	}

	// Verify data pointer (should be inline for short string)
	dataPtr := stubs.GetSSODataPtr(strObj)
	if dataPtr != strObj+1 {
		t.Errorf("Short string data ptr should be inline: got 0x%x, want 0x%x", dataPtr, strObj+1)
	}

	// Verify length
	length := stubs.GetSSOLength(strObj)
	if length != uint64(len(testStr)) {
		t.Errorf("Length mismatch: got %d, want %d", length, len(testStr))
	}

	t.Logf("Short SSO string at 0x%x: %q (len=%d)", strObj, result, length)
}

func TestSSOLongString(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewCxxAbiStubs(emu)

	// Allocate memory for std::string object
	strObj := emu.Malloc(24)

	// Test long string (>= 23 chars)
	testStr := "This is a very long string that exceeds the SSO capacity of 22 characters!"
	err = stubs.WriteSSOString(strObj, testStr)
	if err != nil {
		t.Fatalf("WriteSSOString failed: %v", err)
	}

	// Read it back
	result, ok := stubs.ReadSSOString(strObj)
	if !ok {
		t.Fatal("ReadSSOString failed")
	}

	if result != testStr {
		t.Errorf("String mismatch: got %q, want %q", result, testStr)
	}

	// Verify data pointer (should be heap-allocated for long string)
	dataPtr := stubs.GetSSODataPtr(strObj)
	if dataPtr == strObj+1 {
		t.Error("Long string data ptr should NOT be inline")
	}
	if dataPtr < emulator.HeapBase {
		t.Errorf("Long string data ptr should be in heap: got 0x%x", dataPtr)
	}

	// Verify length
	length := stubs.GetSSOLength(strObj)
	if length != uint64(len(testStr)) {
		t.Errorf("Length mismatch: got %d, want %d", length, len(testStr))
	}

	t.Logf("Long SSO string at 0x%x: data at 0x%x, len=%d", strObj, dataPtr, length)
}

func TestSSOEmptyString(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewCxxAbiStubs(emu)

	// Allocate and zero memory (simulating .bss)
	strObj := emu.Malloc(24)
	zeros := make([]byte, 24)
	emu.MemWrite(strObj, zeros)

	// Read empty string (all zeros = valid empty SSO string)
	result, ok := stubs.ReadSSOString(strObj)
	if !ok {
		t.Fatal("ReadSSOString failed for empty string")
	}

	if result != "" {
		t.Errorf("Empty string mismatch: got %q, want empty", result)
	}

	length := stubs.GetSSOLength(strObj)
	if length != 0 {
		t.Errorf("Empty string length should be 0, got %d", length)
	}
}

func TestStringCtorStub(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewCxxAbiStubs(emu)

	// Track captured strings
	var capturedStrings []string
	stubs.OnStringCapture = func(addr uint64, value string) {
		capturedStrings = append(capturedStrings, value)
		t.Logf("Captured string at 0x%x: %q", addr, value)
	}

	// Allocate std::string object and source string
	strObj := emu.Malloc(24)
	srcStr := "XXTEA_KEY_12345"
	srcAddr := emu.Malloc(uint64(len(srcStr) + 1))
	emu.MemWriteString(srcAddr, srcStr)

	// Install constructor stub
	stubAddr := uint64(emulator.StubBase + 0x1000)
	stubs.InstallAt("string::ctor", stubAddr, stubs.stubStringCtor)

	// Set up arguments
	emu.SetX(0, strObj)  // this
	emu.SetX(1, srcAddr) // const char*
	emu.SetLR(0xDEADBEEF)

	// Write NOP
	nop := []byte{0x1f, 0x20, 0x03, 0xd5}
	emu.MemWrite(stubAddr, nop)

	// Run
	emu.Run(stubAddr, stubAddr+4)

	// Verify string was constructed
	result, ok := stubs.ReadSSOString(strObj)
	if !ok {
		t.Fatal("Failed to read constructed string")
	}
	if result != srcStr {
		t.Errorf("Constructed string mismatch: got %q, want %q", result, srcStr)
	}

	// Verify callback was called
	if len(capturedStrings) != 1 || capturedStrings[0] != srcStr {
		t.Errorf("Expected callback with %q, got %v", srcStr, capturedStrings)
	}

	// Verify tracking
	tracked, ok := stubs.GetTrackedString(strObj)
	if !ok || tracked != srcStr {
		t.Errorf("String not tracked properly: got %q, want %q", tracked, srcStr)
	}
}

func TestStringAssignStub(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewCxxAbiStubs(emu)

	// Allocate std::string object and source string
	strObj := emu.Malloc(24)
	srcStr := "CRYPTO_SECRET"
	srcAddr := emu.Malloc(uint64(len(srcStr) + 1))
	emu.MemWriteString(srcAddr, srcStr)

	// Install assign stub
	stubAddr := uint64(emulator.StubBase + 0x2000)
	stubs.InstallAt("string::assign", stubAddr, stubs.stubStringAssign)

	// Set up arguments
	emu.SetX(0, strObj)  // this
	emu.SetX(1, srcAddr) // const char*
	emu.SetLR(0xDEADBEEF)

	// Write NOP
	nop := []byte{0x1f, 0x20, 0x03, 0xd5}
	emu.MemWrite(stubAddr, nop)

	// Run
	emu.Run(stubAddr, stubAddr+4)

	// Verify string was assigned
	result, ok := stubs.ReadSSOString(strObj)
	if !ok {
		t.Fatal("Failed to read assigned string")
	}
	if result != srcStr {
		t.Errorf("Assigned string mismatch: got %q, want %q", result, srcStr)
	}

	// Verify return value is this
	if emu.X(0) != strObj {
		t.Errorf("assign should return this: got 0x%x, want 0x%x", emu.X(0), strObj)
	}
}

func TestStringCStrStub(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewCxxAbiStubs(emu)

	// Create a short string first
	strObj := emu.Malloc(24)
	testStr := "Hello"
	stubs.WriteSSOString(strObj, testStr)

	// Install c_str stub
	stubAddr := uint64(emulator.StubBase + 0x3000)
	stubs.InstallAt("string::c_str", stubAddr, stubs.stubStringCStr)

	// Set up arguments
	emu.SetX(0, strObj)
	emu.SetLR(0xDEADBEEF)

	// Write NOP
	nop := []byte{0x1f, 0x20, 0x03, 0xd5}
	emu.MemWrite(stubAddr, nop)

	// Run
	emu.Run(stubAddr, stubAddr+4)

	// Verify return value is data pointer
	dataPtr := emu.X(0)
	expectedPtr := strObj + 1 // Short string inline data
	if dataPtr != expectedPtr {
		t.Errorf("c_str returned 0x%x, expected 0x%x", dataPtr, expectedPtr)
	}

	// Read the string data
	readStr, _ := emu.MemReadString(dataPtr, 64)
	if readStr != testStr {
		t.Errorf("String data mismatch: got %q, want %q", readStr, testStr)
	}
}

func TestStringSizeStub(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewCxxAbiStubs(emu)

	// Create a string
	strObj := emu.Malloc(24)
	testStr := "TestString"
	stubs.WriteSSOString(strObj, testStr)

	// Install size stub
	stubAddr := uint64(emulator.StubBase + 0x4000)
	stubs.InstallAt("string::size", stubAddr, stubs.stubStringSize)

	// Set up arguments
	emu.SetX(0, strObj)
	emu.SetLR(0xDEADBEEF)

	// Write NOP
	nop := []byte{0x1f, 0x20, 0x03, 0xd5}
	emu.MemWrite(stubAddr, nop)

	// Run
	emu.Run(stubAddr, stubAddr+4)

	// Verify return value
	size := emu.X(0)
	if size != uint64(len(testStr)) {
		t.Errorf("size() returned %d, expected %d", size, len(testStr))
	}
}

func TestCxaGuardStubs(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewCxxAbiStubs(emu)

	// Allocate guard variable
	guard := emu.Malloc(8)

	// Install guard stubs
	acquireAddr := uint64(emulator.StubBase + 0x5000)
	releaseAddr := uint64(emulator.StubBase + 0x5010)
	stubs.InstallAt("guard_acquire", acquireAddr, stubs.stubCxaGuardAcquire)
	stubs.InstallAt("guard_release", releaseAddr, stubs.stubCxaGuardRelease)

	nop := []byte{0x1f, 0x20, 0x03, 0xd5}
	emu.MemWrite(acquireAddr, nop)
	emu.MemWrite(releaseAddr, nop)

	// First acquire should return 1 (need to initialize)
	emu.SetX(0, guard)
	emu.SetLR(0xDEADBEEF)
	emu.Run(acquireAddr, acquireAddr+4)

	if emu.X(0) != 1 {
		t.Errorf("First guard_acquire should return 1, got %d", emu.X(0))
	}

	// Release the guard
	emu.SetX(0, guard)
	emu.SetLR(0xDEADBEEF)
	emu.Run(releaseAddr, releaseAddr+4)

	// Second acquire should return 0 (already initialized)
	emu.SetX(0, guard)
	emu.SetLR(0xDEADBEEF)
	emu.Run(acquireAddr, acquireAddr+4)

	if emu.X(0) != 0 {
		t.Errorf("Second guard_acquire should return 0, got %d", emu.X(0))
	}
}

func TestCxaAtexitStub(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewCxxAbiStubs(emu)

	// Install atexit stub
	stubAddr := uint64(emulator.StubBase + 0x6000)
	stubs.InstallAt("__cxa_atexit", stubAddr, stubs.stubCxaAtexit)

	nop := []byte{0x1f, 0x20, 0x03, 0xd5}
	emu.MemWrite(stubAddr, nop)

	// Set up arguments
	emu.SetX(0, 0x1000) // destructor func
	emu.SetX(1, 0x2000) // arg
	emu.SetX(2, 0x3000) // dso handle
	emu.SetLR(0xDEADBEEF)

	// Run
	emu.Run(stubAddr, stubAddr+4)

	// Should return 0 for success
	if emu.X(0) != 0 {
		t.Errorf("__cxa_atexit should return 0, got %d", emu.X(0))
	}
}

func TestGetAllTrackedStrings(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewCxxAbiStubs(emu)

	// Write several strings
	addr1 := emu.Malloc(24)
	addr2 := emu.Malloc(24)
	addr3 := emu.Malloc(24)

	stubs.WriteSSOString(addr1, "key1")
	stubs.strings[addr1] = "key1"

	stubs.WriteSSOString(addr2, "key2")
	stubs.strings[addr2] = "key2"

	stubs.WriteSSOString(addr3, "key3")
	stubs.strings[addr3] = "key3"

	// Get all tracked strings
	all := stubs.GetAllTrackedStrings()

	if len(all) != 3 {
		t.Errorf("Expected 3 tracked strings, got %d", len(all))
	}

	for addr, str := range all {
		t.Logf("Tracked: 0x%x -> %q", addr, str)
	}
}
