// Package jni provides JNI/JavaVM mock implementations.
// This package provides fake JNIEnv and JavaVM structures for emulating Android native libraries.
//
// The JNI subsystem is automatically activated when JNI-related imports are detected.
package jni

import (
	"sync"

	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

// currentEnv holds the active JNI environment (singleton per emulation session)
var (
	currentEnv   *Env
	currentEnvMu sync.Mutex
)

func init() {
	// Register JNI as a detector - activates when JNI symbols are found
	stubs.RegisterDetector(stubs.Detector{
		Name: "jni",
		Patterns: []string{
			"JNI_OnLoad",
			"_JNIEnv",
			"JavaVM",
			"GetEnv",
			"AttachCurrentThread",
		},
		Activate:    activateJNI,
		Description: "JNI/JavaVM mock implementation",
	})
}

// activateJNI sets up JNI vtables when JNI symbols are detected.
func activateJNI(emu *emulator.Emulator, imports, symbols map[string]uint64) int {
	currentEnvMu.Lock()
	defer currentEnvMu.Unlock()

	// Create and install JNI environment
	currentEnv = NewEnv(emu)
	currentEnv.Install()

	stubs.DefaultRegistry.Log("jni", "activate", "JNI vtables installed")
	return 1 // Consider the JNI setup as 1 "installed" item
}

// GetCurrentEnv returns the active JNI environment, or nil if not activated.
func GetCurrentEnv() *Env {
	currentEnvMu.Lock()
	defer currentEnvMu.Unlock()
	return currentEnv
}

// GetJNIEnv returns the JNIEnv* pointer for use in function calls.
// Returns 0 if JNI is not activated.
func GetJNIEnv() uint64 {
	env := GetCurrentEnv()
	if env == nil {
		return 0
	}
	return env.GetJNIEnv()
}

// GetJavaVM returns the JavaVM* pointer for use in JNI_OnLoad calls.
// Returns 0 if JNI is not activated.
func GetJavaVM() uint64 {
	env := GetCurrentEnv()
	if env == nil {
		return 0
	}
	return env.GetJavaVM()
}

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
	JNI_GetVersion            = 4
	JNI_DefineClass           = 5
	JNI_FindClass             = 6
	JNI_FromReflectedMethod   = 7
	JNI_FromReflectedField    = 8
	JNI_ToReflectedMethod     = 9
	JNI_GetSuperclass         = 10
	JNI_IsAssignableFrom      = 11
	JNI_ToReflectedField      = 12
	JNI_Throw                 = 13
	JNI_ThrowNew              = 14
	JNI_ExceptionOccurred     = 15
	JNI_ExceptionDescribe     = 16
	JNI_ExceptionClear        = 17
	JNI_FatalError            = 18
	JNI_PushLocalFrame        = 19
	JNI_PopLocalFrame         = 20
	JNI_NewGlobalRef          = 21
	JNI_DeleteGlobalRef       = 22
	JNI_DeleteLocalRef        = 23
	JNI_IsSameObject          = 24
	JNI_NewLocalRef           = 25
	JNI_EnsureLocalCapacity   = 26
	JNI_AllocObject           = 27
	JNI_NewObject             = 28
	JNI_NewObjectV            = 29
	JNI_NewObjectA            = 30
	JNI_GetObjectClass        = 31
	JNI_IsInstanceOf          = 32
	JNI_GetMethodID           = 33
	JNI_CallObjectMethod      = 34
	JNI_CallBooleanMethod     = 37
	JNI_CallByteMethod        = 40
	JNI_CallCharMethod        = 43
	JNI_CallShortMethod       = 46
	JNI_CallIntMethod         = 49
	JNI_CallLongMethod        = 52
	JNI_CallFloatMethod       = 55
	JNI_CallDoubleMethod      = 58
	JNI_CallVoidMethod        = 61
	JNI_GetFieldID            = 94
	JNI_GetObjectField        = 95
	JNI_GetBooleanField       = 96
	JNI_GetByteField          = 97
	JNI_GetCharField          = 98
	JNI_GetShortField         = 99
	JNI_GetIntField           = 100
	JNI_GetLongField          = 101
	JNI_GetFloatField         = 102
	JNI_GetDoubleField        = 103
	JNI_SetObjectField        = 104
	JNI_SetBooleanField       = 105
	JNI_SetByteField          = 106
	JNI_SetCharField          = 107
	JNI_SetShortField         = 108
	JNI_SetIntField           = 109
	JNI_SetLongField          = 110
	JNI_SetFloatField         = 111
	JNI_SetDoubleField        = 112
	JNI_GetStaticMethodID     = 113
	JNI_CallStaticObjectMethod = 114
	JNI_CallStaticBooleanMethod = 117
	JNI_CallStaticByteMethod  = 120
	JNI_CallStaticCharMethod  = 123
	JNI_CallStaticShortMethod = 126
	JNI_CallStaticIntMethod   = 129
	JNI_CallStaticLongMethod  = 132
	JNI_CallStaticFloatMethod = 135
	JNI_CallStaticDoubleMethod = 138
	JNI_CallStaticVoidMethod  = 141
	JNI_GetStaticFieldID      = 144
	JNI_GetStaticObjectField  = 145
	JNI_GetStaticBooleanField = 146
	JNI_GetStaticByteField    = 147
	JNI_GetStaticCharField    = 148
	JNI_GetStaticShortField   = 149
	JNI_GetStaticIntField     = 150
	JNI_GetStaticLongField    = 151
	JNI_GetStaticFloatField   = 152
	JNI_GetStaticDoubleField  = 153
	JNI_SetStaticObjectField  = 154
	JNI_SetStaticBooleanField = 155
	JNI_SetStaticByteField    = 156
	JNI_SetStaticCharField    = 157
	JNI_SetStaticShortField   = 158
	JNI_SetStaticIntField     = 159
	JNI_SetStaticLongField    = 160
	JNI_SetStaticFloatField   = 161
	JNI_SetStaticDoubleField  = 162
	JNI_NewString             = 163
	JNI_GetStringLength       = 164
	JNI_GetStringChars        = 165
	JNI_ReleaseStringChars    = 166
	JNI_NewStringUTF          = 167
	JNI_GetStringUTFLength    = 168
	JNI_GetStringUTFChars     = 169
	JNI_ReleaseStringUTFChars = 170
	JNI_GetArrayLength        = 171
	JNI_NewObjectArray        = 172
	JNI_GetObjectArrayElement = 173
	JNI_SetObjectArrayElement = 174
	JNI_NewBooleanArray       = 175
	JNI_NewByteArray          = 176
	JNI_NewCharArray          = 177
	JNI_NewShortArray         = 178
	JNI_NewIntArray           = 179
	JNI_NewLongArray          = 180
	JNI_NewFloatArray         = 181
	JNI_NewDoubleArray        = 182
	JNI_GetBooleanArrayElements = 183
	JNI_GetByteArrayElements  = 184
	JNI_GetCharArrayElements  = 185
	JNI_GetShortArrayElements = 186
	JNI_GetIntArrayElements   = 187
	JNI_GetLongArrayElements  = 188
	JNI_GetFloatArrayElements = 189
	JNI_GetDoubleArrayElements = 190
	JNI_ReleaseBooleanArrayElements = 191
	JNI_ReleaseByteArrayElements = 192
	JNI_ReleaseCharArrayElements = 193
	JNI_ReleaseShortArrayElements = 194
	JNI_ReleaseIntArrayElements = 195
	JNI_ReleaseLongArrayElements = 196
	JNI_ReleaseFloatArrayElements = 197
	JNI_ReleaseDoubleArrayElements = 198
	JNI_GetBooleanArrayRegion = 199
	JNI_GetByteArrayRegion    = 200
	JNI_GetCharArrayRegion    = 201
	JNI_GetShortArrayRegion   = 202
	JNI_GetIntArrayRegion     = 203
	JNI_GetLongArrayRegion    = 204
	JNI_GetFloatArrayRegion   = 205
	JNI_GetDoubleArrayRegion  = 206
	JNI_SetBooleanArrayRegion = 207
	JNI_SetByteArrayRegion    = 208
	JNI_SetCharArrayRegion    = 209
	JNI_SetShortArrayRegion   = 210
	JNI_SetIntArrayRegion     = 211
	JNI_SetLongArrayRegion    = 212
	JNI_SetFloatArrayRegion   = 213
	JNI_SetDoubleArrayRegion  = 214
	JNI_RegisterNatives       = 215
	JNI_UnregisterNatives     = 216
	JNI_MonitorEnter          = 217
	JNI_MonitorExit           = 218
	JNI_GetJavaVM             = 219
	JNI_GetStringRegion       = 220
	JNI_GetStringUTFRegion    = 221
	JNI_GetPrimitiveArrayCritical = 222
	JNI_ReleasePrimitiveArrayCritical = 223
	JNI_GetStringCritical     = 224
	JNI_ReleaseStringCritical = 225
	JNI_NewWeakGlobalRef      = 226
	JNI_DeleteWeakGlobalRef   = 227
	JNI_ExceptionCheck        = 228
	JNI_NewDirectByteBuffer   = 229
	JNI_GetDirectBufferAddress = 230
	JNI_GetDirectBufferCapacity = 231
	JNI_GetObjectRefType      = 232
	JNI_FUNC_COUNT            = 300
)

