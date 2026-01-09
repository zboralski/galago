package pthread

import (
	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

func init() {
	stubs.RegisterFunc("pthread", "pthread_cond_init", stubCondInit)
	stubs.RegisterFunc("pthread", "pthread_cond_destroy", stubCondDestroy)
	stubs.RegisterFunc("pthread", "pthread_cond_wait", stubCondWait)
	stubs.RegisterFunc("pthread", "pthread_cond_timedwait", stubCondTimedwait)
	stubs.RegisterFunc("pthread", "pthread_cond_signal", stubCondSignal)
	stubs.RegisterFunc("pthread", "pthread_cond_broadcast", stubCondBroadcast)
}

func stubCondInit(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubCondDestroy(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubCondWait(emu *emulator.Emulator) bool {
	// In single-threaded emulation, waiting on a condition variable
	// would deadlock. Just return success immediately.
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubCondTimedwait(emu *emulator.Emulator) bool {
	// Return success (signal received)
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubCondSignal(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubCondBroadcast(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}
