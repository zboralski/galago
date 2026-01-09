package libc

import (
	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

func init() {
	stubs.RegisterFunc("libc", "abort", stubAbort)
	stubs.RegisterFunc("libc", "exit", stubExit)
	stubs.RegisterFunc("libc", "_exit", stubExit)
	stubs.RegisterFunc("libc", "_Exit", stubExit)
	stubs.RegisterFunc("libc", "atexit", stubAtexit)
	// __cxa_atexit is registered in cxxabi package
}

func stubAbort(emu *emulator.Emulator) bool {
	stubs.DefaultRegistry.Log("libc", "abort", "program aborted")
	// Stop emulation - abort() should terminate
	return true
}

func stubExit(emu *emulator.Emulator) bool {
	code := emu.X(0)
	stubs.DefaultRegistry.Log("libc", "exit", stubs.FormatHex(code))
	// Stop emulation
	return true
}

func stubAtexit(emu *emulator.Emulator) bool {
	// int atexit(void (*function)(void))
	// We don't actually register handlers, just return success
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}
