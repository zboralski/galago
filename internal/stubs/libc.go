// Package stubs provides stub implementations for external functions.
// These stubs mock libc, C++ ABI, JNI, and other external dependencies.
package stubs

import (
	"github.com/zboralski/galago/internal/emulator"
)

// LibcStubs provides stub implementations for common libc functions.
type LibcStubs struct {
	emu *emulator.Emulator

	// Trace callback for logging stub calls
	OnCall func(name string, detail string)
}

// NewLibcStubs creates libc stubs for an emulator.
func NewLibcStubs(emu *emulator.Emulator) *LibcStubs {
	return &LibcStubs{
		emu: emu,
	}
}

// Install registers all libc stub hooks.
// Call this after loading the ELF and resolving imports.
func (s *LibcStubs) Install(imports map[string]uint64) {
	stubMap := map[string]func(){
		"malloc":     s.stubMalloc,
		"calloc":     s.stubCalloc,
		"realloc":    s.stubRealloc,
		"free":       s.stubFree,
		"memcpy":     s.stubMemcpy,
		"memset":     s.stubMemset,
		"memmove":    s.stubMemmove,
		"strlen":     s.stubStrlen,
		"strcmp":     s.stubStrcmp,
		"strncmp":    s.stubStrncmp,
		"strcpy":     s.stubStrcpy,
		"strncpy":    s.stubStrncpy,

		// Operator new/delete
		"_Znwm":   s.stubNew,       // operator new(size_t)
		"_Znam":   s.stubNew,       // operator new[](size_t)
		"_ZdlPv":  s.stubDelete,    // operator delete(void*)
		"_ZdaPv":  s.stubDelete,    // operator delete[](void*)
		"_ZdlPvm": s.stubDelete,    // operator delete(void*, size_t)

		// Time functions (mocked for determinism)
		"gettimeofday": s.stubGettimeofday,
		"clock_gettime": s.stubClockGettime,
		"time":         s.stubTime,
	}

	for name, stub := range stubMap {
		if addr, ok := imports[name]; ok && addr != 0 {
			s.installStub(name, addr, stub)
		}
	}
}

// InstallAt installs a stub at a specific address.
func (s *LibcStubs) InstallAt(name string, addr uint64, stub func()) {
	s.installStub(name, addr, stub)
}

func (s *LibcStubs) installStub(name string, addr uint64, stub func()) {
	s.emu.HookAddress(addr, func(e *emulator.Emulator) bool {
		stub()
		return false // continue execution
	})
}

func (s *LibcStubs) log(name, detail string) {
	if s.OnCall != nil {
		s.OnCall(name, detail)
	}
}

// Return from stub by setting PC to LR
func (s *LibcStubs) returnFromStub() {
	lr := s.emu.LR()
	s.emu.SetPC(lr)
}

// stubMalloc implements malloc(size_t size)
func (s *LibcStubs) stubMalloc() {
	size := s.emu.X(0)
	if size == 0 {
		size = 16
	}
	// Align to 16 bytes
	size = (size + 15) & ^uint64(15)

	ptr := s.emu.Malloc(size)

	// Zero-initialize
	zeros := make([]byte, min(size, 4096))
	s.emu.MemWrite(ptr, zeros)

	s.log("malloc", formatPtr("size", size, "->", ptr))
	s.emu.SetX(0, ptr)
	s.returnFromStub()
}

// stubCalloc implements calloc(count, size)
func (s *LibcStubs) stubCalloc() {
	count := s.emu.X(0)
	size := s.emu.X(1)
	total := count * size
	if total == 0 {
		total = 16
	}
	total = (total + 15) & ^uint64(15)

	ptr := s.emu.Malloc(total)

	// Zero-initialize
	zeros := make([]byte, min(total, 4096))
	s.emu.MemWrite(ptr, zeros)

	s.log("calloc", formatPtr("total", total, "->", ptr))
	s.emu.SetX(0, ptr)
	s.returnFromStub()
}

