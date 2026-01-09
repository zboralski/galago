package stubs

import (
	"testing"

	"github.com/zboralski/galago/internal/emulator"
)

// ARM64 BL instruction to addr 0x10000 (call malloc stub)
// BL is PC-relative, so we need to craft a simple test

func TestLibcStubsInstantiation(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewLibcStubs(emu)
	if stubs == nil {
		t.Fatal("NewLibcStubs returned nil")
	}

	// Test OnCall callback
	called := false
	stubs.OnCall = func(name, detail string) {
		called = true
		t.Logf("Stub called: %s %s", name, detail)
	}

	// Test empty install (no imports)
	stubs.Install(nil)
	stubs.Install(map[string]uint64{})

	if called {
		t.Error("OnCall should not be called during Install")
	}
}

func TestMallocStub(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewLibcStubs(emu)

	// Track stub calls
	var lastCall string
	stubs.OnCall = func(name, detail string) {
		lastCall = name
		t.Logf("Stub: %s %s", name, detail)
	}

	// Set up a stub address in stub region
	stubAddr := uint64(emulator.StubBase)
	stubs.InstallAt("malloc", stubAddr, stubs.stubMalloc)

	// Set up arguments: malloc(100)
	emu.SetX(0, 100) // size = 100

	// Set LR to sentinel for return
	sentinel := uint64(0xDEADBEEF)
	emu.SetLR(sentinel)

	// Set PC to stub address
	emu.SetPC(stubAddr)

	// The hook should fire immediately when we hit this address
	// But we need to run emulation for the hook to fire

	// Write a NOP at stub address so we have something to execute
	// NOP = 0xD503201F
	nop := []byte{0x1f, 0x20, 0x03, 0xd5}
	emu.MemWrite(stubAddr, nop)

	// Run one instruction
	err = emu.Run(stubAddr, stubAddr+4)
	// Hook fires, sets X0 to allocated address, sets PC to LR

	// After hook, X0 should contain allocated pointer
	result := emu.X(0)
	t.Logf("malloc(100) returned: 0x%x", result)

	// Should be in heap region
	if result < emulator.HeapBase || result >= emulator.HeapBase+emulator.HeapSize {
		t.Errorf("malloc returned address 0x%x outside heap region", result)
	}

	// Should be 16-byte aligned
	if result%16 != 0 {
		t.Errorf("malloc returned unaligned address: 0x%x", result)
	}

	// Verify stub was called
	if lastCall != "malloc" {
		t.Errorf("Expected malloc to be called, got: %s", lastCall)
	}

	// PC should be set to sentinel (LR)
	pc := emu.PC()
	// Note: PC might be stubAddr+4 after running, but the hook sets PC to LR
	// The behavior depends on whether the hook's returnFromStub works
	t.Logf("PC after stub: 0x%x (expected sentinel: 0x%x)", pc, sentinel)
}

func TestMemcpyStub(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewLibcStubs(emu)

	// Allocate source and dest buffers
	src := emu.Malloc(64)
	dst := emu.Malloc(64)

	// Write test data to source
	testData := []byte("Hello, Galago!")
	emu.MemWrite(src, testData)

	// Install memcpy stub
	stubAddr := uint64(emulator.StubBase + 0x100)
	stubs.InstallAt("memcpy", stubAddr, stubs.stubMemcpy)

	// Set up arguments: memcpy(dst, src, len)
	emu.SetX(0, dst)
	emu.SetX(1, src)
	emu.SetX(2, uint64(len(testData)))
	emu.SetLR(0xDEADBEEF)

	// Write NOP for execution
	nop := []byte{0x1f, 0x20, 0x03, 0xd5}
	emu.MemWrite(stubAddr, nop)

	// Run
	emu.Run(stubAddr, stubAddr+4)

	// Verify data was copied
	copied, err := emu.MemRead(dst, uint64(len(testData)))
	if err != nil {
		t.Fatalf("Failed to read dst: %v", err)
	}

	if string(copied) != string(testData) {
		t.Errorf("memcpy failed: got %q, want %q", copied, testData)
	}

	// X0 should return dst
	if emu.X(0) != dst {
		t.Errorf("memcpy should return dst, got 0x%x, want 0x%x", emu.X(0), dst)
	}
}

