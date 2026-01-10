package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/arch/arm64/arm64asm"

	"github.com/spf13/cobra"
	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/hipaa"
	glog "github.com/zboralski/galago/internal/log"
	"github.com/zboralski/galago/internal/stubs"
	_ "github.com/zboralski/galago/internal/stubs/all"
	"github.com/zboralski/galago/internal/stubs/jni"
	"github.com/zboralski/galago/internal/stubs/setters"
	"github.com/zboralski/galago/internal/trace"
	"github.com/zboralski/galago/internal/ui/colorize"
)

var (
	verbose bool
	quiet   bool
	maxInsn int
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "galago [binary.so]",
		Short: "Extract encryption keys from ARM64 Android native libraries",
		Long: `Galago extracts encryption keys from ARM64 Android native libraries through controlled emulation.

The tool emulates ARM64 code using Unicorn Engine. It loads the ELF binary,
sets up minimal stubs for libc and system calls, then runs from an entry point
until it hits a key-setting function.

Static disassembly fails when keys are:
  - Built in registers via MOVK instructions
  - Decrypted at runtime through XOR routines
  - Stored in object fields populated by constructors

Galago runs the actual code. When execution reaches a setter function, it reads
the arguments and captures the value.

Examples:
  galago libcocos2djs.so              # Extract keys with colorized trace
  galago libcocos2djs.so -q           # Quiet mode - keys and stats only
  galago libcocos2djs.so -v           # Verbose debug output
  galago info libil2cpp.so            # Show binary info`,
		Args:                  cobra.MaximumNArgs(1),
		DisableFlagsInUseLine: true,
		RunE:                  runTrace,
	}

	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose debug output")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode (keys + stats only)")
	rootCmd.Flags().IntVarP(&maxInsn, "num", "n", 500, "max instructions to show")

	infoCmd := &cobra.Command{
		Use:   "info <binary.so>",
		Short: "Show binary information",
		Args:  cobra.ExactArgs(1),
		RunE:  showInfo,
	}
	rootCmd.AddCommand(infoCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type traceCollector struct {
	mu     sync.Mutex
	events []*trace.Event
}

func (tc *traceCollector) Add(e *trace.Event) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.events = append(tc.events, e)
}

func (tc *traceCollector) GetAndClear() []*trace.Event {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	events := tc.events
	tc.events = nil
	return events
}

type outputWriter struct {
	ch     chan string
	done   chan struct{}
	writer *bufio.Writer
}

func newOutputWriter() *outputWriter {
	w := &outputWriter{
		ch:     make(chan string, 2048),
		done:   make(chan struct{}),
		writer: bufio.NewWriterSize(os.Stdout, 64*1024),
	}
	go w.run()
	return w
}

func (w *outputWriter) run() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case line, ok := <-w.ch:
			if !ok {
				w.writer.Flush()
				close(w.done)
				return
			}
			w.writer.WriteString(line)
			w.writer.WriteByte('\n')
		case <-ticker.C:
			w.writer.Flush()
		}
	}
}

func (w *outputWriter) Write(line string) {
	select {
	case w.ch <- line:
	default:
	}
}

func (w *outputWriter) Close() {
	close(w.ch)
	<-w.done
}

