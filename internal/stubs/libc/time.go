package libc

import (
	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

// Mocked time for deterministic execution.
var (
	MockTimeSec  = int64(1704067200) // 2024-01-01 00:00:00 UTC
	MockTimeUSec = int64(0)
	MockTimeNSec = int64(0)
)

func init() {
	stubs.RegisterFunc("libc", "gettimeofday", stubGettimeofday)
	stubs.RegisterFunc("libc", "clock_gettime", stubClockGettime)
	stubs.RegisterFunc("libc", "time", stubTime)
	stubs.RegisterFunc("libc", "clock", stubClock)
	stubs.RegisterFunc("libc", "nanosleep", stubNanosleep)
	stubs.RegisterFunc("libc", "usleep", stubUsleep)
	stubs.RegisterFunc("libc", "sleep", stubSleep)
}

func stubGettimeofday(emu *emulator.Emulator) bool {
	tv := emu.X(0)

	if tv != 0 {
		// struct timeval { time_t tv_sec; suseconds_t tv_usec; }
		emu.MemWriteU64(tv, uint64(MockTimeSec))
		emu.MemWriteU64(tv+8, uint64(MockTimeUSec))
	}

	stubs.DefaultRegistry.Log("libc", "gettimeofday", stubs.FormatPtrPair("tv", tv, "sec", uint64(MockTimeSec)))
	emu.SetX(0, 0) // success
	stubs.ReturnFromStub(emu)
	return false
}

func stubClockGettime(emu *emulator.Emulator) bool {
	_ = emu.X(0) // clockid (ignored)
	tp := emu.X(1)

	if tp != 0 {
		// struct timespec { time_t tv_sec; long tv_nsec; }
		emu.MemWriteU64(tp, uint64(MockTimeSec))
		emu.MemWriteU64(tp+8, uint64(MockTimeNSec))
	}

	stubs.DefaultRegistry.Log("libc", "clock_gettime", stubs.FormatPtrPair("tp", tp, "sec", uint64(MockTimeSec)))
	emu.SetX(0, 0) // success
	stubs.ReturnFromStub(emu)
	return false
}

func stubTime(emu *emulator.Emulator) bool {
	tloc := emu.X(0)

	if tloc != 0 {
		emu.MemWriteU64(tloc, uint64(MockTimeSec))
	}

	stubs.DefaultRegistry.Log("libc", "time", stubs.FormatPtr("sec", uint64(MockTimeSec)))
	emu.SetX(0, uint64(MockTimeSec))
	stubs.ReturnFromStub(emu)
	return false
}

func stubClock(emu *emulator.Emulator) bool {
	// Return a fixed clock value (in ticks)
	emu.SetX(0, 1000000)
	stubs.ReturnFromStub(emu)
	return false
}

func stubNanosleep(emu *emulator.Emulator) bool {
	// Just return success without sleeping
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubUsleep(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubSleep(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}