// JavaVM Function Indices
const (
	JAVAVM_DestroyJavaVM            = 0
	JAVAVM_AttachCurrentThread      = 1
	JAVAVM_DetachCurrentThread      = 2
	JAVAVM_GetEnv                   = 3
	JAVAVM_AttachCurrentThreadAsDaemon = 4
	JAVAVM_FUNC_COUNT               = 10
)

// Env provides JNI/JavaVM stub implementations.
type Env struct {
	emu *emulator.Emulator

	// Memory layout
	jniEnvBase     uint64
	jniVtableBase  uint64
	jniStubBase    uint64
	javaVMBase     uint64
	javaVMVtable   uint64
	javaVMStubBase uint64
	mockObjBase    uint64

	// String tracking for NewStringUTF
	jniStrings    map[uint64]string
	jniStringsMu  sync.RWMutex
	nextStringRef uint64

	// Class references
	classRefs    map[string]uint64
	classRefsMu  sync.RWMutex
	nextClassRef uint64

	// Method references
	methodRefs   map[string]uint64
	methodRefsMu sync.RWMutex
	nextMethodRef uint64

	// Field references
	fieldRefs   map[string]uint64
	fieldRefsMu sync.RWMutex
	nextFieldRef uint64
}

// NewEnv creates a new JNI environment.
func NewEnv(emu *emulator.Emulator) *Env {
	return &Env{
		emu:           emu,
		jniStrings:    make(map[uint64]string),
		classRefs:     make(map[string]uint64),
		methodRefs:    make(map[string]uint64),
		fieldRefs:     make(map[string]uint64),
		nextStringRef: 0x1000,
		nextClassRef:  0x2000,
		nextMethodRef: 0x3000,
		nextFieldRef:  0x4000,
	}
}

