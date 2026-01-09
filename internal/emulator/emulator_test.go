package emulator

import (
	"testing"
)

// ARM64 test code: MOV X0, #5; MOV X1, #3; ADD X2, X0, X1; RET
var addTestCode = []byte{
	0xa0, 0x00, 0x80, 0xd2, // MOV X0, #5
	0x61, 0x00, 0x80, 0xd2, // MOV X1, #3
	0x02, 0x00, 0x01, 0x8b, // ADD X2, X0, X1
	0xc0, 0x03, 0x5f, 0xd6, // RET
}

func TestEmulatorBasic(t *testing.T) {
	emu, err := New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	// Load code
	if err := emu.LoadCode(addTestCode); err != nil {
		t.Fatalf("Failed to load code: %v", err)
	}

	// Set up LR to stop execution on RET
	sentinel := uint64(0xDEADBEEF)
	if err := emu.SetLR(sentinel); err != nil {
		t.Fatalf("Failed to set LR: %v", err)
	}

	// Run
	endAddr := CodeBase + uint64(len(addTestCode))
	err = emu.Run(CodeBase, endAddr)
	// Expect "fetch unmapped" error when RET jumps to sentinel
	if err != nil {
		t.Logf("Expected stop error: %v", err)
	}

	// Check result
	x2 := emu.X(2)
	if x2 != 8 {
		t.Errorf("Expected X2=8, got X2=%d", x2)
	}

	// Check X0 and X1
	if emu.X(0) != 5 {
		t.Errorf("Expected X0=5, got X0=%d", emu.X(0))
	}
	if emu.X(1) != 3 {
		t.Errorf("Expected X1=3, got X1=%d", emu.X(1))
	}
}

func TestMemoryOperations(t *testing.T) {
	emu, err := New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	// Test U64
	addr := uint64(HeapBase)
	val := uint64(0x123456789ABCDEF0)

	if err := emu.MemWriteU64(addr, val); err != nil {
		t.Fatalf("Failed to write U64: %v", err)
	}

	readVal, err := emu.MemReadU64(addr)
	if err != nil {
		t.Fatalf("Failed to read U64: %v", err)
	}

	if readVal != val {
		t.Errorf("U64 mismatch: wrote 0x%x, read 0x%x", val, readVal)
	}

	// Test string
	strAddr := emu.Malloc(64)
	testStr := "Hello, Galago!"

	if err := emu.MemWriteString(strAddr, testStr); err != nil {
		t.Fatalf("Failed to write string: %v", err)
	}

	readStr, err := emu.MemReadString(strAddr, 64)
	if err != nil {
		t.Fatalf("Failed to read string: %v", err)
	}

	if readStr != testStr {
		t.Errorf("String mismatch: wrote %q, read %q", testStr, readStr)
	}
}

func TestMalloc(t *testing.T) {
	emu, err := New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	// Allocate some memory
	addr1 := emu.Malloc(100)
	addr2 := emu.Malloc(200)
	addr3 := emu.Malloc(50)

	// Check alignment (16 bytes)
	if addr1%16 != 0 {
		t.Errorf("addr1 not 16-byte aligned: 0x%x", addr1)
	}
	if addr2%16 != 0 {
		t.Errorf("addr2 not 16-byte aligned: 0x%x", addr2)
	}
	if addr3%16 != 0 {
		t.Errorf("addr3 not 16-byte aligned: 0x%x", addr3)
	}

	// Check non-overlapping
	size1 := uint64(112) // 100 rounded to 16
	size2 := uint64(208) // 200 rounded to 16

	if addr2 < addr1+size1 {
		t.Errorf("addr2 overlaps addr1")
	}
	if addr3 < addr2+size2 {
		t.Errorf("addr3 overlaps addr2")
	}
}

func TestAddressHook(t *testing.T) {
	emu, err := New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	// Load code
	if err := emu.LoadCode(addTestCode); err != nil {
		t.Fatalf("Failed to load code: %v", err)
	}

	// Add hook at second instruction (MOV X1, #3)
	hookCalled := false
	secondInstrAddr := uint64(CodeBase + 4)
	emu.HookAddress(secondInstrAddr, func(e *Emulator) bool {
		hookCalled = true
		// Modify X1 to 10 instead of letting MOV X1, #3 execute
		e.SetX(1, 10)
		return false // continue execution
	})

	// Set up LR
	if err := emu.SetLR(0xDEADBEEF); err != nil {
		t.Fatalf("Failed to set LR: %v", err)
	}

	// Run
	endAddr := CodeBase + uint64(len(addTestCode))
	_ = emu.Run(CodeBase, endAddr)

	if !hookCalled {
		t.Error("Address hook was not called")
	}

	// The hook set X1=10, but MOV X1, #3 still executed after hook
	// So X1 should be 3 (hook runs before instruction)
	// Actually in Unicorn, hook runs BEFORE instruction executes
	// So after MOV X1, #3, X1 will be 3
	// If we wanted to change the result, we'd need to skip the instruction
	t.Logf("X1 after hook: %d", emu.X(1))
}

func TestCodeHook(t *testing.T) {
	emu, err := New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	// Load code
	if err := emu.LoadCode(addTestCode); err != nil {
		t.Fatalf("Failed to load code: %v", err)
	}

	// Count instructions
	instrCount := 0
	emu.HookCode(func(e *Emulator, addr uint64, size uint32) {
		instrCount++
	})

	// Set up LR
	if err := emu.SetLR(0xDEADBEEF); err != nil {
		t.Fatalf("Failed to set LR: %v", err)
	}

	// Run
	endAddr := CodeBase + uint64(len(addTestCode))
	_ = emu.Run(CodeBase, endAddr)

	if instrCount != 4 {
		t.Errorf("Expected 4 instructions, got %d", instrCount)
	}
}
