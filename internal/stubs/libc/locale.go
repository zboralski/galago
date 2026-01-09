// Package libc provides stub implementations for libc functions.
package libc

import (
	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

func init() {
	// Locale functions
	stubs.RegisterFunc("libc", "setlocale", stubSetlocale)
	stubs.RegisterFunc("libc", "newlocale", stubNewlocale)
	stubs.RegisterFunc("libc", "uselocale", stubUselocale)
	stubs.RegisterFunc("libc", "freelocale", stubFreelocale)
	stubs.RegisterFunc("libc", "localeconv", stubLocaleconv)

	// System configuration
	stubs.RegisterFunc("libc", "sysconf", stubSysconf)
	stubs.RegisterFunc("libc", "getenv", stubGetenv)
	stubs.RegisterFunc("libc", "setenv", stubSetenv)
	stubs.RegisterFunc("libc", "unsetenv", stubUnsetenv)
	stubs.RegisterFunc("libc", "putenv", stubPutenv)

	// Character classification
	stubs.RegisterFunc("libc", "isalpha", stubIsalpha)
	stubs.RegisterFunc("libc", "isdigit", stubIsdigit)
	stubs.RegisterFunc("libc", "isalnum", stubIsalnum)
	stubs.RegisterFunc("libc", "isspace", stubIsspace)
	stubs.RegisterFunc("libc", "isupper", stubIsupper)
	stubs.RegisterFunc("libc", "islower", stubIslower)
	stubs.RegisterFunc("libc", "isxdigit", stubIsxdigit)
	stubs.RegisterFunc("libc", "isprint", stubIsprint)
	stubs.RegisterFunc("libc", "iscntrl", stubIscntrl)
	stubs.RegisterFunc("libc", "ispunct", stubIspunct)
	stubs.RegisterFunc("libc", "isgraph", stubIsgraph)
	stubs.RegisterFunc("libc", "isblank", stubIsblank)
	stubs.RegisterFunc("libc", "toupper", stubToupper)
	stubs.RegisterFunc("libc", "tolower", stubTolower)

	// Wide character functions
	stubs.RegisterFunc("libc", "wcscpy", stubWcscpy)
	stubs.RegisterFunc("libc", "wcslen", stubWcslen)
	stubs.RegisterFunc("libc", "wcscmp", stubWcscmp)
	stubs.RegisterFunc("libc", "wcsncpy", stubWcsncpy)
	stubs.RegisterFunc("libc", "wcsncmp", stubWcsncmp)
	stubs.RegisterFunc("libc", "wcschr", stubWcschr)
	stubs.RegisterFunc("libc", "wcsrchr", stubWcsrchr)
	stubs.RegisterFunc("libc", "wcscat", stubWcscat)
	stubs.RegisterFunc("libc", "wcsncat", stubWcsncat)

	// Multibyte/wide conversion
	stubs.RegisterFunc("libc", "mbstowcs", stubMbstowcs)
	stubs.RegisterFunc("libc", "wcstombs", stubWcstombs)
	stubs.RegisterFunc("libc", "mbtowc", stubMbtowc)
	stubs.RegisterFunc("libc", "wctomb", stubWctomb)
	stubs.RegisterFunc("libc", "mblen", stubMblen)
}

var (
	// Static buffer for locale name
	localeNameBuf uint64
	// Static buffer for localeconv result
	localeconvBuf uint64
	// Environment storage (mock)
	envVars = make(map[string]string)
)

func stubSetlocale(emu *emulator.Emulator) bool {
	// char *setlocale(int category, const char *locale)
	// category := emu.X(0)
	localePtr := emu.X(1)

	locale := ""
	if localePtr != 0 {
		locale, _ = emu.MemReadString(localePtr, 64)
	}

	stubs.DefaultRegistry.Log("libc", "setlocale", locale)

	// Return pointer to "C" locale string
	if localeNameBuf == 0 {
		localeNameBuf = emu.Malloc(8)
		emu.MemWriteString(localeNameBuf, "C")
	}
	emu.SetX(0, localeNameBuf)
	stubs.ReturnFromStub(emu)
	return false
}

func stubNewlocale(emu *emulator.Emulator) bool {
	// locale_t newlocale(int category_mask, const char *locale, locale_t base)
	// Return a fake locale handle
	handle := emu.Malloc(8)
	emu.MemWriteU64(handle, 1) // Non-null marker
	emu.SetX(0, handle)
	stubs.ReturnFromStub(emu)
	return false
}

func stubUselocale(emu *emulator.Emulator) bool {
	// locale_t uselocale(locale_t newloc)
	// Return previous locale (fake)
	prev := emu.Malloc(8)
	emu.MemWriteU64(prev, 1)
	emu.SetX(0, prev)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFreelocale(emu *emulator.Emulator) bool {
	// void freelocale(locale_t locale)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLocaleconv(emu *emulator.Emulator) bool {
	// struct lconv *localeconv(void)
	// Return a minimal lconv structure with C locale defaults
	if localeconvBuf == 0 {
		// struct lconv is complex, allocate enough space
		localeconvBuf = emu.Malloc(128)
		// decimal_point = "."
		decPt := emu.Malloc(4)
		emu.MemWriteString(decPt, ".")
		emu.MemWriteU64(localeconvBuf, decPt)
		// thousands_sep = ""
		thousSep := emu.Malloc(4)
		emu.MemWriteString(thousSep, "")
		emu.MemWriteU64(localeconvBuf+8, thousSep)
	}
	emu.SetX(0, localeconvBuf)
	stubs.ReturnFromStub(emu)
	return false
}

func stubSysconf(emu *emulator.Emulator) bool {
	// long sysconf(int name)
	name := emu.X(0)

	var result uint64
	switch name {
	case 30: // _SC_PAGESIZE / _SC_PAGE_SIZE
		result = 4096
	case 84: // _SC_NPROCESSORS_ONLN
		result = 4
	case 83: // _SC_NPROCESSORS_CONF
		result = 4
	case 2: // _SC_CLK_TCK
		result = 100
	case 0: // _SC_ARG_MAX
		result = 131072
	case 1: // _SC_CHILD_MAX
		result = 999
	case 4: // _SC_OPEN_MAX
		result = 1024
	default:
		result = ^uint64(0) // -1 for unknown
	}

	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func stubGetenv(emu *emulator.Emulator) bool {
	// char *getenv(const char *name)
	namePtr := emu.X(0)
	name, _ := emu.MemReadString(namePtr, 256)

	stubs.DefaultRegistry.Log("libc", "getenv", name)

	// Check mock environment
	if val, ok := envVars[name]; ok {
		buf := emu.Malloc(uint64(len(val) + 1))
		emu.MemWriteString(buf, val)
		emu.SetX(0, buf)
	} else {
		// Return common defaults for certain variables
		var result string
		switch name {
		case "PATH":
			result = "/system/bin:/system/xbin"
		case "HOME":
			result = "/data/data/com.app"
		case "TMPDIR", "TEMP", "TMP":
			result = "/data/local/tmp"
		case "LANG", "LC_ALL":
			result = "C"
		default:
			emu.SetX(0, 0) // NULL
			stubs.ReturnFromStub(emu)
			return false
		}
		buf := emu.Malloc(uint64(len(result) + 1))
		emu.MemWriteString(buf, result)
		emu.SetX(0, buf)
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubSetenv(emu *emulator.Emulator) bool {
	// int setenv(const char *name, const char *value, int overwrite)
	namePtr := emu.X(0)
	valuePtr := emu.X(1)
	// overwrite := emu.X(2)

	name, _ := emu.MemReadString(namePtr, 256)
	value, _ := emu.MemReadString(valuePtr, 1024)

	envVars[name] = value
	emu.SetX(0, 0) // Success
	stubs.ReturnFromStub(emu)
	return false
}

func stubUnsetenv(emu *emulator.Emulator) bool {
	// int unsetenv(const char *name)
	namePtr := emu.X(0)
	name, _ := emu.MemReadString(namePtr, 256)

	delete(envVars, name)
	emu.SetX(0, 0) // Success
	stubs.ReturnFromStub(emu)
	return false
}

func stubPutenv(emu *emulator.Emulator) bool {
	// int putenv(char *string)
	emu.SetX(0, 0) // Success
	stubs.ReturnFromStub(emu)
	return false
}

// Character classification helpers
func isInRange(c byte, low, high byte) bool {
	return c >= low && c <= high
}

func stubIsalpha(emu *emulator.Emulator) bool {
	c := byte(emu.X(0))
	result := uint64(0)
	if isInRange(c, 'A', 'Z') || isInRange(c, 'a', 'z') {
		result = 1
	}
	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func stubIsdigit(emu *emulator.Emulator) bool {
	c := byte(emu.X(0))
	result := uint64(0)
	if isInRange(c, '0', '9') {
		result = 1
	}
	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func stubIsalnum(emu *emulator.Emulator) bool {
	c := byte(emu.X(0))
	result := uint64(0)
	if isInRange(c, 'A', 'Z') || isInRange(c, 'a', 'z') || isInRange(c, '0', '9') {
		result = 1
	}
	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func stubIsspace(emu *emulator.Emulator) bool {
	c := byte(emu.X(0))
	result := uint64(0)
	if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v' {
		result = 1
	}
	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func stubIsupper(emu *emulator.Emulator) bool {
	c := byte(emu.X(0))
	result := uint64(0)
	if isInRange(c, 'A', 'Z') {
		result = 1
	}
	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func stubIslower(emu *emulator.Emulator) bool {
	c := byte(emu.X(0))
	result := uint64(0)
	if isInRange(c, 'a', 'z') {
		result = 1
	}
	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func stubIsxdigit(emu *emulator.Emulator) bool {
	c := byte(emu.X(0))
	result := uint64(0)
	if isInRange(c, '0', '9') || isInRange(c, 'A', 'F') || isInRange(c, 'a', 'f') {
		result = 1
	}
	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func stubIsprint(emu *emulator.Emulator) bool {
	c := byte(emu.X(0))
	result := uint64(0)
	if c >= 0x20 && c <= 0x7e {
		result = 1
	}
	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func stubIscntrl(emu *emulator.Emulator) bool {
	c := byte(emu.X(0))
	result := uint64(0)
	if c < 0x20 || c == 0x7f {
		result = 1
	}
	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func stubIspunct(emu *emulator.Emulator) bool {
	c := byte(emu.X(0))
	result := uint64(0)
	if (c >= 0x21 && c <= 0x2f) || (c >= 0x3a && c <= 0x40) ||
		(c >= 0x5b && c <= 0x60) || (c >= 0x7b && c <= 0x7e) {
		result = 1
	}
	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func stubIsgraph(emu *emulator.Emulator) bool {
	c := byte(emu.X(0))
	result := uint64(0)
	if c >= 0x21 && c <= 0x7e {
		result = 1
	}
	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func stubIsblank(emu *emulator.Emulator) bool {
	c := byte(emu.X(0))
	result := uint64(0)
	if c == ' ' || c == '\t' {
		result = 1
	}
	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func stubToupper(emu *emulator.Emulator) bool {
	c := byte(emu.X(0))
	if isInRange(c, 'a', 'z') {
		c = c - 'a' + 'A'
	}
	emu.SetX(0, uint64(c))
	stubs.ReturnFromStub(emu)
	return false
}

func stubTolower(emu *emulator.Emulator) bool {
	c := byte(emu.X(0))
	if isInRange(c, 'A', 'Z') {
		c = c - 'A' + 'a'
	}
	emu.SetX(0, uint64(c))
	stubs.ReturnFromStub(emu)
	return false
}

// Wide character stubs - minimal implementations

func stubWcscpy(emu *emulator.Emulator) bool {
	dest := emu.X(0)
	src := emu.X(1)
	// Copy wide string (4 bytes per wchar_t on most platforms)
	for i := uint64(0); i < 4096; i += 4 {
		wc, _ := emu.MemReadU32(src + i)
		emu.MemWriteU32(dest+i, wc)
		if wc == 0 {
			break
		}
	}
	emu.SetX(0, dest)
	stubs.ReturnFromStub(emu)
	return false
}

func stubWcslen(emu *emulator.Emulator) bool {
	s := emu.X(0)
	length := uint64(0)
	for i := uint64(0); i < 4096; i += 4 {
		wc, _ := emu.MemReadU32(s + i)
		if wc == 0 {
			break
		}
		length++
	}
	emu.SetX(0, length)
	stubs.ReturnFromStub(emu)
	return false
}

func stubWcscmp(emu *emulator.Emulator) bool {
	s1 := emu.X(0)
	s2 := emu.X(1)
	for i := uint64(0); i < 4096; i += 4 {
		wc1, _ := emu.MemReadU32(s1 + i)
		wc2, _ := emu.MemReadU32(s2 + i)
		if wc1 != wc2 {
			if wc1 < wc2 {
				emu.SetX(0, ^uint64(0)) // -1
			} else {
				emu.SetX(0, 1)
			}
			stubs.ReturnFromStub(emu)
			return false
		}
		if wc1 == 0 {
			break
		}
	}
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubWcsncpy(emu *emulator.Emulator) bool {
	dest := emu.X(0)
	src := emu.X(1)
	n := emu.X(2)
	for i := uint64(0); i < n*4 && i < 4096; i += 4 {
		wc, _ := emu.MemReadU32(src + i)
		emu.MemWriteU32(dest+i, wc)
		if wc == 0 {
			// Pad with nulls
			for j := i + 4; j < n*4 && j < 4096; j += 4 {
				emu.MemWriteU32(dest+j, 0)
			}
			break
		}
	}
	emu.SetX(0, dest)
	stubs.ReturnFromStub(emu)
	return false
}

func stubWcsncmp(emu *emulator.Emulator) bool {
	s1 := emu.X(0)
	s2 := emu.X(1)
	n := emu.X(2)
	for i := uint64(0); i < n && i < 1024; i++ {
		wc1, _ := emu.MemReadU32(s1 + i*4)
		wc2, _ := emu.MemReadU32(s2 + i*4)
		if wc1 != wc2 {
			if wc1 < wc2 {
				emu.SetX(0, ^uint64(0))
			} else {
				emu.SetX(0, 1)
			}
			stubs.ReturnFromStub(emu)
			return false
		}
		if wc1 == 0 {
			break
		}
	}
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubWcschr(emu *emulator.Emulator) bool {
	s := emu.X(0)
	wc := uint32(emu.X(1))
	for i := uint64(0); i < 4096; i += 4 {
		ch, _ := emu.MemReadU32(s + i)
		if ch == wc {
			emu.SetX(0, s+i)
			stubs.ReturnFromStub(emu)
			return false
		}
		if ch == 0 {
			break
		}
	}
	emu.SetX(0, 0) // Not found
	stubs.ReturnFromStub(emu)
	return false
}

func stubWcsrchr(emu *emulator.Emulator) bool {
	s := emu.X(0)
	wc := uint32(emu.X(1))
	lastFound := uint64(0)
	for i := uint64(0); i < 4096; i += 4 {
		ch, _ := emu.MemReadU32(s + i)
		if ch == wc {
			lastFound = s + i
		}
		if ch == 0 {
			break
		}
	}
	emu.SetX(0, lastFound)
	stubs.ReturnFromStub(emu)
	return false
}

func stubWcscat(emu *emulator.Emulator) bool {
	dest := emu.X(0)
	src := emu.X(1)
	// Find end of dest
	end := dest
	for i := uint64(0); i < 4096; i += 4 {
		wc, _ := emu.MemReadU32(end + i)
		if wc == 0 {
			end = end + i
			break
		}
	}
	// Copy src
	for i := uint64(0); i < 4096; i += 4 {
		wc, _ := emu.MemReadU32(src + i)
		emu.MemWriteU32(end+i, wc)
		if wc == 0 {
			break
		}
	}
	emu.SetX(0, dest)
	stubs.ReturnFromStub(emu)
	return false
}

func stubWcsncat(emu *emulator.Emulator) bool {
	dest := emu.X(0)
	src := emu.X(1)
	n := emu.X(2)
	// Find end of dest
	end := dest
	for i := uint64(0); i < 4096; i += 4 {
		wc, _ := emu.MemReadU32(end + i)
		if wc == 0 {
			end = end + i
			break
		}
	}
	// Copy up to n chars from src
	for i := uint64(0); i < n && i < 1024; i++ {
		wc, _ := emu.MemReadU32(src + i*4)
		emu.MemWriteU32(end+i*4, wc)
		if wc == 0 {
			stubs.ReturnFromStub(emu)
			emu.SetX(0, dest)
			return false
		}
	}
	// Null terminate
	emu.MemWriteU32(end+n*4, 0)
	emu.SetX(0, dest)
	stubs.ReturnFromStub(emu)
	return false
}

// Multibyte/wide conversion stubs

func stubMbstowcs(emu *emulator.Emulator) bool {
	// size_t mbstowcs(wchar_t *dest, const char *src, size_t n)
	dest := emu.X(0)
	src := emu.X(1)
	n := emu.X(2)

	// Simple ASCII conversion (1 byte -> 1 wchar_t)
	count := uint64(0)
	for i := uint64(0); i < n && i < 1024; i++ {
		b, _ := emu.MemReadU8(src + i)
		if dest != 0 {
			emu.MemWriteU32(dest+i*4, uint32(b))
		}
		if b == 0 {
			break
		}
		count++
	}
	emu.SetX(0, count)
	stubs.ReturnFromStub(emu)
	return false
}

func stubWcstombs(emu *emulator.Emulator) bool {
	// size_t wcstombs(char *dest, const wchar_t *src, size_t n)
	dest := emu.X(0)
	src := emu.X(1)
	n := emu.X(2)

	// Simple conversion (wchar_t -> 1 byte if ASCII)
	count := uint64(0)
	for i := uint64(0); count < n && i < 1024; i++ {
		wc, _ := emu.MemReadU32(src + i*4)
		if dest != 0 && count < n {
			emu.MemWriteU8(dest+count, byte(wc&0xff))
		}
		if wc == 0 {
			break
		}
		count++
	}
	emu.SetX(0, count)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMbtowc(emu *emulator.Emulator) bool {
	// int mbtowc(wchar_t *pwc, const char *s, size_t n)
	pwc := emu.X(0)
	s := emu.X(1)

	if s == 0 {
		emu.SetX(0, 0) // No state dependency
		stubs.ReturnFromStub(emu)
		return false
	}

	b, _ := emu.MemReadU8(s)
	if pwc != 0 {
		emu.MemWriteU32(pwc, uint32(b))
	}
	if b == 0 {
		emu.SetX(0, 0)
	} else {
		emu.SetX(0, 1)
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubWctomb(emu *emulator.Emulator) bool {
	// int wctomb(char *s, wchar_t wc)
	s := emu.X(0)
	wc := uint32(emu.X(1))

	if s == 0 {
		emu.SetX(0, 0) // No state dependency
		stubs.ReturnFromStub(emu)
		return false
	}

	emu.MemWriteU8(s, byte(wc&0xff))
	emu.SetX(0, 1)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMblen(emu *emulator.Emulator) bool {
	// int mblen(const char *s, size_t n)
	s := emu.X(0)

	if s == 0 {
		emu.SetX(0, 0)
		stubs.ReturnFromStub(emu)
		return false
	}

	b, _ := emu.MemReadU8(s)
	if b == 0 {
		emu.SetX(0, 0)
	} else {
		emu.SetX(0, 1) // Simple ASCII - 1 byte per char
	}
	stubs.ReturnFromStub(emu)
	return false
}