// Install sets up the JNI environment in emulator memory.
// Returns JNIEnv* and JavaVM* pointers.
func (e *Env) Install() (jniEnv, javaVM uint64) {
	base := uint64(emulator.StubBase) + 0x10000

	e.jniEnvBase = base
	e.jniVtableBase = base + 0x1000
	e.jniStubBase = base + 0x2000
	e.javaVMBase = base + 0x3000
	e.javaVMVtable = base + 0x4000
	e.javaVMStubBase = base + 0x5000
	e.mockObjBase = base + 0x6000

	retInsn := []byte{0xc0, 0x03, 0x5f, 0xd6} // RET

	// Create JNI function stubs and vtable
	for i := 0; i < JNI_FUNC_COUNT; i++ {
		stubAddr := e.jniStubBase + uint64(i*4)
		e.emu.MemWrite(stubAddr, retInsn)
		e.emu.MemWriteU64(e.jniVtableBase+uint64(i*8), stubAddr)
		e.installJNIHandler(i, stubAddr)
	}

	// Set up JNIEnv structure
	e.emu.MemWriteU64(e.jniEnvBase, e.jniVtableBase)

	// Create JavaVM function stubs and vtable
	for i := 0; i < JAVAVM_FUNC_COUNT; i++ {
		stubAddr := e.javaVMStubBase + uint64(i*4)
		e.emu.MemWrite(stubAddr, retInsn)
		e.emu.MemWriteU64(e.javaVMVtable+uint64(i*8), stubAddr)
		e.installJavaVMHandler(i, stubAddr)
	}

	// Set up JavaVM structure
	e.emu.MemWriteU64(e.javaVMBase, e.javaVMVtable)

	return e.jniEnvBase, e.javaVMBase
}

