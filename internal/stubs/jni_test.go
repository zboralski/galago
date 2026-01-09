package stubs

import (
	"testing"

	"github.com/zboralski/galago/internal/emulator"
)

func TestJNIStubsInstantiation(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewJNIStubs(emu)
	if stubs == nil {
		t.Fatal("NewJNIStubs returned nil")
	}
}

func TestJNIInstall(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewJNIStubs(emu)
	jniEnv, javaVM := stubs.Install()

	if jniEnv == 0 {
		t.Error("JNIEnv should not be 0")
	}
	if javaVM == 0 {
		t.Error("JavaVM should not be 0")
	}

	t.Logf("JNIEnv* = 0x%x, JavaVM* = 0x%x", jniEnv, javaVM)

	// Verify JNIEnv points to vtable
	vtablePtr, err := emu.MemReadU64(jniEnv)
	if err != nil {
		t.Fatalf("Failed to read JNIEnv vtable ptr: %v", err)
	}
	if vtablePtr == 0 {
		t.Error("JNIEnv vtable pointer should not be 0")
	}
	t.Logf("JNIEnv->vtable = 0x%x", vtablePtr)

	// Verify JavaVM points to vtable
	vmVtablePtr, err := emu.MemReadU64(javaVM)
	if err != nil {
		t.Fatalf("Failed to read JavaVM vtable ptr: %v", err)
	}
	if vmVtablePtr == 0 {
		t.Error("JavaVM vtable pointer should not be 0")
	}
	t.Logf("JavaVM->vtable = 0x%x", vmVtablePtr)
}

func TestJNIGetVersion(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewJNIStubs(emu)
	jniEnv, _ := stubs.Install()

	// Get vtable
	vtable, _ := emu.MemReadU64(jniEnv)

	// Get GetVersion function pointer (index 4, offset 0x20)
	getVersionAddr, _ := emu.MemReadU64(vtable + 4*8)

	// Set up call
	emu.SetX(0, jniEnv) // JNIEnv*
	emu.SetLR(0xDEADBEEF)

	// Run
	emu.Run(getVersionAddr, getVersionAddr+4)

	// Verify result
	version := emu.X(0)
	if version != JNI_VERSION_1_6 {
		t.Errorf("GetVersion returned 0x%x, expected 0x%x (JNI_VERSION_1_6)", version, JNI_VERSION_1_6)
	}

	t.Logf("GetVersion() = 0x%x", version)
}

func TestJNIFindClass(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewJNIStubs(emu)
	jniEnv, _ := stubs.Install()

	// Track calls
	var lastCall string
	stubs.OnCall = func(name, detail string) {
		lastCall = name + ": " + detail
		t.Logf("JNI call: %s %s", name, detail)
	}

	// Get vtable
	vtable, _ := emu.MemReadU64(jniEnv)

	// Get FindClass function pointer (index 6, offset 0x30)
	findClassAddr, _ := emu.MemReadU64(vtable + 6*8)

	// Write class name to memory
	classNameAddr := emu.Malloc(64)
	emu.MemWriteString(classNameAddr, "com/example/MyClass")

	// Set up call
	emu.SetX(0, jniEnv)       // JNIEnv*
	emu.SetX(1, classNameAddr) // const char* name
	emu.SetLR(0xDEADBEEF)

	// Run
	emu.Run(findClassAddr, findClassAddr+4)

	// Verify result is non-null
	classRef := emu.X(0)
	if classRef == 0 {
		t.Error("FindClass returned null")
	}

	t.Logf("FindClass(\"com/example/MyClass\") = 0x%x", classRef)

	// Verify callback was called
	if lastCall == "" {
		t.Error("OnCall callback was not called")
	}
}

func TestJNINewStringUTF(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewJNIStubs(emu)
	jniEnv, _ := stubs.Install()

	// Get vtable
	vtable, _ := emu.MemReadU64(jniEnv)

	// Get NewStringUTF function pointer (index 167, offset 0x538)
	newStringUTFAddr, _ := emu.MemReadU64(vtable + 167*8)

	// Write test string to memory
	testStr := "Hello, JNI!"
	strAddr := emu.Malloc(64)
	emu.MemWriteString(strAddr, testStr)

	// Set up call
	emu.SetX(0, jniEnv)  // JNIEnv*
	emu.SetX(1, strAddr) // const char* utf
	emu.SetLR(0xDEADBEEF)

	// Run
	emu.Run(newStringUTFAddr, newStringUTFAddr+4)

	// Verify result is non-null
	jstrRef := emu.X(0)
	if jstrRef == 0 {
		t.Error("NewStringUTF returned null")
	}

	// Verify string is tracked
	tracked, ok := stubs.GetTrackedString(jstrRef)
	if !ok {
		t.Error("String not tracked")
	}
	if tracked != testStr {
		t.Errorf("Tracked string mismatch: got %q, want %q", tracked, testStr)
	}

	t.Logf("NewStringUTF(\"%s\") = 0x%x", testStr, jstrRef)
}

