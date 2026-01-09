package pthread

import (
	"sync"

	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

var (
	tlsData    = make(map[uint64]uint64) // key -> value
	nextTLSKey uint64
	tlsMu      sync.Mutex
)

func init() {
	stubs.RegisterFunc("pthread", "pthread_key_create", stubKeyCreate)
	stubs.RegisterFunc("pthread", "pthread_key_delete", stubKeyDelete)
	stubs.RegisterFunc("pthread", "pthread_setspecific", stubSetspecific)
	stubs.RegisterFunc("pthread", "pthread_getspecific", stubGetspecific)
	stubs.RegisterFunc("pthread", "pthread_once", stubOnce)
}

func stubKeyCreate(emu *emulator.Emulator) bool {
	keyPtr := emu.X(0)
	// destructor := emu.X(1) // ignored

	tlsMu.Lock()
	key := nextTLSKey
	nextTLSKey++
	tlsMu.Unlock()

	if keyPtr != 0 {
		emu.MemWriteU64(keyPtr, key)
	}

	emu.SetX(0, 0) // Success
	stubs.ReturnFromStub(emu)
	return false
}

func stubKeyDelete(emu *emulator.Emulator) bool {
	key := emu.X(0)

	tlsMu.Lock()
	delete(tlsData, key)
	tlsMu.Unlock()

	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubSetspecific(emu *emulator.Emulator) bool {
	key := emu.X(0)
	value := emu.X(1)

	tlsMu.Lock()
	tlsData[key] = value
	tlsMu.Unlock()

	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubGetspecific(emu *emulator.Emulator) bool {
	key := emu.X(0)

	tlsMu.Lock()
	value := tlsData[key]
	tlsMu.Unlock()

	emu.SetX(0, value)
	stubs.ReturnFromStub(emu)
	return false
}

var onceFlags = make(map[uint64]bool)

func stubOnce(emu *emulator.Emulator) bool {
	onceControl := emu.X(0)
	initRoutine := emu.X(1)

	tlsMu.Lock()
	alreadyCalled := onceFlags[onceControl]
	if !alreadyCalled {
		onceFlags[onceControl] = true
	}
	tlsMu.Unlock()

	if !alreadyCalled && initRoutine != 0 {
		// We should call the init routine, but for emulation
		// we just skip it and hope it's not critical
		stubs.DefaultRegistry.Log("pthread", "pthread_once", stubs.FormatPtr("init_routine", initRoutine)+" (skipped)")
	}

	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}
