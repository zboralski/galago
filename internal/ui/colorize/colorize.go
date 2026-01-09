package colorize

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// getAssemblyLexer returns an appropriate assembly lexer with fallbacks
func getAssemblyLexer() chroma.Lexer {
	candidates := []string{"armasm", "gas", "GAS", "Gas", "nasm"}
	for _, name := range candidates {
		if lexer := lexers.Get(name); lexer != nil {
			return lexer
		}
	}
	return nil
}

// getDisasmStyle returns the disassembly style with fallbacks
func getDisasmStyle() *chroma.Style {
	candidates := []string{"disasm-dark", "dracula", "monokai"}
	for _, name := range candidates {
		if style := styles.Get(name); style != nil {
			return style
		}
	}
	return styles.Fallback
}

// getTerminalFormatter returns an appropriate terminal formatter
func getTerminalFormatter() chroma.Formatter {
	candidates := []string{"terminal16m", "terminal256"}
	for _, name := range candidates {
		if formatter := formatters.Get(name); formatter != nil {
			return formatter
		}
	}
	return formatters.Fallback
}

// IsDisabled returns true if colors are disabled via environment
func IsDisabled() bool {
	return os.Getenv("GALAGO_NO_COLOR") != "" || os.Getenv("NO_COLOR") != ""
}

// Instruction colorizes an assembly instruction using Chroma
func Instruction(insn string) string {
	if IsDisabled() {
		return insn
	}

	lexer := lexers.Get("nasm")
	if lexer == nil {
		lexer = getAssemblyLexer()
		if lexer == nil {
			return insn
		}
	}

	_ = DisasmDark // Force registration
	style := getDisasmStyle()
	formatter := getTerminalFormatter()

	iterator, err := lexer.Tokenise(nil, insn)
	if err != nil {
		return insn
	}

	var buf strings.Builder
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return insn
	}

	return strings.TrimSuffix(buf.String(), "\n")
}

// Address formats an address in yellow
func Address(addr uint64) string {
	if IsDisabled() {
		return fmt.Sprintf("%08X", addr)
	}
	return fmt.Sprintf("\033[38;2;255;200;0m%08X\033[0m", addr)
}

// Tag formats a hashtag in light pink
func Tag(tag string) string {
	if IsDisabled() {
		return tag
	}
	return fmt.Sprintf("\033[38;2;255;180;200m%s\033[0m", tag)
}

// FuncName formats a function name in yellow (IDA style labels)
func FuncName(name string) string {
	if IsDisabled() {
		return name
	}
	return fmt.Sprintf("\033[38;2;255;200;0m%s\033[0m", name)
}

// Detail formats detail text in light gray
func Detail(detail string) string {
	if IsDisabled() {
		return detail
	}
	return fmt.Sprintf("\033[38;2;180;180;180m%s\033[0m", detail)
}

// Key formats a captured key in red (high visibility)
func Key(key string) string {
	if IsDisabled() {
		return key
	}
	return fmt.Sprintf("\033[38;2;255;80;80m%s\033[0m", key)
}

// Border formats border characters in dark gray
func Border(s string) string {
	if IsDisabled() {
		return s
	}
	return fmt.Sprintf("\033[38;2;80;80;80m%s\033[0m", s)
}

// Comment formats comments in white
func Comment(s string) string {
	if IsDisabled() {
		return s
	}
	return fmt.Sprintf("\033[38;2;255;255;255m%s\033[0m", s)
}

// Header formats header text in blue (IDA style)
func Header(s string) string {
	if IsDisabled() {
		return s
	}
	return fmt.Sprintf("\033[38;2;86;156;214m%s\033[0m", s)
}

// HexBytes formats hex opcode bytes in light gray
func HexBytes(s string) string {
	if IsDisabled() {
		return s
	}
	return fmt.Sprintf("\033[38;2;180;180;180m%s\033[0m", s)
}

// Error formats error messages in pink
func Error(s string) string {
	if IsDisabled() {
		return s
	}
	return fmt.Sprintf("\033[38;2;255;128;192m%s\033[0m", s)
}

// String formats string values in pink/magenta
func String(s string) string {
	if IsDisabled() {
		return s
	}
	return fmt.Sprintf("\033[38;2;255;128;192m%s\033[0m", s)
}
