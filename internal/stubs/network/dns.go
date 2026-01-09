package network

import (
	"fmt"

	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

func init() {
	stubs.RegisterFunc("network", "getaddrinfo", stubGetaddrinfo)
	stubs.RegisterFunc("network", "freeaddrinfo", stubFreeaddrinfo)
	stubs.RegisterFunc("network", "getnameinfo", stubGetnameinfo)
	stubs.RegisterFunc("network", "gethostbyname", stubGethostbyname)
	stubs.RegisterFunc("network", "gethostbyname2", stubGethostbyname)
	stubs.RegisterFunc("network", "gethostbyaddr", stubGethostbyaddr)
	stubs.RegisterFunc("network", "getpeername", stubGetpeername)
	stubs.RegisterFunc("network", "getsockname", stubGetsockname)
	stubs.RegisterFunc("network", "inet_aton", stubInetAton)
	stubs.RegisterFunc("network", "inet_ntoa", stubInetNtoa)
	stubs.RegisterFunc("network", "inet_pton", stubInetPton)
	stubs.RegisterFunc("network", "inet_ntop", stubInetNtop)
	stubs.RegisterFunc("network", "htons", stubHtons)
	stubs.RegisterFunc("network", "htonl", stubHtonl)
	stubs.RegisterFunc("network", "ntohs", stubNtohs)
	stubs.RegisterFunc("network", "ntohl", stubNtohl)
}

func stubGetaddrinfo(emu *emulator.Emulator) bool {
	nodePtr := emu.X(0)
	servicePtr := emu.X(1)
	// hintsPtr := emu.X(2)
	resPtr := emu.X(3)

	hostname := ""
	if nodePtr != 0 {
		hostname, _ = emu.MemReadString(nodePtr, 256)
	}

	service := ""
	if servicePtr != 0 {
		service, _ = emu.MemReadString(servicePtr, 32)
	}

	// Capture the hostname for later retrieval
	if hostname != "" {
		port := uint16(0)
		if service != "" {
			// Try to parse port from service string
			fmt.Sscanf(service, "%d", &port)
		}
		captureHost("127.0.0.1", port, hostname, "getaddrinfo")
		stubs.DefaultRegistry.Log("network", "getaddrinfo", fmt.Sprintf("host=%s service=%s", hostname, service))
	} else {
		stubs.DefaultRegistry.Log("network", "getaddrinfo", fmt.Sprintf("service=%s", service))
	}

	// Allocate a fake addrinfo structure
	// struct addrinfo {
	//     int ai_flags;           // 0
	//     int ai_family;          // 4
	//     int ai_socktype;        // 8
	//     int ai_protocol;        // 12
	//     socklen_t ai_addrlen;   // 16
	//     struct sockaddr *ai_addr;     // 24
	//     char *ai_canonname;     // 32
	//     struct addrinfo *ai_next;     // 40
	// }
	// Total: 48 bytes

	addrinfo := emu.Malloc(64)
	sockaddr := emu.Malloc(32) // struct sockaddr_in

	// Fill in addrinfo
	emu.MemWriteU32(addrinfo+0, 0)         // ai_flags
	emu.MemWriteU32(addrinfo+4, 2)         // ai_family = AF_INET
	emu.MemWriteU32(addrinfo+8, 1)         // ai_socktype = SOCK_STREAM
	emu.MemWriteU32(addrinfo+12, 0)        // ai_protocol
	emu.MemWriteU32(addrinfo+16, 16)       // ai_addrlen
	emu.MemWriteU64(addrinfo+24, sockaddr) // ai_addr
	emu.MemWriteU64(addrinfo+32, 0)        // ai_canonname = NULL
	emu.MemWriteU64(addrinfo+40, 0)        // ai_next = NULL

	// Fill in sockaddr_in with fake IP 127.0.0.1
	emu.MemWriteU16(sockaddr+0, 2)       // sin_family = AF_INET
	emu.MemWriteU16(sockaddr+2, 0x5000)  // sin_port = 80 (in network byte order)
	emu.MemWriteU32(sockaddr+4, 0x7f000001) // sin_addr = 127.0.0.1

	// Write result pointer
	emu.MemWriteU64(resPtr, addrinfo)

	emu.SetX(0, 0) // Success
	stubs.ReturnFromStub(emu)
	return false
}

func stubFreeaddrinfo(emu *emulator.Emulator) bool {
	// We don't actually free memory
	stubs.ReturnFromStub(emu)
	return false
}

func stubGetnameinfo(emu *emulator.Emulator) bool {
	// Just return error
	emu.SetX(0, 1) // EAI_AGAIN or similar
	stubs.ReturnFromStub(emu)
	return false
}

func stubGethostbyname(emu *emulator.Emulator) bool {
	namePtr := emu.X(0)
	name, _ := emu.MemReadString(namePtr, 256)

	// Capture hostname
	if name != "" {
		captureHost("127.0.0.1", 0, name, "gethostbyname")
	}
	stubs.DefaultRegistry.Log("network", "gethostbyname", name)

	// Allocate and fill struct hostent
	// struct hostent {
	//     char *h_name;        // 0
	//     char **h_aliases;    // 8
	//     int h_addrtype;      // 16
	//     int h_length;        // 20
	//     char **h_addr_list;  // 24
	// }

	hostent := emu.Malloc(64)
	addrList := emu.Malloc(16)
	addr := emu.Malloc(4)

	// Write 127.0.0.1
	emu.MemWrite(addr, []byte{127, 0, 0, 1})

	// addr_list[0] = addr, addr_list[1] = NULL
	emu.MemWriteU64(addrList, addr)
	emu.MemWriteU64(addrList+8, 0)

	// Fill hostent
	emu.MemWriteU64(hostent+0, namePtr)    // h_name
	emu.MemWriteU64(hostent+8, 0)          // h_aliases = NULL
	emu.MemWriteU32(hostent+16, 2)         // h_addrtype = AF_INET
	emu.MemWriteU32(hostent+20, 4)         // h_length
	emu.MemWriteU64(hostent+24, addrList)  // h_addr_list

	emu.SetX(0, hostent)
	stubs.ReturnFromStub(emu)
	return false
}

func stubGethostbyaddr(emu *emulator.Emulator) bool {
	// Return NULL (not found)
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubGetpeername(emu *emulator.Emulator) bool {
	// fd := emu.X(0)
	addrPtr := emu.X(1)
	lenPtr := emu.X(2)

	if addrPtr != 0 {
		// Fill with fake address
		emu.MemWriteU16(addrPtr, 2) // AF_INET
		emu.MemWriteU16(addrPtr+2, 0x5000)
		emu.MemWriteU32(addrPtr+4, 0x7f000001)
	}
	if lenPtr != 0 {
		emu.MemWriteU32(lenPtr, 16)
	}

	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubGetsockname(emu *emulator.Emulator) bool {
	return stubGetpeername(emu)
}

func stubInetAton(emu *emulator.Emulator) bool {
	// cpPtr := emu.X(0)
	inpPtr := emu.X(1)
	// Write 127.0.0.1 to inp
	if inpPtr != 0 {
		emu.MemWriteU32(inpPtr, 0x0100007f) // 127.0.0.1 in network byte order
	}
	emu.SetX(0, 1) // Success
	stubs.ReturnFromStub(emu)
	return false
}

func stubInetNtoa(emu *emulator.Emulator) bool {
	// Allocate and return "127.0.0.1"
	buf := emu.Malloc(16)
	emu.MemWriteString(buf, "127.0.0.1")
	emu.SetX(0, buf)
	stubs.ReturnFromStub(emu)
	return false
}

func stubInetPton(emu *emulator.Emulator) bool {
	// af := emu.X(0)
	// srcPtr := emu.X(1)
	dstPtr := emu.X(2)
	if dstPtr != 0 {
		emu.MemWriteU32(dstPtr, 0x0100007f)
	}
	emu.SetX(0, 1)
	stubs.ReturnFromStub(emu)
	return false
}

func stubInetNtop(emu *emulator.Emulator) bool {
	// af := emu.X(0)
	// srcPtr := emu.X(1)
	dstPtr := emu.X(2)
	// size := emu.X(3)
	if dstPtr != 0 {
		emu.MemWriteString(dstPtr, "127.0.0.1")
	}
	emu.SetX(0, dstPtr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubHtons(emu *emulator.Emulator) bool {
	val := uint16(emu.X(0))
	// Swap bytes
	result := (val >> 8) | (val << 8)
	emu.SetX(0, uint64(result))
	stubs.ReturnFromStub(emu)
	return false
}

func stubHtonl(emu *emulator.Emulator) bool {
	val := uint32(emu.X(0))
	result := ((val >> 24) & 0xFF) | ((val >> 8) & 0xFF00) |
		((val << 8) & 0xFF0000) | ((val << 24) & 0xFF000000)
	emu.SetX(0, uint64(result))
	stubs.ReturnFromStub(emu)
	return false
}

func stubNtohs(emu *emulator.Emulator) bool {
	return stubHtons(emu) // Same operation
}

func stubNtohl(emu *emulator.Emulator) bool {
	return stubHtonl(emu) // Same operation
}