func TestJNIGetStringUTFChars(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewJNIStubs(emu)
	jniEnv, _ := stubs.Install()

	// Get vtable
	vtable, _ := emu.MemReadU64(jniEnv)

	// First create a jstring via NewStringUTF
	newStringUTFAddr, _ := emu.MemReadU64(vtable + 167*8)
	testStr := "TestString"
	strAddr := emu.Malloc(64)
	emu.MemWriteString(strAddr, testStr)

	emu.SetX(0, jniEnv)
	emu.SetX(1, strAddr)
	emu.SetLR(0xDEADBEEF)
	emu.Run(newStringUTFAddr, newStringUTFAddr+4)

	jstrRef := emu.X(0)

	// Now test GetStringUTFChars
	getStringUTFCharsAddr, _ := emu.MemReadU64(vtable + 169*8)

	// Allocate space for isCopy
	isCopyAddr := emu.Malloc(8)

	emu.SetX(0, jniEnv)
	emu.SetX(1, jstrRef)
	emu.SetX(2, isCopyAddr)
	emu.SetLR(0xDEADBEEF)
	emu.Run(getStringUTFCharsAddr, getStringUTFCharsAddr+4)

	// Verify result
	charPtr := emu.X(0)
	if charPtr == 0 {
		t.Error("GetStringUTFChars returned null")
	}

	// Read the string back
	readStr, _ := emu.MemReadString(charPtr, 64)
	if readStr != testStr {
		t.Errorf("GetStringUTFChars returned wrong string: got %q, want %q", readStr, testStr)
	}

	// Check isCopy
	isCopy, _ := emu.MemRead(isCopyAddr, 1)
	if isCopy[0] != 1 {
		t.Errorf("isCopy should be 1, got %d", isCopy[0])
	}

	t.Logf("GetStringUTFChars(0x%x) = 0x%x (%q)", jstrRef, charPtr, readStr)
}

func TestJNIGetJavaVM(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewJNIStubs(emu)
	jniEnv, expectedVM := stubs.Install()

	// Get vtable
	vtable, _ := emu.MemReadU64(jniEnv)

	// Get GetJavaVM function pointer (index 219, offset 0x6D8)
	getJavaVMAddr, _ := emu.MemReadU64(vtable + 219*8)

	// Allocate space for JavaVM**
	vmPtrAddr := emu.Malloc(8)

	// Set up call
	emu.SetX(0, jniEnv)    // JNIEnv*
	emu.SetX(1, vmPtrAddr) // JavaVM**
	emu.SetLR(0xDEADBEEF)

	// Run
	emu.Run(getJavaVMAddr, getJavaVMAddr+4)

	// Verify return value
	retVal := emu.X(0)
	if retVal != JNI_OK {
		t.Errorf("GetJavaVM returned %d, expected JNI_OK (0)", retVal)
	}

	// Verify JavaVM* was written
	vmPtr, _ := emu.MemReadU64(vmPtrAddr)
	if vmPtr != expectedVM {
		t.Errorf("GetJavaVM wrote wrong VM pointer: got 0x%x, want 0x%x", vmPtr, expectedVM)
	}

	t.Logf("GetJavaVM() -> 0x%x", vmPtr)
}