func instructionTags(dis string) []string {
	upper := strings.ToUpper(dis)
	mnemonic := strings.Fields(upper)
	if len(mnemonic) == 0 {
		return nil
	}

	var tags []string
	switch mnemonic[0] {
	case "EOR":
		tags = append(tags, "#xor")
	case "EOR3":
		tags = append(tags, "#xor", "#neon")
	case "BL":
		tags = append(tags, "#call")
	case "BLR":
		tags = append(tags, "#call", "#br")
	case "BR":
		tags = append(tags, "#br")
	case "RET":
		tags = append(tags, "#ret")
	case "SVC":
		tags = append(tags, "#syscall")
	case "AESE", "AESD", "AESMC", "AESIMC":
		tags = append(tags, "#aes", "#crypto")
	case "SHA1C", "SHA1P", "SHA1M", "SHA1H", "SHA1SU0", "SHA1SU1":
		tags = append(tags, "#sha1", "#crypto")
	case "SHA256H", "SHA256H2", "SHA256SU0", "SHA256SU1":
		tags = append(tags, "#sha256", "#crypto")
	}

	if strings.Contains(dis, ".16B") || strings.Contains(dis, ".8B") ||
		strings.Contains(dis, ".4S") || strings.Contains(dis, ".2D") {
		if !containsTag(tags, "#neon") {
			tags = append(tags, "#neon")
		}
	}

	return tags
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func isBlockEnd(dis string) bool {
	upper := strings.ToUpper(dis)
	fields := strings.Fields(upper)
	if len(fields) == 0 {
		return false
	}
	switch fields[0] {
	case "RET", "BR", "B", "ERET":
		return true
	case "B.EQ", "B.NE", "B.LT", "B.LE", "B.GT", "B.GE",
		"B.HI", "B.HS", "B.LO", "B.LS", "B.MI", "B.PL",
		"B.VS", "B.VC", "B.AL", "B.NV":
		return true
	}
	if strings.HasPrefix(fields[0], "CBZ") || strings.HasPrefix(fields[0], "CBNZ") ||
		strings.HasPrefix(fields[0], "TBZ") || strings.HasPrefix(fields[0], "TBNZ") {
		return true
	}
	return false
}

func formatLine(addr uint64, code []byte, dis string, funcName string, events []*trace.Event) string {
	var b strings.Builder
	b.Grow(256)

	visibleLen := 0

	b.WriteString(colorize.Address(addr))
	b.WriteString("  ")
	visibleLen += 8 + 2

	if len(code) >= 4 {
		hexBytes := fmt.Sprintf("%02X%02X%02X%02X", code[3], code[2], code[1], code[0])
		b.WriteString(colorize.HexBytes(hexBytes))
		b.WriteString("  ")
		visibleLen += 8 + 2
	}

	b.WriteString(colorize.Instruction(dis))
	visibleLen += len(dis)

	const insnCol = 50
	for visibleLen < insnCol {
		b.WriteByte(' ')
		visibleLen++
	}

	var comments []string
	for _, e := range events {
		if e.Detail != "" {
			comments = append(comments, e.Detail)
		}
		for k, v := range e.Annotations {
			comments = append(comments, k+"="+v)
		}
	}

	var allTags []string
	insnTags := instructionTags(dis)
	allTags = append(allTags, insnTags...)
	for _, e := range events {
		allTags = append(allTags, e.Tags.Strings()...)
	}

	if len(comments) > 0 || len(allTags) > 0 {
		var commentParts []string
		if len(allTags) > 0 {
			commentParts = append(commentParts, strings.Join(allTags, " "))
		}
		if len(comments) > 0 {
			commentParts = append(commentParts, strings.Join(comments, ", "))
		}

		comment := "; " + strings.Join(commentParts, " ")
		b.WriteString(colorize.Comment(comment))
		visibleLen += len(comment)
		b.WriteString("  ")
		visibleLen += 2
	}

	var hasContent bool

	if funcName != "" {
		b.WriteString(colorize.FuncName(funcName))
		visibleLen += len(funcName)
		hasContent = true
	}

	for _, e := range events {
		if e.Name != "" {
			if hasContent {
				b.WriteByte(' ')
				visibleLen++
			}
			b.WriteString(colorize.FuncName(e.Name))
			visibleLen += len(e.Name)
			hasContent = true
		}
	}

	return b.String()
}

func printHeader(w *outputWriter, binary string, base, entry uint64, numImports, numSymbols, numHooks int, entryName string) {
	if cwd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(cwd, binary); err == nil && !strings.HasPrefix(rel, "..") {
			binary = rel
		}
	}

	w.Write("")
	w.Write(fmt.Sprintf("%s galago ─ ARM64 emulation trace analyzer", colorize.Header("▶")))
	w.Write(fmt.Sprintf("  %s %s", colorize.Detail("Loading:"), binary))
	w.Write(fmt.Sprintf("  %s %s  %s %s",
		colorize.Detail("Base:"), colorize.Address(base),
		colorize.Detail("Entry:"), colorize.Address(entry)))
	w.Write(fmt.Sprintf("  %s %s  %s %s  %s %s",
		colorize.Detail("Imports:"), colorize.FuncName(fmt.Sprintf("%d", numImports)),
		colorize.Detail("Symbols:"), colorize.FuncName(fmt.Sprintf("%d", numSymbols)),
		colorize.Detail("Hooks:"), colorize.FuncName(fmt.Sprintf("%d", numHooks))))
	if entryName != "" {
		w.Write(fmt.Sprintf("  %s %s", colorize.Detail("Entry point:"), colorize.FuncName(entryName)))
	}
	w.Write("")
}

