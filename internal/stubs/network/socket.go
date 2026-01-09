// Package network provides stub implementations for network functions.
package network

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

var (
	nextFD   = 100 // Start from 100 to avoid conflicts with stdin/stdout/stderr
	fdMu     sync.Mutex
	socketFD = make(map[int]bool)
)

// CapturedHost represents a captured network host/IP.
type CapturedHost struct {
	IP       string
	Port     uint16
	Hostname string // From DNS resolution if available
	Source   string // "connect", "getaddrinfo", etc.
}

// capturedHosts stores all captured hosts during emulation.
var (
	capturedHosts   []CapturedHost
	capturedHostsMu sync.Mutex
)

// GetCapturedHosts returns all captured hosts.
func GetCapturedHosts() []CapturedHost {
	capturedHostsMu.Lock()
	defer capturedHostsMu.Unlock()
	result := make([]CapturedHost, len(capturedHosts))
	copy(result, capturedHosts)
	return result
}

// ClearCapturedHosts clears the captured hosts list.
func ClearCapturedHosts() {
	capturedHostsMu.Lock()
	capturedHosts = nil
	capturedHostsMu.Unlock()
}

// captureHost adds a host to the captured list.
func captureHost(ip string, port uint16, hostname, source string) {
	capturedHostsMu.Lock()
	capturedHosts = append(capturedHosts, CapturedHost{
		IP:       ip,
		Port:     port,
		Hostname: hostname,
		Source:   source,
	})
	capturedHostsMu.Unlock()
}

// parseSockaddrIn parses a sockaddr_in structure from memory.
// Returns IP string, port, and whether parsing succeeded.
func parseSockaddrIn(emu *emulator.Emulator, addrPtr uint64) (string, uint16, bool) {
	if addrPtr == 0 {
		return "", 0, false
	}

	// Read sockaddr_in:
	// struct sockaddr_in {
	//     sa_family_t sin_family;    // 2 bytes
	//     in_port_t sin_port;        // 2 bytes (network byte order)
	//     struct in_addr sin_addr;   // 4 bytes
	//     char sin_zero[8];          // padding
	// }
	data, err := emu.MemRead(addrPtr, 8)
	if err != nil || len(data) < 8 {
		return "", 0, false
	}

	family := binary.LittleEndian.Uint16(data[0:2])
	if family != 2 { // AF_INET
		return "", 0, false
	}

	// Port is in network byte order (big endian)
	port := binary.BigEndian.Uint16(data[2:4])

	// IP address
	ip := fmt.Sprintf("%d.%d.%d.%d", data[4], data[5], data[6], data[7])

	return ip, port, true
}

func init() {
	stubs.RegisterFunc("network", "socket", stubSocket)
	stubs.RegisterFunc("network", "connect", stubConnect)
	stubs.RegisterFunc("network", "bind", stubBind)
	stubs.RegisterFunc("network", "listen", stubListen)
	stubs.RegisterFunc("network", "accept", stubAccept)
	stubs.RegisterFunc("network", "send", stubSend)
	stubs.RegisterFunc("network", "recv", stubRecv)
	stubs.RegisterFunc("network", "sendto", stubSendto)
	stubs.RegisterFunc("network", "recvfrom", stubRecvfrom)
	stubs.RegisterFunc("network", "close", stubClose)
	stubs.RegisterFunc("network", "shutdown", stubShutdown)
	stubs.RegisterFunc("network", "setsockopt", stubSetsockopt)
	stubs.RegisterFunc("network", "getsockopt", stubGetsockopt)
	stubs.RegisterFunc("network", "fcntl", stubFcntl)
	stubs.RegisterFunc("network", "ioctl", stubIoctl)
	stubs.RegisterFunc("network", "select", stubSelect)
	stubs.RegisterFunc("network", "poll", stubPoll)
	stubs.RegisterFunc("network", "epoll_create", stubEpollCreate)
	stubs.RegisterFunc("network", "epoll_create1", stubEpollCreate)
	stubs.RegisterFunc("network", "epoll_ctl", stubEpollCtl)
	stubs.RegisterFunc("network", "epoll_wait", stubEpollWait)
}

func allocFD() int {
	fdMu.Lock()
	fd := nextFD
	nextFD++
	socketFD[fd] = true
	fdMu.Unlock()
	return fd
}

func stubSocket(emu *emulator.Emulator) bool {
	// int socket(int domain, int type, int protocol)
	fd := allocFD()
	stubs.DefaultRegistry.Log("network", "socket", stubs.FormatPtr("fd", uint64(fd)))
	emu.SetX(0, uint64(fd))
	stubs.ReturnFromStub(emu)
	return false
}

func stubConnect(emu *emulator.Emulator) bool {
	// int connect(int sockfd, const struct sockaddr *addr, socklen_t addrlen)
	addrPtr := emu.X(1)

	// Parse and capture the connection target
	if ip, port, ok := parseSockaddrIn(emu, addrPtr); ok {
		captureHost(ip, port, "", "connect")
		stubs.DefaultRegistry.Log("network", "connect", fmt.Sprintf("%s:%d", ip, port))
	}

	// Return success (we mock network operations)
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubBind(emu *emulator.Emulator) bool {
	// int bind(int sockfd, const struct sockaddr *addr, socklen_t addrlen)
	addrPtr := emu.X(1)

	if ip, port, ok := parseSockaddrIn(emu, addrPtr); ok {
		stubs.DefaultRegistry.Log("network", "bind", fmt.Sprintf("%s:%d", ip, port))
	}

	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubListen(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubAccept(emu *emulator.Emulator) bool {
	// Return new fake fd
	fd := allocFD()
	emu.SetX(0, uint64(fd))
	stubs.ReturnFromStub(emu)
	return false
}

func stubSend(emu *emulator.Emulator) bool {
	// ssize_t send(int sockfd, const void *buf, size_t len, int flags)
	length := emu.X(2)
	// Pretend we sent all the data
	emu.SetX(0, length)
	stubs.ReturnFromStub(emu)
	return false
}

func stubRecv(emu *emulator.Emulator) bool {
	// ssize_t recv(int sockfd, void *buf, size_t len, int flags)
	// Return 0 (connection closed) or -1 with EAGAIN
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubSendto(emu *emulator.Emulator) bool {
	// ssize_t sendto(int sockfd, const void *buf, size_t len, int flags,
	//                const struct sockaddr *dest_addr, socklen_t addrlen)
	length := emu.X(2)
	destAddrPtr := emu.X(4)

	if ip, port, ok := parseSockaddrIn(emu, destAddrPtr); ok {
		captureHost(ip, port, "", "sendto")
		stubs.DefaultRegistry.Log("network", "sendto", fmt.Sprintf("%s:%d", ip, port))
	}

	emu.SetX(0, length)
	stubs.ReturnFromStub(emu)
	return false
}

func stubRecvfrom(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubClose(emu *emulator.Emulator) bool {
	fd := int(emu.X(0))
	fdMu.Lock()
	delete(socketFD, fd)
	fdMu.Unlock()
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubShutdown(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubSetsockopt(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubGetsockopt(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubFcntl(emu *emulator.Emulator) bool {
	// Return 0 for most operations
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubIoctl(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubSelect(emu *emulator.Emulator) bool {
	// Return 0 (timeout, no fds ready)
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubPoll(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubEpollCreate(emu *emulator.Emulator) bool {
	fd := allocFD()
	emu.SetX(0, uint64(fd))
	stubs.ReturnFromStub(emu)
	return false
}

func stubEpollCtl(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubEpollWait(emu *emulator.Emulator) bool {
	// Return 0 events
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}