// stubRealloc implements realloc(ptr, size)
func (s *LibcStubs) stubRealloc() {
	_ = s.emu.X(0) // old ptr (ignored - we leak)
	size := s.emu.X(1)
	if size == 0 {
		size = 16
	}
	size = (size + 15) & ^uint64(15)

	ptr := s.emu.Malloc(size)

	s.log("realloc", formatPtr("size", size, "->", ptr))
	s.emu.SetX(0, ptr)
	s.returnFromStub()
}

// stubFree implements free(ptr) - no-op
func (s *LibcStubs) stubFree() {
	s.log("free", "")
	s.returnFromStub()
}

// stubNew implements operator new(size_t)
func (s *LibcStubs) stubNew() {
	size := s.emu.X(0)
	if size == 0 {
		size = 16
	}
	size = (size + 15) & ^uint64(15)

	ptr := s.emu.Malloc(size)

	zeros := make([]byte, min(size, 4096))
	s.emu.MemWrite(ptr, zeros)

	s.log("new", formatPtr("size", size, "->", ptr))
	s.emu.SetX(0, ptr)
	s.returnFromStub()
}

// stubDelete implements operator delete(void*) - no-op
func (s *LibcStubs) stubDelete() {
	s.log("delete", "")
	s.returnFromStub()
}

// stubMemcpy implements memcpy(dest, src, n)
func (s *LibcStubs) stubMemcpy() {
	dest := s.emu.X(0)
	src := s.emu.X(1)
	n := s.emu.X(2)

	if n > 0 && n < 0x100000 { // 1MB sanity limit
		data, err := s.emu.MemRead(src, n)
		if err == nil {
			s.emu.MemWrite(dest, data)
		}
	}

	s.log("memcpy", formatMemop(dest, src, n))
	s.emu.SetX(0, dest)
	s.returnFromStub()
}

// stubMemset implements memset(dest, c, n)
func (s *LibcStubs) stubMemset() {
	dest := s.emu.X(0)
	c := byte(s.emu.X(1) & 0xFF)
	n := s.emu.X(2)

	if n > 0 && n < 0x100000 {
		data := make([]byte, n)
		for i := range data {
			data[i] = c
		}
		s.emu.MemWrite(dest, data)
	}

	s.log("memset", formatPtr("dest", dest, "c", uint64(c)))
	s.emu.SetX(0, dest)
	s.returnFromStub()
}

// stubMemmove implements memmove(dest, src, n)
func (s *LibcStubs) stubMemmove() {
	dest := s.emu.X(0)
	src := s.emu.X(1)
	n := s.emu.X(2)

	if n > 0 && n < 0x100000 {
		data, err := s.emu.MemRead(src, n)
		if err == nil {
			s.emu.MemWrite(dest, data)
		}
	}

	s.log("memmove", formatMemop(dest, src, n))
	s.emu.SetX(0, dest)
	s.returnFromStub()
}

// stubStrlen implements strlen(s)
func (s *LibcStubs) stubStrlen() {
	addr := s.emu.X(0)
	str, _ := s.emu.MemReadString(addr, 4096)
	length := uint64(len(str))

	s.log("strlen", formatPtr("len", length, "", 0))
	s.emu.SetX(0, length)
	s.returnFromStub()
}

// stubStrcmp implements strcmp(s1, s2)
func (s *LibcStubs) stubStrcmp() {
	s1, _ := s.emu.MemReadString(s.emu.X(0), 256)
	s2, _ := s.emu.MemReadString(s.emu.X(1), 256)

	var result uint64
	if s1 < s2 {
		result = 0xffffffffffffffff // -1
	} else if s1 > s2 {
		result = 1
	} else {
		result = 0
	}

	s.emu.SetX(0, result)
	s.returnFromStub()
}

// stubStrncmp implements strncmp(s1, s2, n)
func (s *LibcStubs) stubStrncmp() {
	n := int(s.emu.X(2))
	s1, _ := s.emu.MemReadString(s.emu.X(0), n)
	s2, _ := s.emu.MemReadString(s.emu.X(1), n)

	// Truncate to n
	if len(s1) > n {
		s1 = s1[:n]
	}
	if len(s2) > n {
		s2 = s2[:n]
	}

	var result uint64
	if s1 < s2 {
		result = 0xffffffffffffffff
	} else if s1 > s2 {
		result = 1
	} else {
		result = 0
	}

	s.emu.SetX(0, result)
	s.returnFromStub()
}

