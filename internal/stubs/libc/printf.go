package libc

import (
	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

func init() {
	stubs.RegisterFunc("libc", "printf", stubPrintf)
	stubs.RegisterFunc("libc", "fprintf", stubFprintf)
	stubs.RegisterFunc("libc", "vprintf", stubVprintf)
	stubs.RegisterFunc("libc", "vfprintf", stubVfprintf)
	stubs.RegisterFunc("libc", "sprintf", stubSprintf)
	stubs.RegisterFunc("libc", "snprintf", stubSnprintf)
	stubs.RegisterFunc("libc", "vsprintf", stubVsprintf)
	stubs.RegisterFunc("libc", "vsnprintf", stubVsnprintf)
	stubs.RegisterFunc("libc", "asprintf", stubAsprintf)
	stubs.RegisterFunc("libc", "vasprintf", stubVasprintf)

	// Fortified variants (__*_chk)
	stubs.RegisterFunc("libc", "__vsnprintf_chk", stubVsnprintfChk)
	stubs.RegisterFunc("libc", "__snprintf_chk", stubSnprintfChk)
	stubs.RegisterFunc("libc", "__sprintf_chk", stubSprintfChk)
	stubs.RegisterFunc("libc", "__printf_chk", stubPrintfChk)
	stubs.RegisterFunc("libc", "__fprintf_chk", stubFprintfChk)

	stubs.RegisterFunc("libc", "puts", stubPuts)
	stubs.RegisterFunc("libc", "fputs", stubFputs)
	stubs.RegisterFunc("libc", "putchar", stubPutchar)
	stubs.RegisterFunc("libc", "fputc", stubFputc)
	stubs.RegisterFunc("libc", "putc", stubFputc)
	stubs.RegisterFunc("libc", "fwrite", stubFwrite)
	stubs.RegisterFunc("libc", "fread", stubFread)
	stubs.RegisterFunc("libc", "fflush", stubFflush)
	stubs.RegisterFunc("libc", "fclose", stubFclose)
	stubs.RegisterFunc("libc", "fopen", stubFopen)
	stubs.RegisterFunc("libc", "fseek", stubFseek)
	stubs.RegisterFunc("libc", "ftell", stubFtell)
	stubs.RegisterFunc("libc", "rewind", stubRewind)
	stubs.RegisterFunc("libc", "feof", stubFeof)
	stubs.RegisterFunc("libc", "ferror", stubFerror)
	stubs.RegisterFunc("libc", "clearerr", stubClearerr)
	stubs.RegisterFunc("libc", "fileno", stubFileno)

	stubs.RegisterFunc("libc", "perror", stubPerror)
	stubs.RegisterFunc("libc", "strerror", stubStrerror)
	stubs.RegisterFunc("libc", "strerror_r", stubStrerrorR)
}

func stubPrintf(emu *emulator.Emulator) bool {
	fmtPtr := emu.X(0)
	format, _ := emu.MemReadString(fmtPtr, 256)
	stubs.DefaultRegistry.Log("libc", "printf", format)
	emu.SetX(0, uint64(len(format)))
	stubs.ReturnFromStub(emu)
	return false
}

func stubFprintf(emu *emulator.Emulator) bool {
	// stream := emu.X(0)
	fmtPtr := emu.X(1)
	format, _ := emu.MemReadString(fmtPtr, 256)
	stubs.DefaultRegistry.Log("libc", "fprintf", format)
	emu.SetX(0, uint64(len(format)))
	stubs.ReturnFromStub(emu)
	return false
}

func stubVprintf(emu *emulator.Emulator) bool {
	fmtPtr := emu.X(0)
	format, _ := emu.MemReadString(fmtPtr, 256)
	stubs.DefaultRegistry.Log("libc", "vprintf", format)
	emu.SetX(0, uint64(len(format)))
	stubs.ReturnFromStub(emu)
	return false
}

func stubVfprintf(emu *emulator.Emulator) bool {
	fmtPtr := emu.X(1)
	format, _ := emu.MemReadString(fmtPtr, 256)
	stubs.DefaultRegistry.Log("libc", "vfprintf", format)
	emu.SetX(0, uint64(len(format)))
	stubs.ReturnFromStub(emu)
	return false
}

func stubSprintf(emu *emulator.Emulator) bool {
	dest := emu.X(0)
	fmtPtr := emu.X(1)
	format, _ := emu.MemReadString(fmtPtr, 256)

	// Write format string directly (no actual formatting)
	emu.MemWriteString(dest, format)

	emu.SetX(0, uint64(len(format)))
	stubs.ReturnFromStub(emu)
	return false
}

func stubSnprintf(emu *emulator.Emulator) bool {
	dest := emu.X(0)
	n := emu.X(1)
	fmtPtr := emu.X(2)
	format, _ := emu.MemReadString(fmtPtr, int(n))

	if n > 0 {
		if uint64(len(format)) >= n {
			format = format[:n-1]
		}
		emu.MemWriteString(dest, format)
	}

	emu.SetX(0, uint64(len(format)))
	stubs.ReturnFromStub(emu)
	return false
}

func stubVsprintf(emu *emulator.Emulator) bool {
	return stubSprintf(emu)
}

func stubVsnprintf(emu *emulator.Emulator) bool {
	return stubSnprintf(emu)
}

// Fortified variants - __*_chk functions add buffer overflow checking
// They have additional flag/slen parameters before the format string

func stubVsnprintfChk(emu *emulator.Emulator) bool {
	// int __vsnprintf_chk(char *s, size_t maxlen, int flag, size_t slen, const char *format, va_list ap)
	dest := emu.X(0)
	n := emu.X(1)
	// flag := emu.X(2)
	// slen := emu.X(3)
	fmtPtr := emu.X(4)
	format, _ := emu.MemReadString(fmtPtr, int(n))

	if n > 0 {
		if uint64(len(format)) >= n {
			format = format[:n-1]
		}
		emu.MemWriteString(dest, format)
	}

	emu.SetX(0, uint64(len(format)))
	stubs.ReturnFromStub(emu)
	return false
}

func stubSnprintfChk(emu *emulator.Emulator) bool {
	// int __snprintf_chk(char *s, size_t maxlen, int flag, size_t slen, const char *format, ...)
	dest := emu.X(0)
	n := emu.X(1)
	// flag := emu.X(2)
	// slen := emu.X(3)
	fmtPtr := emu.X(4)
	format, _ := emu.MemReadString(fmtPtr, int(n))

	if n > 0 {
		if uint64(len(format)) >= n {
			format = format[:n-1]
		}
		emu.MemWriteString(dest, format)
	}

	emu.SetX(0, uint64(len(format)))
	stubs.ReturnFromStub(emu)
	return false
}

func stubSprintfChk(emu *emulator.Emulator) bool {
	// int __sprintf_chk(char *s, int flag, size_t slen, const char *format, ...)
	dest := emu.X(0)
	// flag := emu.X(1)
	// slen := emu.X(2)
	fmtPtr := emu.X(3)
	format, _ := emu.MemReadString(fmtPtr, 256)

	emu.MemWriteString(dest, format)
	emu.SetX(0, uint64(len(format)))
	stubs.ReturnFromStub(emu)
	return false
}

func stubPrintfChk(emu *emulator.Emulator) bool {
	// int __printf_chk(int flag, const char *format, ...)
	// flag := emu.X(0)
	fmtPtr := emu.X(1)
	format, _ := emu.MemReadString(fmtPtr, 256)
	stubs.DefaultRegistry.Log("libc", "__printf_chk", format)
	emu.SetX(0, uint64(len(format)))
	stubs.ReturnFromStub(emu)
	return false
}

func stubFprintfChk(emu *emulator.Emulator) bool {
	// int __fprintf_chk(FILE *stream, int flag, const char *format, ...)
	// stream := emu.X(0)
	// flag := emu.X(1)
	fmtPtr := emu.X(2)
	format, _ := emu.MemReadString(fmtPtr, 256)
	stubs.DefaultRegistry.Log("libc", "__fprintf_chk", format)
	emu.SetX(0, uint64(len(format)))
	stubs.ReturnFromStub(emu)
	return false
}

func stubAsprintf(emu *emulator.Emulator) bool {
	retPtr := emu.X(0)
	fmtPtr := emu.X(1)
	format, _ := emu.MemReadString(fmtPtr, 256)

	// Allocate buffer and write
	buf := emu.Malloc(uint64(len(format) + 1))
	emu.MemWriteString(buf, format)
	emu.MemWriteU64(retPtr, buf)

	emu.SetX(0, uint64(len(format)))
	stubs.ReturnFromStub(emu)
	return false
}

func stubVasprintf(emu *emulator.Emulator) bool {
	return stubAsprintf(emu)
}

func stubPuts(emu *emulator.Emulator) bool {
	strPtr := emu.X(0)
	str, _ := emu.MemReadString(strPtr, 256)
	stubs.DefaultRegistry.Log("libc", "puts", str)
	emu.SetX(0, 0) // Non-negative on success
	stubs.ReturnFromStub(emu)
	return false
}

func stubFputs(emu *emulator.Emulator) bool {
	strPtr := emu.X(0)
	// stream := emu.X(1)
	str, _ := emu.MemReadString(strPtr, 256)
	stubs.DefaultRegistry.Log("libc", "fputs", str)
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubPutchar(emu *emulator.Emulator) bool {
	c := emu.X(0) & 0xFF
	emu.SetX(0, c)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFputc(emu *emulator.Emulator) bool {
	c := emu.X(0) & 0xFF
	emu.SetX(0, c)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFwrite(emu *emulator.Emulator) bool {
	// ptr := emu.X(0)
	size := emu.X(1)
	nmemb := emu.X(2)
	// stream := emu.X(3)
	emu.SetX(0, nmemb) // Return items written
	_ = size
	stubs.ReturnFromStub(emu)
	return false
}

func stubFread(emu *emulator.Emulator) bool {
	// Return 0 items read
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFflush(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFclose(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFopen(emu *emulator.Emulator) bool {
	// Return NULL (file not found)
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFseek(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFtell(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubRewind(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func stubFeof(emu *emulator.Emulator) bool {
	emu.SetX(0, 1) // Return EOF
	stubs.ReturnFromStub(emu)
	return false
}

func stubFerror(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubClearerr(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func stubFileno(emu *emulator.Emulator) bool {
	emu.SetX(0, 1) // Return stdout fd
	stubs.ReturnFromStub(emu)
	return false
}

func stubPerror(emu *emulator.Emulator) bool {
	strPtr := emu.X(0)
	str, _ := emu.MemReadString(strPtr, 256)
	stubs.DefaultRegistry.Log("libc", "perror", str)
	stubs.ReturnFromStub(emu)
	return false
}

func stubStrerror(emu *emulator.Emulator) bool {
	// errnum := emu.X(0)
	// Return "Unknown error" string
	buf := emu.Malloc(32)
	emu.MemWriteString(buf, "Unknown error")
	emu.SetX(0, buf)
	stubs.ReturnFromStub(emu)
	return false
}

func stubStrerrorR(emu *emulator.Emulator) bool {
	// errnum := emu.X(0)
	buf := emu.X(1)
	// buflen := emu.X(2)
	emu.MemWriteString(buf, "Unknown error")
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}
