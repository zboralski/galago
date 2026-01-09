package libc

import (
	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

func init() {
	stubs.RegisterFunc("libc", "strlen", stubStrlen)
	stubs.RegisterFunc("libc", "memcpy", stubMemcpy)
	stubs.RegisterFunc("libc", "memset", stubMemset)
	stubs.RegisterFunc("libc", "memmove", stubMemmove)
	stubs.RegisterFunc("libc", "memcmp", stubMemcmp)
	stubs.RegisterFunc("libc", "strcmp", stubStrcmp)
	stubs.RegisterFunc("libc", "strncmp", stubStrncmp)
	stubs.RegisterFunc("libc", "strcpy", stubStrcpy)
	stubs.RegisterFunc("libc", "strncpy", stubStrncpy)
	stubs.RegisterFunc("libc", "strcat", stubStrcat)
	stubs.RegisterFunc("libc", "strncat", stubStrncat)
	stubs.RegisterFunc("libc", "strchr", stubStrchr)
	stubs.RegisterFunc("libc", "strrchr", stubStrrchr)
	stubs.RegisterFunc("libc", "strstr", stubStrstr)
	stubs.RegisterFunc("libc", "strdup", stubStrdup)
	stubs.RegisterFunc("libc", "strndup", stubStrndup)
}

func stubStrlen(emu *emulator.Emulator) bool {
	addr := emu.X(0)
	str, _ := emu.MemReadString(addr, 4096)
	length := uint64(len(str))

	stubs.DefaultRegistry.Log("libc", "strlen", stubs.FormatPtr("len", length))
	emu.SetX(0, length)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMemcpy(emu *emulator.Emulator) bool {
	dest := emu.X(0)
	src := emu.X(1)
	n := emu.X(2)

	if n > 0 && n < 0x100000 {
		data, err := emu.MemRead(src, n)
		if err == nil {
			emu.MemWrite(dest, data)
		}
	}

	stubs.DefaultRegistry.Log("libc", "memcpy", formatMemop(dest, src, n))
	emu.SetX(0, dest)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMemset(emu *emulator.Emulator) bool {
	dest := emu.X(0)
	c := byte(emu.X(1) & 0xFF)
	n := emu.X(2)

	if n > 0 && n < 0x100000 {
		data := make([]byte, n)
		for i := range data {
			data[i] = c
		}
		emu.MemWrite(dest, data)
	}

	stubs.DefaultRegistry.Log("libc", "memset", stubs.FormatPtrPair("dest", dest, "c", uint64(c)))
	emu.SetX(0, dest)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMemmove(emu *emulator.Emulator) bool {
	dest := emu.X(0)
	src := emu.X(1)
	n := emu.X(2)

	if n > 0 && n < 0x100000 {
		data, err := emu.MemRead(src, n)
		if err == nil {
			emu.MemWrite(dest, data)
		}
	}

	stubs.DefaultRegistry.Log("libc", "memmove", formatMemop(dest, src, n))
	emu.SetX(0, dest)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMemcmp(emu *emulator.Emulator) bool {
	s1Addr := emu.X(0)
	s2Addr := emu.X(1)
	n := emu.X(2)

	var result uint64
	if n > 0 && n < 0x100000 {
		s1, _ := emu.MemRead(s1Addr, n)
		s2, _ := emu.MemRead(s2Addr, n)
		for i := uint64(0); i < n && i < uint64(len(s1)) && i < uint64(len(s2)); i++ {
			if s1[i] < s2[i] {
				result = 0xffffffffffffffff // -1
				break
			} else if s1[i] > s2[i] {
				result = 1
				break
			}
		}
	}

	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func stubStrcmp(emu *emulator.Emulator) bool {
	s1, _ := emu.MemReadString(emu.X(0), 256)
	s2, _ := emu.MemReadString(emu.X(1), 256)

	var result uint64
	if s1 < s2 {
		result = 0xffffffffffffffff
	} else if s1 > s2 {
		result = 1
	}

	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func stubStrncmp(emu *emulator.Emulator) bool {
	n := int(emu.X(2))
	s1, _ := emu.MemReadString(emu.X(0), n)
	s2, _ := emu.MemReadString(emu.X(1), n)

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
	}

	emu.SetX(0, result)
	stubs.ReturnFromStub(emu)
	return false
}

func stubStrcpy(emu *emulator.Emulator) bool {
	dest := emu.X(0)
	src := emu.X(1)
	str, _ := emu.MemReadString(src, 4096)
	emu.MemWriteString(dest, str)

	emu.SetX(0, dest)
	stubs.ReturnFromStub(emu)
	return false
}

func stubStrncpy(emu *emulator.Emulator) bool {
	dest := emu.X(0)
	src := emu.X(1)
	n := emu.X(2)

	str, _ := emu.MemReadString(src, int(n))
	if uint64(len(str)) < n {
		data := make([]byte, n)
		copy(data, str)
		emu.MemWrite(dest, data)
	} else {
		emu.MemWriteString(dest, str[:n])
	}

	emu.SetX(0, dest)
	stubs.ReturnFromStub(emu)
	return false
}

func stubStrcat(emu *emulator.Emulator) bool {
	dest := emu.X(0)
	src := emu.X(1)

	destStr, _ := emu.MemReadString(dest, 4096)
	srcStr, _ := emu.MemReadString(src, 4096)
	emu.MemWriteString(dest, destStr+srcStr)

	emu.SetX(0, dest)
	stubs.ReturnFromStub(emu)
	return false
}

func stubStrncat(emu *emulator.Emulator) bool {
	dest := emu.X(0)
	src := emu.X(1)
	n := int(emu.X(2))

	destStr, _ := emu.MemReadString(dest, 4096)
	srcStr, _ := emu.MemReadString(src, n)
	if len(srcStr) > n {
		srcStr = srcStr[:n]
	}
	emu.MemWriteString(dest, destStr+srcStr)

	emu.SetX(0, dest)
	stubs.ReturnFromStub(emu)
	return false
}

func stubStrchr(emu *emulator.Emulator) bool {
	addr := emu.X(0)
	c := byte(emu.X(1) & 0xFF)

	str, _ := emu.MemReadString(addr, 4096)
	for i := 0; i < len(str); i++ {
		if str[i] == c {
			emu.SetX(0, addr+uint64(i))
			stubs.ReturnFromStub(emu)
			return false
		}
	}
	// Also check for null terminator
	if c == 0 {
		emu.SetX(0, addr+uint64(len(str)))
		stubs.ReturnFromStub(emu)
		return false
	}

	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubStrrchr(emu *emulator.Emulator) bool {
	addr := emu.X(0)
	c := byte(emu.X(1) & 0xFF)

	str, _ := emu.MemReadString(addr, 4096)
	lastIdx := -1
	for i := 0; i < len(str); i++ {
		if str[i] == c {
			lastIdx = i
		}
	}
	if c == 0 {
		lastIdx = len(str)
	}

	if lastIdx >= 0 {
		emu.SetX(0, addr+uint64(lastIdx))
	} else {
		emu.SetX(0, 0)
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubStrstr(emu *emulator.Emulator) bool {
	haystackAddr := emu.X(0)
	needleAddr := emu.X(1)

	haystack, _ := emu.MemReadString(haystackAddr, 4096)
	needle, _ := emu.MemReadString(needleAddr, 256)

	if len(needle) == 0 {
		emu.SetX(0, haystackAddr)
		stubs.ReturnFromStub(emu)
		return false
	}

	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			emu.SetX(0, haystackAddr+uint64(i))
			stubs.ReturnFromStub(emu)
			return false
		}
	}

	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubStrdup(emu *emulator.Emulator) bool {
	src := emu.X(0)
	str, _ := emu.MemReadString(src, 4096)

	size := uint64(len(str) + 1)
	size = (size + 15) & ^uint64(15)
	ptr := emu.Malloc(size)
	emu.MemWriteString(ptr, str)

	emu.SetX(0, ptr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubStrndup(emu *emulator.Emulator) bool {
	src := emu.X(0)
	n := int(emu.X(1))
	str, _ := emu.MemReadString(src, n)
	if len(str) > n {
		str = str[:n]
	}

	size := uint64(len(str) + 1)
	size = (size + 15) & ^uint64(15)
	ptr := emu.Malloc(size)
	emu.MemWriteString(ptr, str)

	emu.SetX(0, ptr)
	stubs.ReturnFromStub(emu)
	return false
}

func formatMemop(dest, src, n uint64) string {
	return "dst=" + stubs.FormatHex(dest) + " src=" + stubs.FormatHex(src) + " n=" + stubs.FormatHex(n)
}
