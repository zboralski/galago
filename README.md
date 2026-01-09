# Galago

Galago extracts encryption keys from ARM64 Android native libraries through controlled emulation.

![Demo](demo/demo.gif)

## How It Works

The tool emulates ARM64 code using Unicorn Engine. It loads the ELF binary, sets up minimal stubs for libc and system calls, then runs from an entry point until it hits a key-setting function.

Static disassembly fails when keys are:
- Built in registers via MOVK instructions
- Decrypted at runtime through XOR routines
- Stored in object fields populated by constructors

Galago runs the actual code. When execution reaches a setter function, it reads the arguments and captures the value. The emulator treats crashes as valid termination. Once the key is captured, the job is done.

Most samples extract keys within 100-500 instructions.

## Install

```bash
brew install unicorn go
make setup
```

## Usage

```bash
# Extract keys with colorized trace
./galago libcocos2djs.so

# Quiet mode - keys and stats only
./galago -q libcocos2djs.so

# Batch process
ls samples/*.so | xargs -n1 ./galago -q

# Show binary info
./galago info libil2cpp.so
```

## Output

```
libcocos2dlua.so
xxtea = "%aoHg|#|LM"
158 insn  3 hook  3 stub  4 ret  3 br  10 xor
```

The trace shows instruction counts, hook hits, stub calls, and XOR operations. Keys appear as they are captured.

## Architecture

```
cmd/galago/          CLI entry point
internal/
  emulator/          Unicorn wrapper, ELF loader, memory management
  stubs/             Function stubs for libc, pthread, JNI, Lua
    setters/         Key capture hooks
  trace/             Execution event tracking
  ui/colorize/       Terminal output formatting
```

The emulator configures minimal infrastructure:
- Memory allocation via bump allocator
- Mock objects with vtable redirection
- Thread-local storage and stack canaries
- RTTI structures for C++ compatibility

## Name

The galago is a small primate that leaps between branches without touching the ground. Galago leaps through ARM64 code using [Unicorn](https://github.com/unicorn-engine/unicorn).

## License

MIT
