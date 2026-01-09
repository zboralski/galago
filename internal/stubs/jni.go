// Package stubs provides JNI/JavaVM mock implementations.
// This file provides fake JNIEnv and JavaVM structures for emulating Android native libraries.
package stubs

import (
	"github.com/zboralski/galago/internal/emulator"
)

// JNI Constants
const (
	JNI_OK        = 0
	JNI_ERR       = -1
	JNI_EDETACHED = -2
	JNI_EVERSION  = -3

	JNI_VERSION_1_1 = 0x00010001
	JNI_VERSION_1_2 = 0x00010002
	JNI_VERSION_1_4 = 0x00010004
	JNI_VERSION_1_6 = 0x00010006
)

// JNI Function Indices (offset / 8 in JNINativeInterface)
const (
	JNI_GetVersion          = 4   // offset 0x20
	JNI_DefineClass         = 5   // offset 0x28
	JNI_FindClass           = 6   // offset 0x30
	JNI_FromReflectedMethod = 7   // offset 0x38
	JNI_FromReflectedField  = 8   // offset 0x40
	JNI_ToReflectedMethod   = 9   // offset 0x48
	JNI_GetSuperclass       = 10  // offset 0x50
	JNI_GetObjectClass      = 31  // offset 0xF8
	JNI_GetMethodID         = 33  // offset 0x108
	JNI_CallObjectMethod    = 34  // offset 0x110
	JNI_GetStaticMethodID   = 113 // offset 0x388
	JNI_NewStringUTF        = 167 // offset 0x538
	JNI_GetStringUTFChars   = 169 // offset 0x548
	JNI_ReleaseStringUTFChars = 170 // offset 0x550
	JNI_GetArrayLength      = 171 // offset 0x558
	JNI_GetJavaVM           = 219 // offset 0x6D8
	JNI_FUNC_COUNT          = 300 // Total JNI function slots
)

// JavaVM Function Indices
const (
	JAVAVM_DestroyJavaVM  = 0 // Not typically called
	JAVAVM_AttachCurrentThread = 1
	JAVAVM_DetachCurrentThread = 2
	JAVAVM_GetEnv         = 3 // Main entry point
	JAVAVM_AttachCurrentThreadAsDaemon = 4
	JAVAVM_FUNC_COUNT     = 10
)

// JNIStubs provides stub implementations for JNI/JavaVM functions.
type JNIStubs struct {
	emu *emulator.Emulator

	// Trace callback for logging stub calls
	OnCall func(name string, detail string)

	// Memory layout
	jniEnvBase     uint64 // JNIEnv* (points to vtable pointer)
	jniVtableBase  uint64 // JNINativeInterface vtable
	jniStubBase    uint64 // JNI function stubs
	javaVMBase     uint64 // JavaVM*
	javaVMVtable   uint64 // JNIInvokeInterface vtable
	javaVMStubBase uint64 // JavaVM function stubs
	mockObjBase    uint64 // Fake object references

	// String tracking for NewStringUTF
	jniStrings map[uint64]string
	nextStringRef uint64

	// Class references
	classRefs map[string]uint64
	nextClassRef uint64
}

// NewJNIStubs creates JNI/JavaVM stubs for an emulator.
func NewJNIStubs(emu *emulator.Emulator) *JNIStubs {
	return &JNIStubs{
		emu:          emu,
		jniStrings:   make(map[uint64]string),
		classRefs:    make(map[string]uint64),
		nextStringRef: 0x1000, // Starting fake reference for jstrings
		nextClassRef:  0x2000, // Starting fake reference for jclasses
	}
}