func printKeys(keys []setters.CapturedKey) {
	if len(keys) == 0 {
		return
	}
	fmt.Println()
	eq := colorize.Detail("=")
	for i := 0; i < len(keys); i++ {
		k := keys[i]
		if k.KeyType == "xxtea" {
			if i+1 < len(keys) && keys[i+1].KeyType == "signature" {
				fmt.Printf("xxtea %s %s  signature %s %s\n",
					eq,
					colorize.String(fmt.Sprintf("%q", k.Value)),
					eq,
					colorize.String(fmt.Sprintf("%q", keys[i+1].Value)))
				i++
			} else {
				fmt.Printf("xxtea %s %s\n", eq, colorize.String(fmt.Sprintf("%q", k.Value)))
			}
		} else if k.KeyType == "signature" {
			fmt.Printf("signature %s %s\n", eq, colorize.String(fmt.Sprintf("%q", k.Value)))
		} else {
			fmt.Printf("%s %s %s\n", k.KeyType, eq, colorize.String(fmt.Sprintf("%q", k.Value)))
		}
	}
}

func printStats(count int, keys []setters.CapturedKey, err error) {
	fmt.Println()
	fmt.Print(colorize.Border("───────────────────────────────────────── "))
	fmt.Printf("%s insn  %s keys",
		colorize.FuncName(fmt.Sprintf("%d", count)),
		colorize.FuncName(fmt.Sprintf("%d", len(keys))))
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "UC_ERR_READ_UNMAPPED") || strings.Contains(errStr, "UC_ERR_WRITE_UNMAPPED") {
			fmt.Printf("  %s", colorize.Detail(errStr))
		} else {
			fmt.Printf("  %s", colorize.Error(errStr))
		}
	}
	fmt.Println()
}

func printQuietSummary(binary string, count, xorCount, retCount, brCount, stubCount, hookCount, hitCount int, keys []setters.CapturedKey) {
	name := filepath.Base(binary)
	fmt.Printf("%s\n", colorize.FuncName(name))

	eq := colorize.Detail("=")
	for i := 0; i < len(keys); i++ {
		k := keys[i]
		if k.KeyType == "xxtea" {
			if i+1 < len(keys) && keys[i+1].KeyType == "signature" {
				fmt.Printf("xxtea %s %s  signature %s %s\n",
					eq,
					colorize.String(fmt.Sprintf("%q", k.Value)),
					eq,
					colorize.String(fmt.Sprintf("%q", keys[i+1].Value)))
				i++
			} else {
				fmt.Printf("xxtea %s %s\n", eq, colorize.String(fmt.Sprintf("%q", k.Value)))
			}
		} else if k.KeyType == "signature" {
			fmt.Printf("signature %s %s\n", eq, colorize.String(fmt.Sprintf("%q", k.Value)))
		} else {
			fmt.Printf("%s %s %s\n", k.KeyType, eq, colorize.String(fmt.Sprintf("%q", k.Value)))
		}
	}

	fmt.Printf("%d %s", count, colorize.Detail("insn"))
	if hitCount > 0 {
		fmt.Printf("  %d %s", hitCount, colorize.Detail("hook"))
	}
	if stubCount > 0 {
		fmt.Printf("  %d %s", stubCount, colorize.Detail("stub"))
	}
	if retCount > 0 {
		fmt.Printf("  %d %s", retCount, colorize.Detail("ret"))
	}
	if brCount > 0 {
		fmt.Printf("  %d %s", brCount, colorize.Detail("br"))
	}
	if xorCount > 0 {
		fmt.Printf("  %d %s", xorCount, colorize.String("xor"))
	}
	fmt.Println()
	fmt.Println()
}

func disasm(code []byte) string {
	if len(code) < 4 {
		return "???"
	}
	inst, err := arm64asm.Decode(code)
	if err != nil {
		return fmt.Sprintf(".word 0x%08x", uint32(code[0])|uint32(code[1])<<8|uint32(code[2])<<16|uint32(code[3])<<24)
	}
	return inst.String()
}