func (e *Env) installJNIHandler(index int, stubAddr uint64) {
	switch index {
	case JNI_GetVersion:
		e.emu.HookAddress(stubAddr, e.stubGetVersion)
	case JNI_FindClass:
		e.emu.HookAddress(stubAddr, e.stubFindClass)
	case JNI_GetMethodID:
		e.emu.HookAddress(stubAddr, e.stubGetMethodID)
	case JNI_GetStaticMethodID:
		e.emu.HookAddress(stubAddr, e.stubGetStaticMethodID)
	case JNI_GetObjectClass:
		e.emu.HookAddress(stubAddr, e.stubGetObjectClass)
	case JNI_NewStringUTF:
		e.emu.HookAddress(stubAddr, e.stubNewStringUTF)
	case JNI_GetStringUTFChars:
		e.emu.HookAddress(stubAddr, e.stubGetStringUTFChars)
	case JNI_ReleaseStringUTFChars:
		e.emu.HookAddress(stubAddr, e.stubReleaseStringUTFChars)
	case JNI_GetStringUTFLength:
		e.emu.HookAddress(stubAddr, e.stubGetStringUTFLength)
	case JNI_GetJavaVM:
		e.emu.HookAddress(stubAddr, e.stubGetJavaVM)
	case JNI_CallObjectMethod, JNI_CallBooleanMethod, JNI_CallIntMethod, JNI_CallLongMethod:
		e.emu.HookAddress(stubAddr, e.stubCallMethod)
	case JNI_CallStaticObjectMethod, JNI_CallStaticBooleanMethod, JNI_CallStaticIntMethod, JNI_CallStaticLongMethod:
		e.emu.HookAddress(stubAddr, e.stubCallStaticMethod)
	case JNI_CallVoidMethod:
		e.emu.HookAddress(stubAddr, e.stubCallVoidMethod)
	case JNI_CallStaticVoidMethod:
		e.emu.HookAddress(stubAddr, e.stubCallStaticVoidMethod)
	case JNI_GetFieldID:
		e.emu.HookAddress(stubAddr, e.stubGetFieldID)
	case JNI_GetStaticFieldID:
		e.emu.HookAddress(stubAddr, e.stubGetStaticFieldID)
	case JNI_GetObjectField, JNI_GetIntField, JNI_GetLongField, JNI_GetBooleanField:
		e.emu.HookAddress(stubAddr, e.stubGetField)
	case JNI_GetStaticObjectField, JNI_GetStaticIntField, JNI_GetStaticLongField, JNI_GetStaticBooleanField:
		e.emu.HookAddress(stubAddr, e.stubGetStaticField)
	case JNI_SetObjectField, JNI_SetIntField, JNI_SetLongField, JNI_SetBooleanField:
		e.emu.HookAddress(stubAddr, e.stubSetField)
	case JNI_NewGlobalRef, JNI_NewLocalRef, JNI_NewWeakGlobalRef:
		e.emu.HookAddress(stubAddr, e.stubNewRef)
	case JNI_DeleteGlobalRef, JNI_DeleteLocalRef, JNI_DeleteWeakGlobalRef:
		e.emu.HookAddress(stubAddr, e.stubDeleteRef)
	case JNI_ExceptionCheck:
		e.emu.HookAddress(stubAddr, e.stubExceptionCheck)
	case JNI_ExceptionClear:
		e.emu.HookAddress(stubAddr, e.stubExceptionClear)
	case JNI_ExceptionOccurred:
		e.emu.HookAddress(stubAddr, e.stubExceptionOccurred)
	case JNI_PushLocalFrame:
		e.emu.HookAddress(stubAddr, e.stubPushLocalFrame)
	case JNI_PopLocalFrame:
		e.emu.HookAddress(stubAddr, e.stubPopLocalFrame)
	case JNI_EnsureLocalCapacity:
		e.emu.HookAddress(stubAddr, e.stubEnsureLocalCapacity)
	case JNI_NewByteArray:
		e.emu.HookAddress(stubAddr, e.stubNewByteArray)
	case JNI_GetByteArrayElements:
		e.emu.HookAddress(stubAddr, e.stubGetByteArrayElements)
	case JNI_ReleaseByteArrayElements:
		e.emu.HookAddress(stubAddr, e.stubReleaseByteArrayElements)
	case JNI_GetArrayLength:
		e.emu.HookAddress(stubAddr, e.stubGetArrayLength)
	case JNI_RegisterNatives:
		e.emu.HookAddress(stubAddr, e.stubRegisterNatives)
	case JNI_MonitorEnter, JNI_MonitorExit:
		e.emu.HookAddress(stubAddr, e.stubMonitor)
	case JNI_IsSameObject:
		e.emu.HookAddress(stubAddr, e.stubIsSameObject)
	default:
		e.emu.HookAddress(stubAddr, e.stubJNIGeneric)
	}
}

