package pthread

import (
	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

func init() {
	stubs.RegisterFunc("pthread", "pthread_mutex_init", stubMutexInit)
	stubs.RegisterFunc("pthread", "pthread_mutex_destroy", stubMutexDestroy)
	stubs.RegisterFunc("pthread", "pthread_mutex_lock", stubMutexLock)
	stubs.RegisterFunc("pthread", "pthread_mutex_trylock", stubMutexTrylock)
	stubs.RegisterFunc("pthread", "pthread_mutex_unlock", stubMutexUnlock)

	// Rwlock
	stubs.RegisterFunc("pthread", "pthread_rwlock_init", stubRwlockInit)
	stubs.RegisterFunc("pthread", "pthread_rwlock_destroy", stubRwlockDestroy)
	stubs.RegisterFunc("pthread", "pthread_rwlock_rdlock", stubRwlockRdlock)
	stubs.RegisterFunc("pthread", "pthread_rwlock_wrlock", stubRwlockWrlock)
	stubs.RegisterFunc("pthread", "pthread_rwlock_unlock", stubRwlockUnlock)

	// Spinlock
	stubs.RegisterFunc("pthread", "pthread_spin_init", stubSpinInit)
	stubs.RegisterFunc("pthread", "pthread_spin_destroy", stubSpinDestroy)
	stubs.RegisterFunc("pthread", "pthread_spin_lock", stubSpinLock)
	stubs.RegisterFunc("pthread", "pthread_spin_unlock", stubSpinUnlock)
}

func stubMutexInit(emu *emulator.Emulator) bool {
	// mutex := emu.X(0)
	// attr := emu.X(1)
	emu.SetX(0, 0) // Success
	stubs.ReturnFromStub(emu)
	return false
}

func stubMutexDestroy(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMutexLock(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMutexTrylock(emu *emulator.Emulator) bool {
	emu.SetX(0, 0) // Always succeed
	stubs.ReturnFromStub(emu)
	return false
}

func stubMutexUnlock(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubRwlockInit(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubRwlockDestroy(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubRwlockRdlock(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubRwlockWrlock(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubRwlockUnlock(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubSpinInit(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubSpinDestroy(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubSpinLock(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubSpinUnlock(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}