// Install sets up the JNI environment in emulator memory.
// Returns JNIEnv* and JavaVM* pointers for use in entry point calls.
func (s *JNIStubs) Install() (jniEnv, javaVM uint64) {
	// Layout in StubBase region:
	// StubBase + 0x0000: JNIEnv structure (8 bytes - pointer to vtable)
	// StubBase + 0x1000: JNINativeInterface vtable (300 * 8 = 2400 bytes)
	// StubBase + 0x2000: JNI function stubs (300 * 4 = 1200 bytes)
	// StubBase + 0x3000: JavaVM structure (8 bytes - pointer to vtable)
	// StubBase + 0x4000: JNIInvokeInterface vtable (10 * 8 = 80 bytes)
	// StubBase + 0x5000: JavaVM function stubs (10 * 4 = 40 bytes)
	// StubBase + 0x6000: Mock object pool

	base := uint64(emulator.StubBase) + 0x10000 // Leave room for libc/cxx stubs

	s.jniEnvBase = base
	s.jniVtableBase = base + 0x1000
	s.jniStubBase = base + 0x2000
	s.javaVMBase = base + 0x3000
	s.javaVMVtable = base + 0x4000
	s.javaVMStubBase = base + 0x5000
	s.mockObjBase = base + 0x6000

	// RET instruction for stubs
	retInsn := []byte{0xc0, 0x03, 0x5f, 0xd6} // RET

	// Create JNI function stubs and vtable
	for i := 0; i < JNI_FUNC_COUNT; i++ {
		stubAddr := s.jniStubBase + uint64(i*4)

		// Write RET instruction at stub address
		s.emu.MemWrite(stubAddr, retInsn)

		// Write vtable entry (pointer to stub)
		s.emu.MemWriteU64(s.jniVtableBase+uint64(i*8), stubAddr)

		// Install specific hook handlers
		switch i {
		case JNI_GetVersion:
			s.emu.HookAddress(stubAddr, s.makeStubHandler("JNI::GetVersion", s.stubGetVersion))
		case JNI_FindClass:
			s.emu.HookAddress(stubAddr, s.makeStubHandler("JNI::FindClass", s.stubFindClass))
		case JNI_GetMethodID:
			s.emu.HookAddress(stubAddr, s.makeStubHandler("JNI::GetMethodID", s.stubGetMethodID))
		case JNI_GetStaticMethodID:
			s.emu.HookAddress(stubAddr, s.makeStubHandler("JNI::GetStaticMethodID", s.stubGetStaticMethodID))
		case JNI_GetObjectClass:
			s.emu.HookAddress(stubAddr, s.makeStubHandler("JNI::GetObjectClass", s.stubGetObjectClass))
		case JNI_NewStringUTF:
			s.emu.HookAddress(stubAddr, s.makeStubHandler("JNI::NewStringUTF", s.stubNewStringUTF))
		case JNI_GetStringUTFChars:
			s.emu.HookAddress(stubAddr, s.makeStubHandler("JNI::GetStringUTFChars", s.stubGetStringUTFChars))
		case JNI_ReleaseStringUTFChars:
			s.emu.HookAddress(stubAddr, s.makeStubHandler("JNI::ReleaseStringUTFChars", s.stubReleaseStringUTFChars))
		case JNI_GetJavaVM:
			s.emu.HookAddress(stubAddr, s.makeStubHandler("JNI::GetJavaVM", s.stubGetJavaVM))
		case JNI_CallObjectMethod:
			s.emu.HookAddress(stubAddr, s.makeStubHandler("JNI::CallObjectMethod", s.stubCallObjectMethod))
		default:
			s.emu.HookAddress(stubAddr, s.makeStubHandler("JNI::Generic", s.stubJNIGeneric))
		}
	}

	// Set up JNIEnv structure: first pointer is to vtable
	s.emu.MemWriteU64(s.jniEnvBase, s.jniVtableBase)

	// Create JavaVM function stubs and vtable
	for i := 0; i < JAVAVM_FUNC_COUNT; i++ {
		stubAddr := s.javaVMStubBase + uint64(i*4)

		// Write RET instruction
		s.emu.MemWrite(stubAddr, retInsn)

		// Write vtable entry
		s.emu.MemWriteU64(s.javaVMVtable+uint64(i*8), stubAddr)

		// Install specific handlers
		switch i {
		case JAVAVM_GetEnv:
			s.emu.HookAddress(stubAddr, s.makeStubHandler("JavaVM::GetEnv", s.stubJavaVMGetEnv))
		case JAVAVM_AttachCurrentThread:
			s.emu.HookAddress(stubAddr, s.makeStubHandler("JavaVM::AttachCurrentThread", s.stubAttachCurrentThread))
		case JAVAVM_DetachCurrentThread:
			s.emu.HookAddress(stubAddr, s.makeStubHandler("JavaVM::DetachCurrentThread", s.stubDetachCurrentThread))
		default:
			s.emu.HookAddress(stubAddr, s.makeStubHandler("JavaVM::Generic", s.stubJavaVMGeneric))
		}
	}

	// Set up JavaVM structure: first pointer is to vtable
	s.emu.MemWriteU64(s.javaVMBase, s.javaVMVtable)

	return s.jniEnvBase, s.javaVMBase
}

