// Package internal provides hooks for internal functions that need mocking.
// These are TEXT symbols (not PLT imports) that require valid state to execute,
// such as Lua API, cocos2d singletons, and C++ RTTI operations.
package internal

import (
	"fmt"
	"strings"

	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

func init() {
	// Register internal function detector
	stubs.RegisterDetector(stubs.Detector{
		Name: "internal-mock",
		Patterns: []string{
			"lua_",
			"luaL_",
			"tolua_",
			"getInstance",
			"LuaEngine",
			"LuaStack",
			"ResourcesDecode",
		},
		Activate:    activateInternalMock,
		Description: "Internal function mocking for Lua/cocos2d",
	})
}

// activateInternalMock installs hooks for internal functions that need mocking.
func activateInternalMock(emu *emulator.Emulator, imports, symbols map[string]uint64) int {
	installed := 0

	for name, addr := range symbols {
		if addr == 0 {
			continue
		}

		// Skip PLT imports - they have their own stubs
		if _, isPLT := imports[name]; isPLT {
			continue
		}

		if shouldMockInternal(name) {
			behavior := inferReturnBehavior(name)
			emu.HookAddress(addr, makeMockHook(name, behavior))
			installed++
			// Debug log for RTTI functions
			if stubs.Debug && strings.Contains(strings.ToLower(name), "__do_") {
				stubs.DefaultRegistry.Log("internal", "rtti", fmt.Sprintf("%s @ 0x%x -> %s", name, addr, behavior))
			}
			// Debug log for _Map_base functions
			if stubs.Debug && strings.Contains(strings.ToLower(name), "_map_base") {
				stubs.DefaultRegistry.Log("internal", "map_base", fmt.Sprintf("%s @ 0x%x -> %s", name, addr, behavior))
			}
		}
	}

	if stubs.Debug {
		stubs.DefaultRegistry.Log("internal", "mock", fmt.Sprintf("%d internal functions installed (from %d symbols)", installed, len(symbols)))
	}
	return installed
}

// shouldMockInternal checks if an internal function should be mocked.
// Based on Python galago.py patterns.
func shouldMockInternal(name string) bool {
	lower := strings.ToLower(name)

	// IMPORTANT: Skip XXTEA/crypto key setter functions - these are handled by
	// the cocos2dx detector with proper key extraction hooks
	if strings.Contains(lower, "setxxteakey") || strings.Contains(lower, "setcryptokey") ||
		strings.Contains(lower, "xxteakeyandsign") {
		return false
	}

	// Mock getInstance-like functions that return singletons
	if strings.Contains(lower, "getinstance") {
		return true
	}

	// Mock start/init/create functions that do complex setup
	if strings.Contains(lower, "::start") || strings.Contains(lower, "::init") || strings.Contains(lower, "::create") {
		return true
	}

	// Mock ALL standard Lua C API functions when statically linked
	luaAPIFuncs := []string{
		// Stack manipulation
		"lua_gettop", "lua_settop", "lua_pushvalue", "lua_remove",
		"lua_insert", "lua_replace", "lua_checkstack", "lua_xmove",
		// Type checking
		"lua_type", "lua_typename", "lua_isnumber", "lua_isstring",
		"lua_iscfunction", "lua_isuserdata", "lua_isfunction",
		"lua_istable", "lua_isnil", "lua_isboolean", "lua_isthread",
		// Value access
		"lua_tonumber", "lua_tointeger", "lua_toboolean", "lua_tolstring",
		"lua_tostring", "lua_tocfunction", "lua_touserdata", "lua_tothread",
		"lua_topointer", "lua_objlen", "lua_rawlen",
		// Push functions
		"lua_pushnil", "lua_pushnumber", "lua_pushinteger", "lua_pushlstring",
		"lua_pushstring", "lua_pushcclosure", "lua_pushcfunction",
		"lua_pushboolean", "lua_pushlightuserdata", "lua_pushthread",
		"lua_pushvfstring", "lua_pushfstring",
		// Table functions
		"lua_gettable", "lua_getfield", "lua_rawget", "lua_rawgeti",
		"lua_createtable", "lua_newtable", "lua_newuserdata",
		"lua_settable", "lua_setfield", "lua_rawset", "lua_rawseti",
		"lua_setmetatable", "lua_getmetatable",
		// Global registry
		"lua_setglobal", "lua_getglobal", "lua_register",
		// Execution
		"lua_call", "lua_pcall", "lua_cpcall", "lua_load", "lua_dump",
		// Misc
		"lua_gc", "lua_error", "lua_next", "lua_concat", "lua_getallocf",
		"lua_setallocf", "lua_getupvalue", "lua_setupvalue",
		"lua_setlevel", "lua_atpanic", "lua_newthread", "lua_newstate",
		"lua_close", "lua_status",
		// luaL_ auxiliary functions
		"lual_newstate", "lual_openlibs", "lual_register", "lual_getmetafield",
		"lual_callmeta", "lual_typerror", "lual_argerror", "lual_checknumber",
		"lual_optnumber", "lual_checkinteger", "lual_optinteger",
		"lual_checkstring", "lual_optstring", "lual_checklstring",
		"lual_optlstring", "lual_checkudata", "lual_checktype",
		"lual_checkany", "lual_newmetatable", "lual_checkstack",
		"lual_loadfile", "lual_loadbuffer", "lual_loadstring",
		"lual_ref", "lual_unref", "lual_gsub", "lual_findtable",
		"lual_where", "lual_error", "lual_dofile", "lual_dostring",
	}
	for _, f := range luaAPIFuncs {
		if name == f || strings.HasPrefix(name, f+"@") || strings.Contains(lower, f) {
			return true
		}
	}

	// Mock Lua/cocos2d registration and setup functions
	mockPatterns := []string{
		"addsearchpath",
		"lua_module_register",
		"lua_register",
		"luaengine",
		"luahelper",
		"luaopen_",
		"luastack",
		"register_all_cocos2d",
		"register_all_cocos2dx",
		"register_custom",
		"register_hummer",
		"removescriptengine",
		"resourcesdecode",
		"restart_lua",
		"schedule",
		"scheduler::schedule",
		"scheduleupdate",
		"setanimationinterval",
		"setscriptengine",
		"shareddecode",
		"tolua_",
		"unschedule",
		"init_adjust",
		"init_appsflyer",
		"init_facebook",
		"initplatform",
		"initcrashreport",
		"crashreport",
		// STL container operations
		"_map_base",
		"_hashtable",
		"_select1st",
		"_prime_rehash",
		"getvaluemap",
		"getarchtype",
		"cocos2d::log",
		"cocos2d::value",
		"buglyluaagent",
		"pluginjnihelper",
		// Registration functions
		"register_bole",
		"register_all_pluginx",
		"package_quick_register",
		"loadchunksfromzip",
	}
	for _, p := range mockPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}

	// Mock C++ RTTI functions
	rttiPatterns := []string{
		"__do_catch",
		"__do_dyncast",
		"__do_upcast",
		"__is_pointer_p",
		"__is_function_p",
	}
	for _, p := range rttiPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}

	// Mock cocos2d::Application constructor
	if strings.Contains(lower, "cocos2d") && strings.Contains(lower, "application") &&
		(strings.Contains(lower, "c2ev") || strings.Contains(lower, "c1ev")) {
		return true
	}

	// Mock std::ctype constructor (accesses _ctype_ libc global which isn't initialized)
	// Symbol: _ZNSt5ctypeIcEC2EPKcbm = std::ctype<char>::ctype(char const*, bool, unsigned long)
	if strings.Contains(lower, "st5ctype") && (strings.Contains(lower, "c2e") || strings.Contains(lower, "c1e")) {
		return true
	}

	// Mock std::locale and related functions (access libc globals)
	if strings.Contains(lower, "st6locale") || strings.Contains(lower, "st7collate") ||
		strings.Contains(lower, "st7codecvt") || strings.Contains(lower, "st7num_get") ||
		strings.Contains(lower, "st7num_put") || strings.Contains(lower, "st8numpunct") ||
		strings.Contains(lower, "st8time_get") || strings.Contains(lower, "st8time_put") ||
		strings.Contains(lower, "st8messages") || strings.Contains(lower, "st9money_get") ||
		strings.Contains(lower, "st9money_put") || strings.Contains(lower, "st10moneypunct") {
		return true
	}

	// Mock std::_Rb_tree operations (used by std::set/map) - uninitialized global containers
	// These functions traverse the tree and crash if header node isn't initialized
	if strings.Contains(lower, "st8_rb_tree") {
		return true
	}

	// Mock cocos2d::experimental::FrameBuffer constructor
	// It inserts into global _frameBuffers set which is uninitialized
	if strings.Contains(lower, "framebuffer") && (strings.Contains(lower, "c1e") || strings.Contains(lower, "c2e")) {
		return true
	}

	// Mock cocos2d::AsyncTaskPool destructor - complex pthread/deque operations
	if strings.Contains(lower, "asynctaskpool") && (strings.Contains(lower, "d1e") || strings.Contains(lower, "d2e")) {
		return true
	}

	// Mock std::deque operations (complex internal state)
	if strings.Contains(lower, "st5deque") {
		return true
	}

	// Mock cocos2d::Ref::autorelease - calls into PoolManager which isn't initialized
	if strings.Contains(lower, "cocos2d") && strings.Contains(lower, "autorelease") {
		return true
	}

	// Mock cocos2d::PoolManager/AutoreleasePool - memory management singletons
	if strings.Contains(lower, "poolmanager") || strings.Contains(lower, "autoreleasepool") {
		return true
	}

	// Mock cocos2d::LuaEngine and ScriptEngineManager - complex script engine state
	if strings.Contains(lower, "luaengine") || strings.Contains(lower, "scriptenginemanager") ||
		strings.Contains(lower, "luastack") {
		return true
	}

	// Mock cocos2d::Director - complex singleton with lots of state
	if strings.Contains(lower, "cocos2d") && strings.Contains(lower, "director") {
		return true
	}

	// Mock CCGameMain - complex game initialization
	// Mock the lua_State* variant of applicationDidFinishLaunching (crashes on BSS string access)
	// but NOT the void variant of AppDelegate::applicationDidFinishLaunching (sets the key)
	if strings.Contains(lower, "ccgamemain") {
		// Always mock CCGameMain functions except the simple void-returning one
		// The lua_State* variant (_ZN7cocos2d10CCGameMain29applicationDidFinishLaunchingEP9lua_State)
		// crashes because it accesses uninitialized BSS strings
		if strings.Contains(lower, "lua_state") || !strings.Contains(lower, "applicationdidfinishlaunching") {
			return true
		}
	}

	// Mock lua_module_register - Lua binding setup
	if strings.Contains(lower, "lua_module_register") {
		return true
	}

	return false
}

