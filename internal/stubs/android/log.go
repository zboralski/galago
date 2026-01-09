// Package android provides stub implementations for Android-specific functions.
package android

import (
	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

func init() {
	stubs.RegisterFunc("android", "__android_log_print", stubAndroidLogPrint)
	stubs.RegisterFunc("android", "__android_log_write", stubAndroidLogWrite)
	stubs.RegisterFunc("android", "__android_log_vprint", stubAndroidLogVprint)
	stubs.RegisterFunc("android", "__android_log_buf_print", stubAndroidLogBufPrint)
	stubs.RegisterFunc("android", "__android_log_buf_write", stubAndroidLogBufWrite)
	stubs.RegisterFunc("android", "__android_log_assert", stubAndroidLogAssert)

	// Syslog
	stubs.RegisterFunc("android", "openlog", stubOpenlog)
	stubs.RegisterFunc("android", "syslog", stubSyslog)
	stubs.RegisterFunc("android", "closelog", stubCloselog)
}

func stubAndroidLogPrint(emu *emulator.Emulator) bool {
	// int __android_log_print(int prio, const char *tag, const char *fmt, ...)
	// prio := emu.X(0)
	tagPtr := emu.X(1)
	fmtPtr := emu.X(2)

	tag, _ := emu.MemReadString(tagPtr, 64)
	format, _ := emu.MemReadString(fmtPtr, 256)

	stubs.DefaultRegistry.Log("android", "__android_log_print", tag+": "+format)

	emu.SetX(0, 0) // Return number of bytes written
	stubs.ReturnFromStub(emu)
	return false
}

func stubAndroidLogWrite(emu *emulator.Emulator) bool {
	// int __android_log_write(int prio, const char *tag, const char *text)
	tagPtr := emu.X(1)
	textPtr := emu.X(2)

	tag, _ := emu.MemReadString(tagPtr, 64)
	text, _ := emu.MemReadString(textPtr, 256)

	stubs.DefaultRegistry.Log("android", "__android_log_write", tag+": "+text)

	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubAndroidLogVprint(emu *emulator.Emulator) bool {
	// Like log_print but with va_list
	tagPtr := emu.X(1)
	tag, _ := emu.MemReadString(tagPtr, 64)

	stubs.DefaultRegistry.Log("android", "__android_log_vprint", tag)

	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubAndroidLogBufPrint(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubAndroidLogBufWrite(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubAndroidLogAssert(emu *emulator.Emulator) bool {
	// Usually aborts, but we just log and continue
	condPtr := emu.X(0)
	tagPtr := emu.X(1)

	cond, _ := emu.MemReadString(condPtr, 64)
	tag, _ := emu.MemReadString(tagPtr, 64)

	stubs.DefaultRegistry.Log("android", "__android_log_assert", tag+": "+cond)

	stubs.ReturnFromStub(emu)
	return false
}

func stubOpenlog(emu *emulator.Emulator) bool {
	// void openlog(const char *ident, int option, int facility)
	stubs.ReturnFromStub(emu)
	return false
}

func stubSyslog(emu *emulator.Emulator) bool {
	// void syslog(int priority, const char *format, ...)
	fmtPtr := emu.X(1)
	format, _ := emu.MemReadString(fmtPtr, 256)

	stubs.DefaultRegistry.Log("android", "syslog", format)

	stubs.ReturnFromStub(emu)
	return false
}

func stubCloselog(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}