// makeStubHandler creates a hook handler with logging.
func (s *JNIStubs) makeStubHandler(name string, handler func() string) func(*emulator.Emulator) bool {
	return func(e *emulator.Emulator) bool {
		detail := handler()
		if s.OnCall != nil {
			s.OnCall(name, detail)
		}
		return false
	}
}

func (s *JNIStubs) returnFromStub() {
	lr := s.emu.LR()
	s.emu.SetPC(lr)
}

// GetJNIEnv returns the JNIEnv* pointer for use in function calls.
func (s *JNIStubs) GetJNIEnv() uint64 {
	return s.jniEnvBase
}

// GetJavaVM returns the JavaVM* pointer for use in JNI_OnLoad.
func (s *JNIStubs) GetJavaVM() uint64 {
	return s.javaVMBase
}

// JNI Stub Implementations

// stubGetVersion returns JNI_VERSION_1_6
func (s *JNIStubs) stubGetVersion() string {
	s.emu.SetX(0, JNI_VERSION_1_6)
	s.returnFromStub()
	return "-> 0x10006 (JNI 1.6)"
}

// stubFindClass returns a fake class reference
func (s *JNIStubs) stubFindClass() string {
	// X0 = JNIEnv*, X1 = const char* name
	namePtr := s.emu.X(1)
	className, _ := s.emu.MemReadString(namePtr, 256)

	// Get or create fake class reference
	ref, ok := s.classRefs[className]
	if !ok {
		ref = s.mockObjBase + s.nextClassRef
		s.classRefs[className] = ref
		s.nextClassRef += 8
	}

	s.emu.SetX(0, ref)
	s.returnFromStub()
	return "class=\"" + className + "\" -> " + formatHex(ref)
}

// stubGetMethodID returns a fake method ID
func (s *JNIStubs) stubGetMethodID() string {
	// X0 = JNIEnv*, X1 = jclass, X2 = const char* name, X3 = const char* sig
	namePtr := s.emu.X(2)
	sigPtr := s.emu.X(3)
	methodName, _ := s.emu.MemReadString(namePtr, 256)
	methodSig, _ := s.emu.MemReadString(sigPtr, 256)

	// Return fake method ID
	methodID := s.mockObjBase + 0x10000 + uint64(len(methodName))
	s.emu.SetX(0, methodID)
	s.returnFromStub()
	return "method=\"" + methodName + "\" sig=\"" + methodSig + "\""
}

// stubGetStaticMethodID returns a fake static method ID
func (s *JNIStubs) stubGetStaticMethodID() string {
	// X0 = JNIEnv*, X1 = jclass, X2 = const char* name, X3 = const char* sig
	namePtr := s.emu.X(2)
	sigPtr := s.emu.X(3)
	methodName, _ := s.emu.MemReadString(namePtr, 256)
	methodSig, _ := s.emu.MemReadString(sigPtr, 256)

	methodID := s.mockObjBase + 0x20000 + uint64(len(methodName))
	s.emu.SetX(0, methodID)
	s.returnFromStub()
	return "static method=\"" + methodName + "\" sig=\"" + methodSig + "\""
}

// stubGetObjectClass returns a fake class for an object
func (s *JNIStubs) stubGetObjectClass() string {
	// X0 = JNIEnv*, X1 = jobject
	obj := s.emu.X(1)

	// Return a fake class reference
	classRef := s.mockObjBase + 0x30000
	s.emu.SetX(0, classRef)
	s.returnFromStub()
	return "obj=" + formatHex(obj) + " -> class=" + formatHex(classRef)
}

// stubNewStringUTF creates a fake jstring from a UTF-8 string
func (s *JNIStubs) stubNewStringUTF() string {
	// X0 = JNIEnv*, X1 = const char* utf
	utfPtr := s.emu.X(1)
	str, _ := s.emu.MemReadString(utfPtr, 4096)

	// Create fake jstring reference and track the string
	ref := s.mockObjBase + s.nextStringRef
	s.jniStrings[ref] = str
	s.nextStringRef += 8

	s.emu.SetX(0, ref)
	s.returnFromStub()

	displayStr := str
	if len(displayStr) > 30 {
		displayStr = displayStr[:30] + "..."
	}
	return "\"" + displayStr + "\" -> " + formatHex(ref)
}

