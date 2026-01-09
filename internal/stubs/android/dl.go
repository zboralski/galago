package android

import (
	"sync"

	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

var (
	dlHandles   = make(map[uint64]string) // handle -> library name
	nextHandle  uint64 = 0x7F000000
	dlLastError string
	dlMu        sync.Mutex
)

func init() {
	stubs.RegisterFunc("android", "dlopen", stubDlopen)
	stubs.RegisterFunc("android", "dlsym", stubDlsym)
	stubs.RegisterFunc("android", "dlclose", stubDlclose)
	stubs.RegisterFunc("android", "dlerror", stubDlerror)
	stubs.RegisterFunc("android", "dladdr", stubDladdr)

	// Android-specific
	stubs.RegisterFunc("android", "android_dlopen_ext", stubAndroidDlopenExt)
	stubs.RegisterFunc("android", "dl_iterate_phdr", stubDlIteratePhdr)
}

func stubDlopen(emu *emulator.Emulator) bool {
	filenamePtr := emu.X(0)
	// flags := emu.X(1)

	filename := ""
	if filenamePtr != 0 {
		filename, _ = emu.MemReadString(filenamePtr, 256)
	}

	dlMu.Lock()
	handle := nextHandle
	nextHandle += 0x1000
	dlHandles[handle] = filename
	dlLastError = ""
	dlMu.Unlock()

	stubs.DefaultRegistry.Log("android", "dlopen", filename+" -> "+stubs.FormatHex(handle))

	emu.SetX(0, handle)
	stubs.ReturnFromStub(emu)
	return false
}

func stubAndroidDlopenExt(emu *emulator.Emulator) bool {
	// Same as dlopen but with extinfo parameter
	return stubDlopen(emu)
}

func stubDlsym(emu *emulator.Emulator) bool {
	handle := emu.X(0)
	symbolPtr := emu.X(1)

	symbol, _ := emu.MemReadString(symbolPtr, 128)

	dlMu.Lock()
	lib, ok := dlHandles[handle]
	dlMu.Unlock()

	if !ok && handle != 0 {
		// Unknown handle
		dlMu.Lock()
		dlLastError = "invalid handle"
		dlMu.Unlock()
		emu.SetX(0, 0)
		stubs.ReturnFromStub(emu)
		return false
	}

	// Return a fake symbol address in the stub region
	fakeAddr := uint64(0xDEAE0000) + uint64(len(symbol))*8

	stubs.DefaultRegistry.Log("android", "dlsym", lib+":"+symbol+" -> "+stubs.FormatHex(fakeAddr))

	emu.SetX(0, fakeAddr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubDlclose(emu *emulator.Emulator) bool {
	handle := emu.X(0)

	dlMu.Lock()
	delete(dlHandles, handle)
	dlMu.Unlock()

	emu.SetX(0, 0) // Success
	stubs.ReturnFromStub(emu)
	return false
}

func stubDlerror(emu *emulator.Emulator) bool {
	dlMu.Lock()
	err := dlLastError
	dlLastError = ""
	dlMu.Unlock()

	if err == "" {
		emu.SetX(0, 0)
	} else {
		// Allocate error string
		ptr := emu.Malloc(uint64(len(err) + 1))
		emu.MemWriteString(ptr, err)
		emu.SetX(0, ptr)
	}

	stubs.ReturnFromStub(emu)
	return false
}

func stubDladdr(emu *emulator.Emulator) bool {
	// int dladdr(const void *addr, Dl_info *info)
	// We don't track symbol addresses, so just return 0 (not found)
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubDlIteratePhdr(emu *emulator.Emulator) bool {
	// int dl_iterate_phdr(int (*callback)(struct dl_phdr_info *, size_t, void *), void *data)
	// Return 0 without calling the callback - we don't expose ELF headers
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}