func TestJavaVMGetEnv(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewJNIStubs(emu)
	expectedEnv, javaVM := stubs.Install()

	// Get vtable
	vtable, _ := emu.MemReadU64(javaVM)

	// Get GetEnv function pointer (index 3)
	getEnvAddr, _ := emu.MemReadU64(vtable + 3*8)

	// Allocate space for JNIEnv**
	envPtrAddr := emu.Malloc(8)

	// Set up call
	emu.SetX(0, javaVM)       // JavaVM*
	emu.SetX(1, envPtrAddr)   // void** penv
	emu.SetX(2, JNI_VERSION_1_6) // jint version
	emu.SetLR(0xDEADBEEF)

	// Run
	emu.Run(getEnvAddr, getEnvAddr+4)

	// Verify return value
	retVal := emu.X(0)
	if retVal != JNI_OK {
		t.Errorf("GetEnv returned %d, expected JNI_OK (0)", retVal)
	}

	// Verify JNIEnv* was written
	envPtr, _ := emu.MemReadU64(envPtrAddr)
	if envPtr != expectedEnv {
		t.Errorf("GetEnv wrote wrong env pointer: got 0x%x, want 0x%x", envPtr, expectedEnv)
	}

	t.Logf("GetEnv(JNI_VERSION_1_6) -> 0x%x", envPtr)
}

func TestJNIOnLoadSimulation(t *testing.T) {
	// This test simulates what happens in a typical JNI_OnLoad call
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewJNIStubs(emu)
	jniEnv, javaVM := stubs.Install()

	// Track all JNI calls
	var calls []string
	stubs.OnCall = func(name, detail string) {
		calls = append(calls, name)
		t.Logf("JNI: %s %s", name, detail)
	}

	// Simulate typical JNI_OnLoad sequence:
	// 1. Call (*vm)->GetEnv(vm, &env, JNI_VERSION_1_6)

	vmVtable, _ := emu.MemReadU64(javaVM)
	getEnvAddr, _ := emu.MemReadU64(vmVtable + 3*8)

	envPtr := emu.Malloc(8)
	emu.SetX(0, javaVM)
	emu.SetX(1, envPtr)
	emu.SetX(2, JNI_VERSION_1_6)
	emu.SetLR(0xDEADBEEF)
	emu.Run(getEnvAddr, getEnvAddr+4)

	if emu.X(0) != JNI_OK {
		t.Fatal("GetEnv failed")
	}

	// 2. Use JNIEnv to call FindClass
	readEnv, _ := emu.MemReadU64(envPtr)
	if readEnv != jniEnv {
		t.Errorf("GetEnv returned wrong JNIEnv: got 0x%x, want 0x%x", readEnv, jniEnv)
	}

	envVtable, _ := emu.MemReadU64(readEnv)
	findClassAddr, _ := emu.MemReadU64(envVtable + 6*8)

	classNameAddr := emu.Malloc(64)
	emu.MemWriteString(classNameAddr, "com/example/NativeHelper")

	emu.SetX(0, readEnv)
	emu.SetX(1, classNameAddr)
	emu.SetLR(0xDEADBEEF)
	emu.Run(findClassAddr, findClassAddr+4)

	classRef := emu.X(0)
	if classRef == 0 {
		t.Fatal("FindClass returned null")
	}

	// Verify expected calls were made
	expectedCalls := []string{"JavaVM::GetEnv", "JNI::FindClass"}
	for i, expected := range expectedCalls {
		if i >= len(calls) {
			t.Errorf("Missing call: %s", expected)
		} else if calls[i] != expected {
			t.Errorf("Call %d: got %s, want %s", i, calls[i], expected)
		}
	}

	t.Logf("JNI_OnLoad simulation completed with %d JNI calls", len(calls))
}

func TestJNIGetAllTrackedStrings(t *testing.T) {
	emu, err := emulator.New()
	if err != nil {
		t.Fatalf("Failed to create emulator: %v", err)
	}
	defer emu.Close()

	stubs := NewJNIStubs(emu)
	jniEnv, _ := stubs.Install()

	// Get vtable
	vtable, _ := emu.MemReadU64(jniEnv)
	newStringUTFAddr, _ := emu.MemReadU64(vtable + 167*8)

	// Create several strings
	testStrings := []string{"key1", "key2", "secretKey"}
	for _, s := range testStrings {
		strAddr := emu.Malloc(64)
		emu.MemWriteString(strAddr, s)

		emu.SetX(0, jniEnv)
		emu.SetX(1, strAddr)
		emu.SetLR(0xDEADBEEF)
		emu.Run(newStringUTFAddr, newStringUTFAddr+4)
	}

	// Get all tracked strings
	all := stubs.GetAllTrackedStrings()
	if len(all) != len(testStrings) {
		t.Errorf("Expected %d tracked strings, got %d", len(testStrings), len(all))
	}

	for ref, str := range all {
		t.Logf("Tracked JNI string: 0x%x -> %q", ref, str)
	}
}
