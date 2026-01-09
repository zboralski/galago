package pthread

import (
	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

func init() {
	stubs.RegisterFunc("pthread", "pthread_attr_init", stubAttrInit)
	stubs.RegisterFunc("pthread", "pthread_attr_destroy", stubAttrDestroy)
	stubs.RegisterFunc("pthread", "pthread_attr_setstacksize", stubAttrSetstacksize)
	stubs.RegisterFunc("pthread", "pthread_attr_getstacksize", stubAttrGetstacksize)
	stubs.RegisterFunc("pthread", "pthread_attr_setdetachstate", stubAttrSetdetachstate)
	stubs.RegisterFunc("pthread", "pthread_attr_getdetachstate", stubAttrGetdetachstate)
	stubs.RegisterFunc("pthread", "pthread_attr_setschedparam", stubAttrSetschedparam)
	stubs.RegisterFunc("pthread", "pthread_attr_getschedparam", stubAttrGetschedparam)
	stubs.RegisterFunc("pthread", "pthread_mutexattr_init", stubMutexattrInit)
	stubs.RegisterFunc("pthread", "pthread_mutexattr_destroy", stubMutexattrDestroy)
	stubs.RegisterFunc("pthread", "pthread_mutexattr_settype", stubMutexattrSettype)
	stubs.RegisterFunc("pthread", "pthread_condattr_init", stubCondattrInit)
	stubs.RegisterFunc("pthread", "pthread_condattr_destroy", stubCondattrDestroy)
}

func stubAttrInit(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubAttrDestroy(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubAttrSetstacksize(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubAttrGetstacksize(emu *emulator.Emulator) bool {
	// attr := emu.X(0)
	sizePtr := emu.X(1)
	if sizePtr != 0 {
		emu.MemWriteU64(sizePtr, 8*1024*1024) // 8MB default
	}
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubAttrSetdetachstate(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubAttrGetdetachstate(emu *emulator.Emulator) bool {
	// attr := emu.X(0)
	statePtr := emu.X(1)
	if statePtr != 0 {
		emu.MemWriteU32(statePtr, 0) // PTHREAD_CREATE_JOINABLE
	}
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubAttrSetschedparam(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubAttrGetschedparam(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMutexattrInit(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMutexattrDestroy(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMutexattrSettype(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubCondattrInit(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubCondattrDestroy(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}