func runTrace(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
	binaryPath := args[0]

	if verbose {
		glog.Init(true)
		stubs.Debug = true
	} else {
		glog.Init(false)
		stubs.Debug = false
	}

	emu, err := emulator.New()
	if err != nil {
		return fmt.Errorf("create emulator: %w", err)
	}

	info, err := emu.LoadELF(binaryPath)
	if err != nil {
		return fmt.Errorf("load ELF: %w", err)
	}

	setters.ClearCapturedKeys()

	installed := stubs.Install(emu, info.Imports, info.Symbols)

	hookHitCount := 0
	installVtableStubHooks(emu, info, &hookHitCount)

	collector := &traceCollector{}
	stubCallCount := 0
	stubs.DefaultRegistry.OnCall = func(category, name, detail string) {
		stubCallCount++
		e := trace.NewEvent(emu.PC(), category, name, detail)
		trace.DefaultEnricher(e)
		collector.Add(e)
	}

	entry := info.FindEntryPoint("")
	entryName := ""
	for name, addr := range info.Symbols {
		if addr == entry {
			entryName = name
			break
		}
	}
	if entryName == "" {
		entryName = "unknown"
	}

	if verbose {
		fmt.Printf("Entry point: 0x%x %s\n", entry, entryName)
	}

	javaVM := jni.GetJavaVM()
	mockObj := emu.GetMockObject()

	if strings.Contains(entryName, "cocos_android_app_init") {
		emu.SetX(0, javaVM)
		emu.SetX(1, mockObj)
	} else if strings.Contains(entryName, "lua_State") {
		emu.SetX(0, mockObj)
		emu.SetX(1, mockObj)
	} else {
		emu.SetX(0, mockObj)
		emu.SetX(1, mockObj)
	}

	sentinel := uint64(0xDEADBEEF)
	emu.SetLR(sentinel)
	emu.HookAddress(sentinel, func(e *emulator.Emulator) bool {
		return true
	})

	addrToSym := make(map[uint64]string, len(info.Symbols))
	for name, addr := range info.Symbols {
		if existing, ok := addrToSym[addr]; !ok || len(name) < len(existing) {
			addrToSym[addr] = name
		}
	}

	var out *outputWriter
	if !quiet {
		out = newOutputWriter()
	}

	if verbose {
		fmt.Printf("Loaded: %s\n", info.Path)
		fmt.Printf("Base: 0x%x, End: 0x%x\n", info.BaseAddr, info.EndAddr)
		fmt.Printf("Imports: %d, Symbols: %d\n", len(info.Imports), len(info.Symbols))
		fmt.Printf("Installed %d hooks\n", installed)
		fmt.Printf("Entry: 0x%x (%s)\n", entry, entryName)
		fmt.Println("\nStarting emulation...")
	} else if !quiet {
		printHeader(out, binaryPath, info.BaseAddr, entry, len(info.Imports), len(info.Symbols), installed, entryName)
	}

	count := 0
	xorCount := 0
	retCount := 0
	brCount := 0
	emu.HookCode(func(e *emulator.Emulator, addr uint64, size uint32) {
		count++
		if count > maxInsn {
			return
		}

		code, _ := e.MemRead(addr, 4)
		dis := disasm(code)
		events := collector.GetAndClear()
		funcName := addrToSym[addr]

		tags := instructionTags(dis)
		for _, tag := range tags {
			switch tag {
			case "#xor":
				xorCount++
			case "#ret":
				retCount++
			case "#br":
				brCount++
			}
		}

		if quiet {
			return
		}

		if verbose {
			fmt.Printf("  [%3d] 0x%08x  %s", count, addr, dis)
			if funcName != "" {
				fmt.Printf("  <%s>", funcName)
			}
			for _, ev := range events {
				fmt.Printf("  %s %s", ev.PrimaryTag(), ev.Name)
			}
			fmt.Println()
		} else {
			out.Write(formatLine(addr, code, dis, funcName, events))
			if isBlockEnd(dis) {
				out.Write("")
			}
		}
	})

	err = emu.RunFrom(entry)
	if out != nil {
		out.Close()
	}

	keys := setters.GetCapturedKeys()
	if verbose {
		fmt.Printf("\nEmulation finished: %v\n", err)
		fmt.Printf("Instructions: %d\n", count)
		fmt.Printf("\nRegisters: PC=0x%x LR=0x%x SP=0x%x\n", emu.PC(), emu.LR(), emu.SP())
		fmt.Printf("X0=0x%x X1=0x%x X2=0x%x X3=0x%x\n", emu.X(0), emu.X(1), emu.X(2), emu.X(3))

		if len(keys) > 0 {
			fmt.Println("\n=== CAPTURED KEYS ===")
			for _, k := range keys {
				fmt.Printf("  [%s] %s: %q (from %s @ 0x%x)\n", k.RiskLevel, k.KeyType, decryptKeyValue(k.Value), k.Source, k.Address)
			}
		} else {
			fmt.Println("\nNo keys captured")
		}
	} else if quiet {
		printQuietSummary(binaryPath, count, xorCount, retCount, brCount, stubCallCount, installed, hookHitCount, keys)
	} else {
		printKeys(keys)
		printStats(count, keys, err)
	}

	return nil
}