// inferReturnBehavior determines how to mock a function based on naming conventions.
// Based on Python galago.py patterns.
func inferReturnBehavior(name string) string {
	lower := strings.ToLower(name)

	// Extract the method name (after last ::)
	method := lower
	if idx := strings.LastIndex(lower, "::"); idx != -1 {
		method = lower[idx+2:]
	}
	// Remove any trailing mangling or parentheses
	if idx := strings.Index(method, "("); idx != -1 {
		method = method[:idx]
	}

	// RTTI: C++ runtime type info functions - must return 0 (false)
	if strings.Contains(lower, "__do_upcast") || strings.Contains(lower, "__do_catch") ||
		strings.Contains(lower, "__do_dyncast") || strings.Contains(lower, "__is_pointer_p") ||
		strings.Contains(lower, "__is_function_p") {
		return "rtti"
	}

	// VALUE_REF: cocos2d::Value& returning functions (map operators)
	if (strings.Contains(lower, "_map_base") || strings.Contains(lower, "valuemap")) &&
		strings.Contains(lower, "value") {
		return "value_ref"
	}

	// STRING: Functions that return std::string by value (use x8 convention)
	stringPatterns := []string{
		"getpath", "getstring", "getname", "tostring", "getwritable",
		"fullpath", "getarch", "getkey", "getsign", "geturl", "geturi",
		"gettext", "getlabel", "gettitle", "getdescription",
	}
	for _, p := range stringPatterns {
		if strings.Contains(lower, p) {
			return "string"
		}
	}

	// VOID: Setters, initializers, handlers, callbacks, lifecycle methods
	voidPrefixes := []string{
		"set", "init", "start", "stop", "reset", "clear", "release",
		"destroy", "remove", "delete", "add", "insert", "push", "pop",
		"on", "handle", "process", "update", "visit", "draw", "render",
		"register", "unregister", "schedule", "unschedule", "cleanup",
		"load", "save", "write", "close", "open", "begin", "end",
		"enter", "exit", "pause", "resume", "retain", "autorelease",
	}
	voidExclusions := []string{"getset", "isset", "offset", "onset", "getinstance", "setget", "setup"}

	isVoidPrefix := false
	for _, prefix := range voidPrefixes {
		if strings.HasPrefix(method, prefix) {
			isVoidPrefix = true
			break
		}
	}
	isExcluded := false
	for _, excl := range voidExclusions {
		if strings.Contains(lower, excl) {
			isExcluded = true
			break
		}
	}
	if isVoidPrefix && !isExcluded {
		return "void"
	}

	// BOOL: Query methods (is*, has*, can*, should*, will*, did*, was*)
	boolPrefixes := []string{"is", "has", "can", "should", "will", "did", "was", "check", "valid"}
	for _, prefix := range boolPrefixes {
		if strings.HasPrefix(method, prefix) {
			return "bool"
		}
	}

	// INT: Count/size/length getters
	intPatterns := []string{"count", "size", "length", "index", "getint", "getcount",
		"getnumber", "getindex", "getsize", "getlength", "gettag"}
	for _, p := range intPatterns {
		if strings.Contains(lower, p) {
			return "int"
		}
	}

	// OBJECT: Default - getInstance, create, get* (object getters), new, clone, copy
	// These return pointers to objects
	return "object"
}