func (e *Env) installJavaVMHandler(index int, stubAddr uint64) {
	switch index {
	case JAVAVM_GetEnv:
		e.emu.HookAddress(stubAddr, e.stubJavaVMGetEnv)
	case JAVAVM_AttachCurrentThread, JAVAVM_AttachCurrentThreadAsDaemon:
		e.emu.HookAddress(stubAddr, e.stubAttachCurrentThread)
	case JAVAVM_DetachCurrentThread:
		e.emu.HookAddress(stubAddr, e.stubDetachCurrentThread)
	default:
		e.emu.HookAddress(stubAddr, e.stubJavaVMGeneric)
	}
}

// GetJNIEnv returns the JNIEnv* pointer.
func (e *Env) GetJNIEnv() uint64 {
	return e.jniEnvBase
}

// GetJavaVM returns the JavaVM* pointer.
func (e *Env) GetJavaVM() uint64 {
	return e.javaVMBase
}

// GetTrackedStrings returns all JNI strings.
func (e *Env) GetTrackedStrings() map[uint64]string {
	e.jniStringsMu.RLock()
	defer e.jniStringsMu.RUnlock()
	result := make(map[uint64]string, len(e.jniStrings))
	for k, v := range e.jniStrings {
		result[k] = v
	}
	return result
}

// JNI Stub Implementations

