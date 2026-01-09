// Package pthread provides stub implementations for pthread functions.
package pthread

import (
	"sync"

	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

var (
	nextThreadID uint64 = 1
	threadMu     sync.Mutex
)

func init() {
	stubs.RegisterFunc("pthread", "pthread_create", stubPthreadCreate)
	stubs.RegisterFunc("pthread", "pthread_join", stubPthreadJoin)
	stubs.RegisterFunc("pthread", "pthread_detach", stubPthreadDetach)
	stubs.RegisterFunc("pthread", "pthread_equal", stubPthreadEqual)
	stubs.RegisterFunc("pthread", "pthread_self", stubPthreadSelf)
	stubs.RegisterFunc("pthread", "pthread_setname_np", stubPthreadSetnamNp)
	stubs.RegisterFunc("pthread", "pthread_getname_np", stubPthreadGetnamNp)
	stubs.RegisterFunc("pthread", "pthread_exit", stubPthreadExit)
	stubs.RegisterFunc("pthread", "pthread_cancel", stubPthreadCancel)
	stubs.RegisterFunc("pthread", "sched_yield", stubSchedYield)
}

func stubPthreadCreate(emu *emulator.Emulator) bool {
	threadPtr := emu.X(0)
	// attr := emu.X(1)  // ignored
	// startRoutine := emu.X(2)  // ignored - we don't actually spawn threads
	// arg := emu.X(3)

	// Generate fake thread ID
	threadMu.Lock()
	tid := nextThreadID
	nextThreadID++
	threadMu.Unlock()

	// Write thread ID to output pointer
	if threadPtr != 0 {
		emu.MemWriteU64(threadPtr, tid)
	}

	stubs.DefaultRegistry.Log("pthread", "pthread_create", stubs.FormatPtrPair("tid", tid, "->", threadPtr))
	emu.SetX(0, 0) // Success
	stubs.ReturnFromStub(emu)
	return false
}

func stubPthreadJoin(emu *emulator.Emulator) bool {
	// tid := emu.X(0)
	retvalPtr := emu.X(1)

	// Write NULL to retval
	if retvalPtr != 0 {
		emu.MemWriteU64(retvalPtr, 0)
	}

	emu.SetX(0, 0) // Success
	stubs.ReturnFromStub(emu)
	return false
}

func stubPthreadDetach(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubPthreadEqual(emu *emulator.Emulator) bool {
	t1 := emu.X(0)
	t2 := emu.X(1)

	if t1 == t2 {
		emu.SetX(0, 1) // Equal
	} else {
		emu.SetX(0, 0) // Not equal
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubPthreadSelf(emu *emulator.Emulator) bool {
	// Return thread ID 1 (main thread)
	emu.SetX(0, 1)
	stubs.ReturnFromStub(emu)
	return false
}

func stubPthreadSetnamNp(emu *emulator.Emulator) bool {
	// Just ignore the name
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubPthreadGetnamNp(emu *emulator.Emulator) bool {
	// tid := emu.X(0)
	buf := emu.X(1)
	// bufLen := emu.X(2)

	if buf != 0 {
		emu.MemWriteString(buf, "main")
	}
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubPthreadExit(emu *emulator.Emulator) bool {
	// Don't actually exit - just return
	stubs.ReturnFromStub(emu)
	return false
}

func stubPthreadCancel(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubSchedYield(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}
