# Galago

Galago does not emulate Android.

It emulates ARM64 execution, not an operating system.

Only the minimal CPU and memory state required to execute the target logic is provided. No ART, no Java, no system services. Just instructions running until the secret appears.

![Demo](demo/demo.gif)

It runs:
- ARM64 instructions
- inside a controlled, synthetic execution environment
- with just enough memory and state to let the binary reach the point where secrets are derived

Think of it as instruction-level execution, not platform emulation.

The goal is not to recreate the Android environment.
The goal is to let the binary execute far enough to reveal runtime values.

If a value is computed purely in native code, Galago can observe it without Android ever existing.

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


## Name

The galago is a small primate that leaps between branches without touching the ground. Galago leaps through ARM64 code using [Unicorn](https://github.com/unicorn-engine/unicorn).

## License

MIT
