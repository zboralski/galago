// Package libc provides stub implementations for libc functions.
package libc

import (
	"sync"

	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

var (
	// File descriptor management
	nextFileFD   = 10 // Start from 10 to avoid stdin/stdout/stderr
	fileFDMu     sync.Mutex
	openFiles    = make(map[int]string) // fd -> path
	filePosition = make(map[int]int64)  // fd -> current position
)

func init() {
	// Basic file operations
	stubs.RegisterFunc("libc", "open", stubOpen)
	stubs.RegisterFunc("libc", "open64", stubOpen)
	stubs.RegisterFunc("libc", "openat", stubOpenat)
	stubs.RegisterFunc("libc", "openat64", stubOpenat)
	stubs.RegisterFunc("libc", "creat", stubCreat)
	stubs.RegisterFunc("libc", "creat64", stubCreat)

	// Read/write
	stubs.RegisterFunc("libc", "read", stubRead)
	stubs.RegisterFunc("libc", "write", stubWrite)
	stubs.RegisterFunc("libc", "pread", stubPread)
	stubs.RegisterFunc("libc", "pread64", stubPread)
	stubs.RegisterFunc("libc", "pwrite", stubPwrite)
	stubs.RegisterFunc("libc", "pwrite64", stubPwrite)
	stubs.RegisterFunc("libc", "readv", stubReadv)
	stubs.RegisterFunc("libc", "writev", stubWritev)

	// Seek/position
	stubs.RegisterFunc("libc", "lseek", stubLseek)
	stubs.RegisterFunc("libc", "lseek64", stubLseek)

	// File status
	stubs.RegisterFunc("libc", "stat", stubStat)
	stubs.RegisterFunc("libc", "stat64", stubStat)
	stubs.RegisterFunc("libc", "lstat", stubLstat)
	stubs.RegisterFunc("libc", "lstat64", stubLstat)
	stubs.RegisterFunc("libc", "fstat", stubFstat)
	stubs.RegisterFunc("libc", "fstat64", stubFstat)
	stubs.RegisterFunc("libc", "fstatat", stubFstatat)
	stubs.RegisterFunc("libc", "fstatat64", stubFstatat)
	stubs.RegisterFunc("libc", "access", stubAccess)
	stubs.RegisterFunc("libc", "faccessat", stubFaccessat)

	// File control
	stubs.RegisterFunc("libc", "dup", stubDup)
	stubs.RegisterFunc("libc", "dup2", stubDup2)
	stubs.RegisterFunc("libc", "dup3", stubDup3)
	stubs.RegisterFunc("libc", "pipe", stubPipe)
	stubs.RegisterFunc("libc", "pipe2", stubPipe2)

	// Memory mapping
	stubs.RegisterFunc("libc", "mmap", stubMmap)
	stubs.RegisterFunc("libc", "mmap64", stubMmap)
	stubs.RegisterFunc("libc", "munmap", stubMunmap)
	stubs.RegisterFunc("libc", "mprotect", stubMprotect)
	stubs.RegisterFunc("libc", "msync", stubMsync)
	stubs.RegisterFunc("libc", "madvise", stubMadvise)

	// Directory operations
	stubs.RegisterFunc("libc", "mkdir", stubMkdir)
	stubs.RegisterFunc("libc", "mkdirat", stubMkdirat)
	stubs.RegisterFunc("libc", "rmdir", stubRmdir)
	stubs.RegisterFunc("libc", "getcwd", stubGetcwd)
	stubs.RegisterFunc("libc", "chdir", stubChdir)
	stubs.RegisterFunc("libc", "fchdir", stubFchdir)
	stubs.RegisterFunc("libc", "opendir", stubOpendir)
	stubs.RegisterFunc("libc", "fdopendir", stubFdopendir)
	stubs.RegisterFunc("libc", "readdir", stubReaddir)
	stubs.RegisterFunc("libc", "readdir_r", stubReaddirR)
	stubs.RegisterFunc("libc", "closedir", stubClosedir)
	stubs.RegisterFunc("libc", "rewinddir", stubRewinddir)

	// File manipulation
	stubs.RegisterFunc("libc", "rename", stubRename)
	stubs.RegisterFunc("libc", "renameat", stubRenameat)
	stubs.RegisterFunc("libc", "unlink", stubUnlink)
	stubs.RegisterFunc("libc", "unlinkat", stubUnlinkat)
	stubs.RegisterFunc("libc", "remove", stubRemove)
	stubs.RegisterFunc("libc", "link", stubLink)
	stubs.RegisterFunc("libc", "linkat", stubLinkat)
	stubs.RegisterFunc("libc", "symlink", stubSymlink)
	stubs.RegisterFunc("libc", "symlinkat", stubSymlinkat)
	stubs.RegisterFunc("libc", "readlink", stubReadlink)
	stubs.RegisterFunc("libc", "readlinkat", stubReadlinkat)

	// File permissions
	stubs.RegisterFunc("libc", "chmod", stubChmod)
	stubs.RegisterFunc("libc", "fchmod", stubFchmod)
	stubs.RegisterFunc("libc", "fchmodat", stubFchmodat)
	stubs.RegisterFunc("libc", "chown", stubChown)
	stubs.RegisterFunc("libc", "fchown", stubFchown)
	stubs.RegisterFunc("libc", "lchown", stubLchown)
	stubs.RegisterFunc("libc", "fchownat", stubFchownat)

	// File locking
	stubs.RegisterFunc("libc", "flock", stubFlock)
	stubs.RegisterFunc("libc", "lockf", stubLockf)
	stubs.RegisterFunc("libc", "fcntl", stubFcntlFile)

	// File truncation
	stubs.RegisterFunc("libc", "truncate", stubTruncate)
	stubs.RegisterFunc("libc", "truncate64", stubTruncate)
	stubs.RegisterFunc("libc", "ftruncate", stubFtruncate)
	stubs.RegisterFunc("libc", "ftruncate64", stubFtruncate)

	// Sync
	stubs.RegisterFunc("libc", "sync", stubSync)
	stubs.RegisterFunc("libc", "fsync", stubFsync)
	stubs.RegisterFunc("libc", "fdatasync", stubFdatasync)

	// Temp files
	stubs.RegisterFunc("libc", "mkstemp", stubMkstemp)
	stubs.RegisterFunc("libc", "mkdtemp", stubMkdtemp)
	stubs.RegisterFunc("libc", "tmpfile", stubTmpfile)
	stubs.RegisterFunc("libc", "tmpfile64", stubTmpfile)

	// Realpath
	stubs.RegisterFunc("libc", "realpath", stubRealpath)

	// umask
	stubs.RegisterFunc("libc", "umask", stubUmask)
}

func allocFileFD(path string) int {
	fileFDMu.Lock()
	fd := nextFileFD
	nextFileFD++
	openFiles[fd] = path
	filePosition[fd] = 0
	fileFDMu.Unlock()
	return fd
}

func freeFileFD(fd int) {
	fileFDMu.Lock()
	delete(openFiles, fd)
	delete(filePosition, fd)
	fileFDMu.Unlock()
}

func stubOpen(emu *emulator.Emulator) bool {
	// int open(const char *pathname, int flags, mode_t mode)
	pathPtr := emu.X(0)
	// flags := emu.X(1)

	path, _ := emu.MemReadString(pathPtr, 512)
	stubs.DefaultRegistry.Log("libc", "open", path)

	// Return a fake fd
	fd := allocFileFD(path)
	emu.SetX(0, uint64(fd))
	stubs.ReturnFromStub(emu)
	return false
}

func stubOpenat(emu *emulator.Emulator) bool {
	// int openat(int dirfd, const char *pathname, int flags, mode_t mode)
	// dirfd := emu.X(0)
	pathPtr := emu.X(1)

	path, _ := emu.MemReadString(pathPtr, 512)
	stubs.DefaultRegistry.Log("libc", "openat", path)

	fd := allocFileFD(path)
	emu.SetX(0, uint64(fd))
	stubs.ReturnFromStub(emu)
	return false
}

func stubCreat(emu *emulator.Emulator) bool {
	pathPtr := emu.X(0)
	path, _ := emu.MemReadString(pathPtr, 512)
	stubs.DefaultRegistry.Log("libc", "creat", path)

	fd := allocFileFD(path)
	emu.SetX(0, uint64(fd))
	stubs.ReturnFromStub(emu)
	return false
}

func stubRead(emu *emulator.Emulator) bool {
	// ssize_t read(int fd, void *buf, size_t count)
	// Return 0 (EOF) for most reads
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubWrite(emu *emulator.Emulator) bool {
	// ssize_t write(int fd, const void *buf, size_t count)
	count := emu.X(2)
	// Pretend we wrote everything
	emu.SetX(0, count)
	stubs.ReturnFromStub(emu)
	return false
}

func stubPread(emu *emulator.Emulator) bool {
	emu.SetX(0, 0) // EOF
	stubs.ReturnFromStub(emu)
	return false
}

func stubPwrite(emu *emulator.Emulator) bool {
	count := emu.X(2)
	emu.SetX(0, count)
	stubs.ReturnFromStub(emu)
	return false
}

func stubReadv(emu *emulator.Emulator) bool {
	emu.SetX(0, 0) // EOF
	stubs.ReturnFromStub(emu)
	return false
}

func stubWritev(emu *emulator.Emulator) bool {
	// Just return 0 bytes written
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLseek(emu *emulator.Emulator) bool {
	// off_t lseek(int fd, off_t offset, int whence)
	offset := emu.X(1)
	// Return the offset
	emu.SetX(0, offset)
	stubs.ReturnFromStub(emu)
	return false
}

func stubStat(emu *emulator.Emulator) bool {
	// int stat(const char *pathname, struct stat *statbuf)
	pathPtr := emu.X(0)
	statPtr := emu.X(1)

	path, _ := emu.MemReadString(pathPtr, 512)
	stubs.DefaultRegistry.Log("libc", "stat", path)

	// Fill in a minimal stat structure (zeros mostly)
	if statPtr != 0 {
		// struct stat is large (144 bytes on arm64)
		// Just zero it out and set st_mode to regular file
		for i := uint64(0); i < 144; i += 8 {
			emu.MemWriteU64(statPtr+i, 0)
		}
		// st_mode at offset 16 (0100644 = regular file, rw-r--r--)
		emu.MemWriteU32(statPtr+16, 0100644)
		// st_size at offset 48
		emu.MemWriteU64(statPtr+48, 0)
	}

	emu.SetX(0, 0) // Success
	stubs.ReturnFromStub(emu)
	return false
}

func stubLstat(emu *emulator.Emulator) bool {
	return stubStat(emu)
}

func stubFstat(emu *emulator.Emulator) bool {
	// int fstat(int fd, struct stat *statbuf)
	statPtr := emu.X(1)

	if statPtr != 0 {
		for i := uint64(0); i < 144; i += 8 {
			emu.MemWriteU64(statPtr+i, 0)
		}
		emu.MemWriteU32(statPtr+16, 0100644)
	}

	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFstatat(emu *emulator.Emulator) bool {
	// int fstatat(int dirfd, const char *pathname, struct stat *statbuf, int flags)
	pathPtr := emu.X(1)
	statPtr := emu.X(2)

	path, _ := emu.MemReadString(pathPtr, 512)
	stubs.DefaultRegistry.Log("libc", "fstatat", path)

	if statPtr != 0 {
		for i := uint64(0); i < 144; i += 8 {
			emu.MemWriteU64(statPtr+i, 0)
		}
		emu.MemWriteU32(statPtr+16, 0100644)
	}

	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubAccess(emu *emulator.Emulator) bool {
	// int access(const char *pathname, int mode)
	pathPtr := emu.X(0)
	path, _ := emu.MemReadString(pathPtr, 512)
	stubs.DefaultRegistry.Log("libc", "access", path)

	emu.SetX(0, 0) // Success (file exists and is accessible)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFaccessat(emu *emulator.Emulator) bool {
	pathPtr := emu.X(1)
	path, _ := emu.MemReadString(pathPtr, 512)
	stubs.DefaultRegistry.Log("libc", "faccessat", path)

	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubDup(emu *emulator.Emulator) bool {
	oldfd := int(emu.X(0))
	fileFDMu.Lock()
	path := openFiles[oldfd]
	fileFDMu.Unlock()

	newfd := allocFileFD(path)
	emu.SetX(0, uint64(newfd))
	stubs.ReturnFromStub(emu)
	return false
}

func stubDup2(emu *emulator.Emulator) bool {
	// int dup2(int oldfd, int newfd)
	newfd := emu.X(1)
	emu.SetX(0, newfd)
	stubs.ReturnFromStub(emu)
	return false
}

func stubDup3(emu *emulator.Emulator) bool {
	newfd := emu.X(1)
	emu.SetX(0, newfd)
	stubs.ReturnFromStub(emu)
	return false
}

func stubPipe(emu *emulator.Emulator) bool {
	// int pipe(int pipefd[2])
	pipePtr := emu.X(0)
	if pipePtr != 0 {
		fd1 := allocFileFD("pipe[0]")
		fd2 := allocFileFD("pipe[1]")
		emu.MemWriteU32(pipePtr, uint32(fd1))
		emu.MemWriteU32(pipePtr+4, uint32(fd2))
	}
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubPipe2(emu *emulator.Emulator) bool {
	return stubPipe(emu)
}

func stubMmap(emu *emulator.Emulator) bool {
	// void *mmap(void *addr, size_t length, int prot, int flags, int fd, off_t offset)
	length := emu.X(1)

	// Allocate memory and return pointer
	ptr := emu.Malloc(length)
	stubs.DefaultRegistry.Log("libc", "mmap", stubs.FormatPtrPair("ptr", ptr, "size", length))
	emu.SetX(0, ptr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMunmap(emu *emulator.Emulator) bool {
	// We don't actually free mmap'd memory in the emulator
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMprotect(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMsync(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMadvise(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMkdir(emu *emulator.Emulator) bool {
	pathPtr := emu.X(0)
	path, _ := emu.MemReadString(pathPtr, 512)
	stubs.DefaultRegistry.Log("libc", "mkdir", path)
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMkdirat(emu *emulator.Emulator) bool {
	pathPtr := emu.X(1)
	path, _ := emu.MemReadString(pathPtr, 512)
	stubs.DefaultRegistry.Log("libc", "mkdirat", path)
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubRmdir(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubGetcwd(emu *emulator.Emulator) bool {
	// char *getcwd(char *buf, size_t size)
	buf := emu.X(0)
	// size := emu.X(1)

	cwd := "/data/data/com.app"
	if buf != 0 {
		emu.MemWriteString(buf, cwd)
		emu.SetX(0, buf)
	} else {
		ptr := emu.Malloc(uint64(len(cwd) + 1))
		emu.MemWriteString(ptr, cwd)
		emu.SetX(0, ptr)
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubChdir(emu *emulator.Emulator) bool {
	pathPtr := emu.X(0)
	path, _ := emu.MemReadString(pathPtr, 512)
	stubs.DefaultRegistry.Log("libc", "chdir", path)
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFchdir(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubOpendir(emu *emulator.Emulator) bool {
	pathPtr := emu.X(0)
	path, _ := emu.MemReadString(pathPtr, 512)
	stubs.DefaultRegistry.Log("libc", "opendir", path)

	// Return a fake DIR pointer
	dir := emu.Malloc(64)
	emu.SetX(0, dir)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFdopendir(emu *emulator.Emulator) bool {
	dir := emu.Malloc(64)
	emu.SetX(0, dir)
	stubs.ReturnFromStub(emu)
	return false
}

func stubReaddir(emu *emulator.Emulator) bool {
	// Return NULL (no more entries)
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubReaddirR(emu *emulator.Emulator) bool {
	// int readdir_r(DIR *dirp, struct dirent *entry, struct dirent **result)
	resultPtr := emu.X(2)
	if resultPtr != 0 {
		emu.MemWriteU64(resultPtr, 0) // No more entries
	}
	emu.SetX(0, 0) // Success
	stubs.ReturnFromStub(emu)
	return false
}

func stubClosedir(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubRewinddir(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func stubRename(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubRenameat(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubUnlink(emu *emulator.Emulator) bool {
	pathPtr := emu.X(0)
	path, _ := emu.MemReadString(pathPtr, 512)
	stubs.DefaultRegistry.Log("libc", "unlink", path)
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubUnlinkat(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubRemove(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLink(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLinkat(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubSymlink(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubSymlinkat(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubReadlink(emu *emulator.Emulator) bool {
	pathPtr := emu.X(0)
	buf := emu.X(1)
	// bufSize := emu.X(2)

	path, _ := emu.MemReadString(pathPtr, 512)
	stubs.DefaultRegistry.Log("libc", "readlink", path)

	// Return the path itself as the link target
	if buf != 0 {
		emu.MemWriteString(buf, path)
	}
	emu.SetX(0, uint64(len(path)))
	stubs.ReturnFromStub(emu)
	return false
}

func stubReadlinkat(emu *emulator.Emulator) bool {
	pathPtr := emu.X(1)
	buf := emu.X(2)

	path, _ := emu.MemReadString(pathPtr, 512)
	if buf != 0 {
		emu.MemWriteString(buf, path)
	}
	emu.SetX(0, uint64(len(path)))
	stubs.ReturnFromStub(emu)
	return false
}

func stubChmod(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFchmod(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFchmodat(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubChown(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFchown(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLchown(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFchownat(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFlock(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLockf(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFcntlFile(emu *emulator.Emulator) bool {
	// int fcntl(int fd, int cmd, ...)
	cmd := emu.X(1)

	switch cmd {
	case 1: // F_GETFD
		emu.SetX(0, 0)
	case 2: // F_SETFD
		emu.SetX(0, 0)
	case 3: // F_GETFL
		emu.SetX(0, 0)
	case 4: // F_SETFL
		emu.SetX(0, 0)
	default:
		emu.SetX(0, 0)
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubTruncate(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFtruncate(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubSync(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func stubFsync(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFdatasync(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubMkstemp(emu *emulator.Emulator) bool {
	// int mkstemp(char *template)
	templatePtr := emu.X(0)

	// Modify template and return fd
	if templatePtr != 0 {
		// Replace XXXXXX with fake values
		emu.MemWriteString(templatePtr, "/tmp/tmp.123456")
	}
	fd := allocFileFD("/tmp/tmp.123456")
	emu.SetX(0, uint64(fd))
	stubs.ReturnFromStub(emu)
	return false
}

func stubMkdtemp(emu *emulator.Emulator) bool {
	templatePtr := emu.X(0)
	if templatePtr != 0 {
		emu.MemWriteString(templatePtr, "/tmp/tmp.123456")
	}
	emu.SetX(0, templatePtr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubTmpfile(emu *emulator.Emulator) bool {
	// Return fake FILE pointer
	ptr := emu.Malloc(256)
	emu.SetX(0, ptr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubRealpath(emu *emulator.Emulator) bool {
	// char *realpath(const char *path, char *resolved_path)
	pathPtr := emu.X(0)
	resolved := emu.X(1)

	path, _ := emu.MemReadString(pathPtr, 512)
	stubs.DefaultRegistry.Log("libc", "realpath", path)

	// Return the path as-is (simplified)
	if resolved != 0 {
		emu.MemWriteString(resolved, path)
		emu.SetX(0, resolved)
	} else {
		ptr := emu.Malloc(uint64(len(path) + 1))
		emu.MemWriteString(ptr, path)
		emu.SetX(0, ptr)
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubUmask(emu *emulator.Emulator) bool {
	// mode_t umask(mode_t mask)
	// Return old mask (022 is common default)
	emu.SetX(0, 022)
	stubs.ReturnFromStub(emu)
	return false
}