func installVtableStubHooks(emu *emulator.Emulator, info *emulator.ELFInfo, hitCount *int) {
	vtableBase := emu.GetVtableStubs()
	mockObj := emu.GetMockObject()

	setterPatterns := []string{
		"xxteakey",
		"cryptokey",
		"encryptionkey",
		"decryptionkey",
		"secretkey",
	}

	setterSlots := make(map[uint64]emulator.SlotInfo)
	if info != nil && info.VTables != nil {
		for _, tbl := range info.VTables.Tables {
			for _, slot := range tbl.Slots {
				if slot.SlotIndex >= 0 && isSetterSymbol(slot.SymName, setterPatterns) {
					setterSlots[uint64(slot.SlotIndex)] = slot
				}
			}
		}
	}

	for i := uint64(0); i < emulator.VtableStubCount; i++ {
		stubAddr := vtableBase + (i * 4)
		slotIdx := i

		emu.HookAddress(stubAddr, func(e *emulator.Emulator) bool {
			if hitCount != nil {
				*hitCount++
			}
			if slot, isSetter := setterSlots[slotIdx]; isSetter {
				x1 := e.X(1)
				x2 := e.X(2)
				x3 := e.X(3)
				x4 := e.X(4)

				if x2 > 0 && x2 < 256 && x1 > 0x1000 {
					if keyBytes, err := e.MemRead(x1, x2); err == nil {
						key := string(keyBytes)
						if isPrintableKey(key) {
							setters.CaptureKeyDirect(key, fmt.Sprintf("vtable[%d]->%s", slotIdx, slot.SymName), e.PC())
						}
					}
				}

				if x4 > 0 && x4 < 256 && x3 > 0x1000 {
					if signBytes, err := e.MemRead(x3, x4); err == nil {
						sign := string(signBytes)
						if isPrintableKey(sign) {
							setters.CaptureKeyDirect(sign, fmt.Sprintf("vtable[%d]->%s[signature]", slotIdx, slot.SymName), e.PC())
						}
					}
				}
			}

			e.SetX(0, mockObj)
			return false
		})
	}
}

func isSetterSymbol(symName string, patterns []string) bool {
	if symName == "" {
		return false
	}
	lower := strings.ToLower(symName)
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func decryptKeyValue(encrypted string) string {
	if hipaa.SessionEncryptor != nil {
		if decrypted, err := hipaa.SessionEncryptor.DecryptString(encrypted); err == nil {
			return decrypted
		}
	}
	return encrypted // Return as is if decryption fails
}

func showInfo(cmd *cobra.Command, args []string) error {
	binaryPath := args[0]

	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("file not found: %s", absPath)
	}

	emu, err := emulator.New()
	if err != nil {
		return fmt.Errorf("create emulator: %w", err)
	}

	elfInfo, err := emu.LoadELF(absPath)
	if err != nil {
		return fmt.Errorf("load binary: %w", err)
	}

	fmt.Printf("Binary: %s\n", filepath.Base(absPath))
	fmt.Printf("Base:   0x%x\n", elfInfo.BaseAddr)
	fmt.Printf("End:    0x%x\n", elfInfo.EndAddr)
	fmt.Printf("Entry:  0x%x\n", elfInfo.Entry)
	fmt.Printf("Symbols: %d\n\n", len(elfInfo.Symbols))

	fmt.Println("Key entry points:")
	entryPoint := elfInfo.FindEntryPoint("")
	if entryPoint != 0 {
		fmt.Printf("  Auto-detected: 0x%x\n", entryPoint)
	}

	if jniOnLoad := elfInfo.FindJNIOnLoad(); jniOnLoad != 0 {
		fmt.Printf("  JNI_OnLoad: 0x%x\n", jniOnLoad)
	}

	interesting := []string{
		"JNI_OnLoad",
		"il2cpp_init",
		"cocos_android_app_init",
		"setXXTeaKey",
		"setCryptoKey",
	}

	found := false
	for _, name := range interesting {
		syms := elfInfo.FindSymbolsBySubstring(name)
		for symName, addr := range syms {
			if !found {
				fmt.Println("\nInteresting symbols:")
				found = true
			}
			fmt.Printf("  0x%x %s\n", addr, symName)
		}
	}

	return nil
}
