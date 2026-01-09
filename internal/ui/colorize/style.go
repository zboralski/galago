// Package colorize provides syntax highlighting for disassembly output.
// Shares the same color scheme as ~/re/reverse for consistency.
package colorize

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
)

func init() {
	// Register our custom disassembly style on package initialization
	_ = DisasmDark
}

// IDA-style theme colors
const (
	IDAAddress   = "#808080" // Gray for addresses
	IDAMnemonic  = "#FFFFFF" // White for mnemonics
	IDARegister  = "#87CEEB" // Light blue for registers
	IDANumber    = "#FF80C0" // Light pink for numbers
	IDALabel     = "#FFC800" // Yellow for labels/function names
	IDAComment   = "#FF8000" // Orange for comments
	IDAString    = "#00FF00" // Green for strings
	IDAHexBytes  = "#646464" // Dark gray for hex bytes
)

// DisasmDark is a custom style for disassembly - IDA Pro style
var DisasmDark = styles.Register(chroma.MustNewStyle("disasm-dark", chroma.StyleEntries{
	chroma.Text:           "#FFFFFF",    // White default
	chroma.Background:     "bg:#000000", // Pure black background
	chroma.Comment:        "#FF8000",    // Orange comments
	chroma.CommentPreproc: "#FF8000",    // Same for preprocessor comments

	// For NASM lexer mappings
	chroma.Keyword:       "#FFFFFF", // Instructions in white
	chroma.KeywordPseudo: "#FFFFFF", // Pseudo instructions in white
	chroma.Name:          "#87CEEB", // Generic names (registers) in cyan
	chroma.NameBuiltin:   "#87CEEB", // Builtin names (sp, lr) in cyan
	chroma.NameVariable:  "#87CEEB", // Variables/registers in cyan

	// Numbers - pink like IDA
	chroma.LiteralNumber:        "#FF80C0", // Decimal numbers in pink
	chroma.LiteralNumberHex:     "#FF80C0", // Hex numbers in pink
	chroma.LiteralNumberBin:     "#FF80C0", // Binary numbers in pink
	chroma.LiteralNumberOct:     "#FF80C0", // Octal numbers in pink
	chroma.LiteralNumberInteger: "#FF80C0", // Integer literals in pink
	chroma.LiteralNumberFloat:   "#FF80C0", // Float literals in pink

	// Labels and symbols
	chroma.NameLabel:    "#FFC800", // Labels in yellow
	chroma.NameFunction: "#FFFFFF", // Instructions as functions in white

	// Operators and punctuation
	chroma.Operator:    "#FFFFFF", // Operators in white
	chroma.Punctuation: "#FFFFFF", // Punctuation in white

	// Strings
	chroma.String: "#00FF00", // Strings in green
}))