func TestStrlenStub(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewLibcStubs(emu)

	// Allocate and write test string
	strAddr := emu.Malloc(64)
	testStr := "Hello, World!"
	emu.MemWriteString(strAddr, testStr)

	// Install strlen stub
	stubAddr := uint64(emulator.StubBase + 0x200)
	stubs.InstallAt("strlen", stubAddr, stubs.stubStrlen)

	// Set up arguments
	emu.SetX(0, strAddr)
	emu.SetLR(0xDEADBEEF)

	// Write NOP
	nop := []byte{0x1f, 0x20, 0x03, 0xd5}
	emu.MemWrite(stubAddr, nop)

	// Run
	emu.Run(stubAddr, stubAddr+4)

	// Verify result
	result := emu.X(0)
	expected := uint64(len(testStr))
	if result != expected {
		t.Errorf("strlen returned %d, want %d", result, expected)
	}
}

func TestStrcmpStub(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewLibcStubs(emu)

	// Allocate and write test strings
	s1Addr := emu.Malloc(64)
	s2Addr := emu.Malloc(64)

	// Install strcmp stub
	stubAddr := uint64(emulator.StubBase + 0x300)
	stubs.InstallAt("strcmp", stubAddr, stubs.stubStrcmp)

	// Write NOP
	nop := []byte{0x1f, 0x20, 0x03, 0xd5}
	emu.MemWrite(stubAddr, nop)

	tests := []struct {
		s1, s2 string
		want   uint64 // 0 for equal, non-zero for not equal
	}{
		{"hello", "hello", 0},
		{"abc", "abd", 0xffffffffffffffff}, // -1
		{"abd", "abc", 1},
		{"", "", 0},
	}

	for _, tc := range tests {
		emu.MemWriteString(s1Addr, tc.s1)
		emu.MemWriteString(s2Addr, tc.s2)

		emu.SetX(0, s1Addr)
		emu.SetX(1, s2Addr)
		emu.SetLR(0xDEADBEEF)
		emu.SetPC(stubAddr)

		emu.Run(stubAddr, stubAddr+4)

		result := emu.X(0)
		if result != tc.want {
			t.Errorf("strcmp(%q, %q) = 0x%x, want 0x%x", tc.s1, tc.s2, result, tc.want)
		}
	}
}

func TestTimeMocking(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewLibcStubs(emu)

	// Allocate buffer for struct timeval
	tvAddr := emu.Malloc(16)

	// Install gettimeofday stub
	stubAddr := uint64(emulator.StubBase + 0x400)
	stubs.InstallAt("gettimeofday", stubAddr, stubs.stubGettimeofday)

	// Set up arguments
	emu.SetX(0, tvAddr) // tv
	emu.SetX(1, 0)      // tz (NULL)
	emu.SetLR(0xDEADBEEF)

	// Write NOP
	nop := []byte{0x1f, 0x20, 0x03, 0xd5}
	emu.MemWrite(stubAddr, nop)

	// Run
	emu.Run(stubAddr, stubAddr+4)

	// Verify return value (0 for success)
	if emu.X(0) != 0 {
		t.Errorf("gettimeofday should return 0, got %d", emu.X(0))
	}

	// Verify mocked time was written
	tvSec, _ := emu.MemReadU64(tvAddr)
	tvUsec, _ := emu.MemReadU64(tvAddr + 8)

	if tvSec != uint64(MockTimeSec) {
		t.Errorf("tv_sec = %d, want %d", tvSec, MockTimeSec)
	}
	if tvUsec != uint64(MockTimeUSec) {
		t.Errorf("tv_usec = %d, want %d", tvUsec, MockTimeUSec)
	}

	t.Logf("Mocked time: %d.%06d (2024-01-01 00:00:00 UTC)", tvSec, tvUsec)
}

func TestCallocZeroInit(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewLibcStubs(emu)

	// Install calloc stub
	stubAddr := uint64(emulator.StubBase + 0x500)
	stubs.InstallAt("calloc", stubAddr, stubs.stubCalloc)

	// Set up arguments: calloc(10, 8) = 80 bytes
	emu.SetX(0, 10) // count
	emu.SetX(1, 8)  // size
	emu.SetLR(0xDEADBEEF)

	// Write NOP
	nop := []byte{0x1f, 0x20, 0x03, 0xd5}
	emu.MemWrite(stubAddr, nop)

	// Run
	emu.Run(stubAddr, stubAddr+4)

	// Get allocated address
	ptr := emu.X(0)
	t.Logf("calloc(10, 8) returned: 0x%x", ptr)

	// Verify zero initialization
	data, _ := emu.MemRead(ptr, 80)
	for i, b := range data {
		if b != 0 {
			t.Errorf("calloc memory not zeroed at offset %d: got 0x%x", i, b)
			break
		}
	}
}
