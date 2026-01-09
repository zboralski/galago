# Galago Demo - XXTEA Key Extraction
# Run with: vhs demo.tape
#
# IMPORTANT: Use Wait /^> /+Screen for command completion, NOT Sleep.
# Sleep is only for brief pauses (display timing). Wait /^> /+Screen
# respects Wait /^> /Timeout and handles variable execution times.

Output demo.gif
Output demo.mp4

Set FontFamily "Monaco"
Set FontSize 16
Set Width 1200
Set Height 800
Set Framerate 60
Set Shell zsh

Type "# Galago executes Android ARM64 native libraries as raw code."
Sleep 1
Ctrl+U

Type "#It recovers secrets by emulating the instructions that compute them."
Sleep 1
Ctrl+U

Type "#No Android device. No Android emulator."
Sleep 1.5
Ctrl+U

Sleep 0.5

# Run galago - the header explains what it's doing
Type "./galago demo/libcocos2dlua.so"
Enter

Wait@3

# Brief pause before replay
Sleep 300ms

# Show slow motion explanation, then erase it
Type "# If we slow down we can see some XORs execute"
Sleep 1.5
Ctrl+U

# Run slow motion version (~5KB/sec = ~9s for 45KB output)
Type "./galago demo/libcocos2dlua.so | pv -qL 5000"
Enter

# Wait /^> / for prompt to return
Wait@10

# Pause before batch run
Sleep 0.5

# Run all samples in quiet mode
Type "# Now let's recover SAMPLE_COUNT samples"
Sleep 1500ms
Ctrl+U
Type "ls -r samples/PASS/*.so | xargs -n1 -x ./galago -q"
Enter 1

Wait@400 

# Final pause to show results
Sleep 2s

# Poetic ending
Type "# The tape moves. The head bites. The symbols turn."
Sleep 2.5
Ctrl+U

Type "# Every key was always there, enfolded, waiting."
Sleep 2.5
Ctrl+U

Type "# The machine completes its report by falling silent."
Sleep 10s
