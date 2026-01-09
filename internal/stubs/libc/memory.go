// Package libc provides stub implementations for libc memory functions.
package libc

import (
	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

func init() {
	stubs.Register(stubs.StubDef{Name: "malloc", Hook: stubMalloc, Category: "libc"})
	stubs.Register(stubs.StubDef{Name: "calloc", Hook: stubCalloc, Category: "libc"})
	stubs.Register(stubs.StubDef{Name: "realloc", Hook: stubRealloc, Category: "libc"})
	stubs.Register(stubs.StubDef{Name: "free", Hook: stubFree, Category: "libc"})

	// Memory info
	stubs.Register(stubs.StubDef{Name: "getpagesize", Hook: stubGetPageSize, Category: "libc"})

	// C++ operator new/delete
	stubs.Register(stubs.StubDef{
		Name:     "_Znwm",
		Aliases:  []string{"_Znam", "_ZnwmSt11align_val_t", "_ZnamSt11align_val_t"},
		Hook:     stubNew,
		Category: "libc",
	})
	stubs.Register(stubs.StubDef{
		Name:     "_ZdlPv",
		Aliases:  []string{"_ZdaPv", "_ZdlPvm", "_ZdaPvm"},
		Hook:     stubDelete,
		Category: "libc",
	})
}

func stubMalloc(emu *emulator.Emulator) bool {
	size := emu.X(0)
	if size == 0 {
		size = 16
	}
	size = (size + 15) & ^uint64(15) // Align to 16 bytes

	ptr := emu.Malloc(size)

	// Zero-initialize
	zeros := make([]byte, min(size, 4096))
	emu.MemWrite(ptr, zeros)

	stubs.DefaultRegistry.Log("libc", "malloc", stubs.FormatPtrPair("size", size, "->", ptr))
	emu.SetX(0, ptr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubCalloc(emu *emulator.Emulator) bool {
	count := emu.X(0)
	size := emu.X(1)
	total := count * size
	if total == 0 {
		total = 16
	}
	total = (total + 15) & ^uint64(15)

	ptr := emu.Malloc(total)

	zeros := make([]byte, min(total, 4096))
	emu.MemWrite(ptr, zeros)

	stubs.DefaultRegistry.Log("libc", "calloc", stubs.FormatPtrPair("total", total, "->", ptr))
	emu.SetX(0, ptr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubRealloc(emu *emulator.Emulator) bool {
	_ = emu.X(0) // old ptr (ignored - we leak)
	size := emu.X(1)
	if size == 0 {
		size = 16
	}
	size = (size + 15) & ^uint64(15)

	ptr := emu.Malloc(size)

	stubs.DefaultRegistry.Log("libc", "realloc", stubs.FormatPtrPair("size", size, "->", ptr))
	emu.SetX(0, ptr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFree(emu *emulator.Emulator) bool {
	stubs.DefaultRegistry.Log("libc", "free", "")
	stubs.ReturnFromStub(emu)
	return false
}

func stubNew(emu *emulator.Emulator) bool {
	size := emu.X(0)
	if size == 0 {
		size = 16
	}
	size = (size + 15) & ^uint64(15)

	ptr := emu.Malloc(size)

	zeros := make([]byte, min(size, 4096))
	emu.MemWrite(ptr, zeros)

	stubs.DefaultRegistry.Log("libc", "new", stubs.FormatPtrPair("size", size, "->", ptr))
	emu.SetX(0, ptr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubDelete(emu *emulator.Emulator) bool {
	stubs.DefaultRegistry.Log("libc", "delete", "")
	stubs.ReturnFromStub(emu)
	return false
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func stubGetPageSize(emu *emulator.Emulator) bool {
	stubs.DefaultRegistry.Log("libc", "getpagesize", "-> 4096")
	emu.SetX(0, 4096)
	stubs.ReturnFromStub(emu)
	return false
}