// stubGetStringUTFChars returns a pointer to the UTF-8 string data
func (s *JNIStubs) stubGetStringUTFChars() string {
	// X0 = JNIEnv*, X1 = jstring, X2 = jboolean* isCopy
	jstr := s.emu.X(1)
	isCopyPtr := s.emu.X(2)

	// Look up the string
	str, ok := s.jniStrings[jstr]
	if !ok {
		str = ""
	}

	// Allocate memory for the string
	bufSize := uint64(len(str) + 1)
	buf := s.emu.Malloc(bufSize)
	s.emu.MemWriteString(buf, str)

	// Set isCopy to true if pointer is valid
	if isCopyPtr != 0 {
		s.emu.MemWrite(isCopyPtr, []byte{1})
	}

	s.emu.SetX(0, buf)
	s.returnFromStub()
	return "jstr=" + formatHex(jstr) + " -> " + formatHex(buf)
}

// stubReleaseStringUTFChars does nothing (we don't free)
func (s *JNIStubs) stubReleaseStringUTFChars() string {
	// X0 = JNIEnv*, X1 = jstring, X2 = const char* utf
	s.returnFromStub()
	return "(no-op)"
}

// stubGetJavaVM writes JavaVM* to output parameter
func (s *JNIStubs) stubGetJavaVM() string {
	// X0 = JNIEnv*, X1 = JavaVM**
	vmPtr := s.emu.X(1)

	// Write JavaVM pointer to output
	s.emu.MemWriteU64(vmPtr, s.javaVMBase)

	s.emu.SetX(0, JNI_OK)
	s.returnFromStub()
	return "vm_ptr=" + formatHex(vmPtr) + " -> " + formatHex(s.javaVMBase)
}

// stubCallObjectMethod returns a mock object
func (s *JNIStubs) stubCallObjectMethod() string {
	// X0 = JNIEnv*, X1 = jobject, X2 = jmethodID, ...args
	obj := s.emu.X(1)
	methodID := s.emu.X(2)

	// Return mock object
	result := s.mockObjBase + 0x40000
	s.emu.SetX(0, result)
	s.returnFromStub()
	return "obj=" + formatHex(obj) + " method=" + formatHex(methodID) + " -> " + formatHex(result)
}

// stubJNIGeneric handles unimplemented JNI functions
func (s *JNIStubs) stubJNIGeneric() string {
	// Return mock object for most JNI functions
	s.emu.SetX(0, s.mockObjBase)
	s.returnFromStub()
	return "-> " + formatHex(s.mockObjBase)
}

// JavaVM Stub Implementations

// stubJavaVMGetEnv implements GetEnv(JavaVM*, void** penv, jint version)
func (s *JNIStubs) stubJavaVMGetEnv() string {
	// X0 = JavaVM*, X1 = void** penv, X2 = jint version
	penvPtr := s.emu.X(1)
	version := s.emu.X(2)

	// Write JNIEnv* to output parameter
	s.emu.MemWriteU64(penvPtr, s.jniEnvBase)

	s.emu.SetX(0, JNI_OK)
	s.returnFromStub()
	return "version=" + formatHex(version) + " *penv=" + formatHex(s.jniEnvBase)
}

// stubAttachCurrentThread attaches current thread and returns JNIEnv*
func (s *JNIStubs) stubAttachCurrentThread() string {
	// X0 = JavaVM*, X1 = JNIEnv**, X2 = void* args
	penvPtr := s.emu.X(1)

	// Write JNIEnv* to output parameter
	s.emu.MemWriteU64(penvPtr, s.jniEnvBase)

	s.emu.SetX(0, JNI_OK)
	s.returnFromStub()
	return "*penv=" + formatHex(s.jniEnvBase)
}

// stubDetachCurrentThread does nothing
func (s *JNIStubs) stubDetachCurrentThread() string {
	s.emu.SetX(0, JNI_OK)
	s.returnFromStub()
	return "(no-op)"
}

// stubJavaVMGeneric handles unimplemented JavaVM functions
func (s *JNIStubs) stubJavaVMGeneric() string {
	s.emu.SetX(0, JNI_OK)
	s.returnFromStub()
	return "-> JNI_OK"
}

// GetTrackedString returns a JNI string by its reference.
func (s *JNIStubs) GetTrackedString(ref uint64) (string, bool) {
	str, ok := s.jniStrings[ref]
	return str, ok
}

// GetAllTrackedStrings returns all JNI strings.
func (s *JNIStubs) GetAllTrackedStrings() map[uint64]string {
	result := make(map[uint64]string)
	for k, v := range s.jniStrings {
		result[k] = v
	}
	return result
}