// stubStrcpy implements strcpy(dest, src)
func (s *LibcStubs) stubStrcpy() {
	dest := s.emu.X(0)
	src := s.emu.X(1)
	str, _ := s.emu.MemReadString(src, 4096)
	s.emu.MemWriteString(dest, str)

	s.emu.SetX(0, dest)
	s.returnFromStub()
}

// stubStrncpy implements strncpy(dest, src, n)
func (s *LibcStubs) stubStrncpy() {
	dest := s.emu.X(0)
	src := s.emu.X(1)
	n := s.emu.X(2)

	str, _ := s.emu.MemReadString(src, int(n))
	if uint64(len(str)) < n {
		// Pad with zeros
		data := make([]byte, n)
		copy(data, str)
		s.emu.MemWrite(dest, data)
	} else {
		s.emu.MemWriteString(dest, str[:n])
	}

	s.emu.SetX(0, dest)
	s.returnFromStub()
}

// Mocked time for determinism
var (
	MockTimeSec  = int64(1704067200) // 2024-01-01 00:00:00 UTC
	MockTimeUSec = int64(0)
	MockTimeNSec = int64(0)
)

// stubGettimeofday implements gettimeofday(tv, tz)
func (s *LibcStubs) stubGettimeofday() {
	tv := s.emu.X(0)

	if tv != 0 {
		// struct timeval { time_t tv_sec; suseconds_t tv_usec; }
		s.emu.MemWriteU64(tv, uint64(MockTimeSec))
		s.emu.MemWriteU64(tv+8, uint64(MockTimeUSec))
	}

	s.log("gettimeofday", formatPtr("tv", tv, "sec", uint64(MockTimeSec)))
	s.emu.SetX(0, 0) // success
	s.returnFromStub()
}

// stubClockGettime implements clock_gettime(clockid, tp)
func (s *LibcStubs) stubClockGettime() {
	_ = s.emu.X(0) // clockid (ignored)
	tp := s.emu.X(1)

	if tp != 0 {
		// struct timespec { time_t tv_sec; long tv_nsec; }
		s.emu.MemWriteU64(tp, uint64(MockTimeSec))
		s.emu.MemWriteU64(tp+8, uint64(MockTimeNSec))
	}

	s.log("clock_gettime", formatPtr("tp", tp, "sec", uint64(MockTimeSec)))
	s.emu.SetX(0, 0) // success
	s.returnFromStub()
}

// stubTime implements time(tloc)
func (s *LibcStubs) stubTime() {
	tloc := s.emu.X(0)

	if tloc != 0 {
		s.emu.MemWriteU64(tloc, uint64(MockTimeSec))
	}

	s.log("time", formatPtr("sec", uint64(MockTimeSec), "", 0))
	s.emu.SetX(0, uint64(MockTimeSec))
	s.returnFromStub()
}

// Helper functions

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func formatPtr(name string, val uint64, name2 string, val2 uint64) string {
	if name2 == "" {
		return name + "=" + formatHex(val)
	}
	return name + "=" + formatHex(val) + " " + name2 + "=" + formatHex(val2)
}

func formatHex(v uint64) string {
	if v < 0x10000 {
		return string(rune('0'+v/1000%10)) + string(rune('0'+v/100%10)) + string(rune('0'+v/10%10)) + string(rune('0'+v%10))
	}
	// Simple hex formatting without fmt
	const digits = "0123456789abcdef"
	buf := make([]byte, 18) // "0x" + 16 hex digits
	buf[0] = '0'
	buf[1] = 'x'
	for i := 17; i >= 2; i-- {
		buf[i] = digits[v&0xf]
		v >>= 4
	}
	// Trim leading zeros
	start := 2
	for start < 17 && buf[start] == '0' {
		start++
	}
	return string(buf[:2]) + string(buf[start:])
}

func formatMemop(dest, src, n uint64) string {
	return "dst=" + formatHex(dest) + " src=" + formatHex(src) + " n=" + formatHex(n)
}
