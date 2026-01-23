# Galago

Galago does not emulate Android.

Galago executes ARM64 instructions.

- There is no operating system.
- No runtime.
- No platform reconstruction.

Only the minimum CPU state and memory required for the target logic to run is provided. Instructions execute until the value of interest is derived or execution can no longer continue.

If Android never becomes necessary, it never exists.

![Demo](demo/demo.gif)

## How It Works

Galago works by refusing to simulate what is not required.

Most emulators start by recreating a world.
An OS. A runtime. Services. APIs. Assumptions.

Galago starts by asking a narrower question:
what must exist for these instructions to execute.

Anything not required is never instantiated.

The executable is not treated as an Android app.
It is treated as a sequence of ARM64 instructions.

Execution space is reduced to:

- Registers.
- Memory.
- Control flow.
- Minimal state required to avoid a fault.

- No Java.
- No ART.
- No framework.
- No system services.

Those are not missing.
They were never admissible.

Instructions run until they can no longer continue or until the target value is derived.
If a secret is computed entirely in native code, the environment above it is irrelevant.
So it is removed.

This is why Galago is not platform emulation.
Platform emulation preserves context.
Galago eliminates context.

Android is metadata.
Metadata is not evidence.
So it does not participate.

The binary executes in a synthetic environment that exists only to satisfy hard execution constraints.
If a memory region is needed, it exists.
If a syscall is never reached, it does not.

There is no attempt to be faithful.
Only to be sufficient.

The result is collapse without reconstruction.

Instead of rebuilding Android to reach a value,
Galago strips execution down until only the path that can produce the value remains.

Secrets appear not because the environment was recreated,
but because everything that was unnecessary was cut away.

Only what must execute, executes.

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