func (e *Env) stubGetVersion(emu *emulator.Emulator) bool {
	emu.SetX(0, JNI_VERSION_1_6)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubFindClass(emu *emulator.Emulator) bool {
	namePtr := emu.X(1)
	className, _ := emu.MemReadString(namePtr, 256)

	e.classRefsMu.Lock()
	ref, ok := e.classRefs[className]
	if !ok {
		ref = e.mockObjBase + e.nextClassRef
		e.classRefs[className] = ref
		e.nextClassRef += 8
	}
	e.classRefsMu.Unlock()

	stubs.DefaultRegistry.Log("jni", "FindClass", className)
	emu.SetX(0, ref)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubGetMethodID(emu *emulator.Emulator) bool {
	namePtr := emu.X(2)
	sigPtr := emu.X(3)
	methodName, _ := emu.MemReadString(namePtr, 256)
	methodSig, _ := emu.MemReadString(sigPtr, 256)

	key := methodName + methodSig
	e.methodRefsMu.Lock()
	ref, ok := e.methodRefs[key]
	if !ok {
		ref = e.mockObjBase + 0x10000 + e.nextMethodRef
		e.methodRefs[key] = ref
		e.nextMethodRef += 8
	}
	e.methodRefsMu.Unlock()

	stubs.DefaultRegistry.Log("jni", "GetMethodID", methodName+methodSig)
	emu.SetX(0, ref)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubGetStaticMethodID(emu *emulator.Emulator) bool {
	namePtr := emu.X(2)
	sigPtr := emu.X(3)
	methodName, _ := emu.MemReadString(namePtr, 256)
	methodSig, _ := emu.MemReadString(sigPtr, 256)

	key := "static:" + methodName + methodSig
	e.methodRefsMu.Lock()
	ref, ok := e.methodRefs[key]
	if !ok {
		ref = e.mockObjBase + 0x20000 + e.nextMethodRef
		e.methodRefs[key] = ref
		e.nextMethodRef += 8
	}
	e.methodRefsMu.Unlock()

	stubs.DefaultRegistry.Log("jni", "GetStaticMethodID", methodName+methodSig)
	emu.SetX(0, ref)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubGetObjectClass(emu *emulator.Emulator) bool {
	classRef := e.mockObjBase + 0x30000
	emu.SetX(0, classRef)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubNewStringUTF(emu *emulator.Emulator) bool {
	utfPtr := emu.X(1)
	str, _ := emu.MemReadString(utfPtr, 4096)

	e.jniStringsMu.Lock()
	ref := e.mockObjBase + e.nextStringRef
	e.jniStrings[ref] = str
	e.nextStringRef += 8
	e.jniStringsMu.Unlock()

	truncated := str
	if len(truncated) > 40 {
		truncated = truncated[:40] + "..."
	}
	stubs.DefaultRegistry.Log("jni", "NewStringUTF", "\""+truncated+"\"")

	emu.SetX(0, ref)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubGetStringUTFChars(emu *emulator.Emulator) bool {
	jstr := emu.X(1)
	isCopyPtr := emu.X(2)

	e.jniStringsMu.RLock()
	str := e.jniStrings[jstr]
	e.jniStringsMu.RUnlock()

	buf := emu.Malloc(uint64(len(str) + 1))
	emu.MemWriteString(buf, str)

	if isCopyPtr != 0 {
		emu.MemWriteU8(isCopyPtr, 1)
	}

	emu.SetX(0, buf)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubReleaseStringUTFChars(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubGetStringUTFLength(emu *emulator.Emulator) bool {
	jstr := emu.X(1)

	e.jniStringsMu.RLock()
	str := e.jniStrings[jstr]
	e.jniStringsMu.RUnlock()

	emu.SetX(0, uint64(len(str)))
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubGetJavaVM(emu *emulator.Emulator) bool {
	vmPtr := emu.X(1)
	emu.MemWriteU64(vmPtr, e.javaVMBase)
	emu.SetX(0, JNI_OK)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubCallMethod(emu *emulator.Emulator) bool {
	result := e.mockObjBase + 0x40000
	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubCallStaticMethod(emu *emulator.Emulator) bool {
	result := e.mockObjBase + 0x50000
	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubCallVoidMethod(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubCallStaticVoidMethod(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubGetFieldID(emu *emulator.Emulator) bool {
	namePtr := emu.X(2)
	sigPtr := emu.X(3)
	fieldName, _ := emu.MemReadString(namePtr, 256)
	fieldSig, _ := emu.MemReadString(sigPtr, 256)

	key := fieldName + fieldSig
	e.fieldRefsMu.Lock()
	ref, ok := e.fieldRefs[key]
	if !ok {
		ref = e.mockObjBase + 0x60000 + e.nextFieldRef
		e.fieldRefs[key] = ref
		e.nextFieldRef += 8
	}
	e.fieldRefsMu.Unlock()

	stubs.DefaultRegistry.Log("jni", "GetFieldID", fieldName)
	emu.SetX(0, ref)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubGetStaticFieldID(emu *emulator.Emulator) bool {
	namePtr := emu.X(2)
	sigPtr := emu.X(3)
	fieldName, _ := emu.MemReadString(namePtr, 256)
	fieldSig, _ := emu.MemReadString(sigPtr, 256)

	key := "static:" + fieldName + fieldSig
	e.fieldRefsMu.Lock()
	ref, ok := e.fieldRefs[key]
	if !ok {
		ref = e.mockObjBase + 0x70000 + e.nextFieldRef
		e.fieldRefs[key] = ref
		e.nextFieldRef += 8
	}
	e.fieldRefsMu.Unlock()

	stubs.DefaultRegistry.Log("jni", "GetStaticFieldID", fieldName)
	emu.SetX(0, ref)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubGetField(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubGetStaticField(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubSetField(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubNewRef(emu *emulator.Emulator) bool {
	obj := emu.X(1)
	emu.SetX(0, obj)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubDeleteRef(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubExceptionCheck(emu *emulator.Emulator) bool {
	emu.SetX(0, 0) // No exception
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubExceptionClear(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubExceptionOccurred(emu *emulator.Emulator) bool {
	emu.SetX(0, 0) // No exception
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubPushLocalFrame(emu *emulator.Emulator) bool {
	emu.SetX(0, JNI_OK)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubPopLocalFrame(emu *emulator.Emulator) bool {
	result := emu.X(1)
	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubEnsureLocalCapacity(emu *emulator.Emulator) bool {
	emu.SetX(0, JNI_OK)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubNewByteArray(emu *emulator.Emulator) bool {
	length := emu.X(1)
	arr := emu.Malloc(length + 16) // Extra for array header
	emu.MemWriteU64(arr, length)    // Store length at start
	emu.SetX(0, arr)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubGetByteArrayElements(emu *emulator.Emulator) bool {
	arr := emu.X(1)
	isCopyPtr := emu.X(2)

	if isCopyPtr != 0 {
		emu.MemWriteU8(isCopyPtr, 0)
	}

	// Return pointer past length header
	emu.SetX(0, arr+8)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubReleaseByteArrayElements(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubGetArrayLength(emu *emulator.Emulator) bool {
	arr := emu.X(1)
	length, _ := emu.MemReadU64(arr)
	emu.SetX(0, length)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubRegisterNatives(emu *emulator.Emulator) bool {
	stubs.DefaultRegistry.Log("jni", "RegisterNatives", "")
	emu.SetX(0, JNI_OK)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubMonitor(emu *emulator.Emulator) bool {
	emu.SetX(0, JNI_OK)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubIsSameObject(emu *emulator.Emulator) bool {
	obj1 := emu.X(1)
	obj2 := emu.X(2)
	if obj1 == obj2 {
		emu.SetX(0, 1)
	} else {
		emu.SetX(0, 0)
	}
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubJNIGeneric(emu *emulator.Emulator) bool {
	emu.SetX(0, e.mockObjBase)
	stubs.ReturnFromStub(emu)
	return false
}

// JavaVM Stub Implementations

func (e *Env) stubJavaVMGetEnv(emu *emulator.Emulator) bool {
	penvPtr := emu.X(1)
	emu.MemWriteU64(penvPtr, e.jniEnvBase)
	emu.SetX(0, JNI_OK)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubAttachCurrentThread(emu *emulator.Emulator) bool {
	penvPtr := emu.X(1)
	emu.MemWriteU64(penvPtr, e.jniEnvBase)
	emu.SetX(0, JNI_OK)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubDetachCurrentThread(emu *emulator.Emulator) bool {
	emu.SetX(0, JNI_OK)
	stubs.ReturnFromStub(emu)
	return false
}

func (e *Env) stubJavaVMGeneric(emu *emulator.Emulator) bool {
	emu.SetX(0, JNI_OK)
	stubs.ReturnFromStub(emu)
	return false
}