// makeMockHook creates a hook that mocks an internal function.
func makeMockHook(name, behavior string) func(*emulator.Emulator) bool {
	return func(emu *emulator.Emulator) bool {
		if stubs.Debug && strings.Contains(strings.ToLower(name), "_map_base") {
			stubs.DefaultRegistry.Log("internal", "map_base_HOOK", fmt.Sprintf("Hook fired for %s, behavior=%s, PC=0x%x, LR=0x%x", name, behavior, emu.PC(), emu.LR()))
		}
		switch behavior {
		case "rtti":
			// RTTI functions return 0 (false/no match)
			emu.SetX(0, 0)
		case "object":
			// Object getters return mock object pointer
			emu.SetX(0, emu.GetMockObject())
		case "value_ref":
			// Value& returns pointer to mock value object
			// cocos2d::Value layout (ARM64):
			//   offset 0: union _field (8 bytes) - actual value data
			//   offset 8: Type _type (4 bytes) - enum {NONE=0, BYTE, INT, UINT, FLOAT, DOUBLE, BOOLEAN, STRING=7, ...}
			// We initialize type to NONE (0) to avoid RTTI paths in asString() etc.
			valuePtr := emu.GetMockObject() + 0x100
			// Initialize the entire Value struct to zeros (type=NONE, field=0)
			emu.MemWrite(valuePtr, []byte{
				0, 0, 0, 0, 0, 0, 0, 0, // _field union (8 bytes)
				0, 0, 0, 0, // _type = NONE (0)
				0, 0, 0, 0, // padding
			})
			emu.SetX(0, valuePtr)
		case "bool":
			// Boolean functions return true (1)
			emu.SetX(0, 1)
		case "void":
			// Void functions don't set a return value
			// X0 is unchanged (often 'this' pointer)
		case "string":
			// String-returning functions use x8 for return value pointer
			// Write empty string to the buffer pointed to by X8
			x8 := emu.X(8)
			if x8 != 0 && x8 > 0x1000 && x8 < 0x7000000000000000 {
				// Write SSO-formatted empty string (length=0, no long flag)
				emu.MemWrite(x8, []byte{0, 0, 0, 0, 0, 0, 0, 0})
			}
			emu.SetX(0, emu.X(8)) // Return the string pointer
		case "int":
			// Integer-returning functions return 0
			emu.SetX(0, 0)
		default:
			// Default: return mock object (safer than 0)
			emu.SetX(0, emu.GetMockObject())
		}

		stubs.ReturnFromStub(emu)
		return false
	}
}
