// Package lua provides stub implementations for Lua C API functions.
// These stubs allow emulation to continue when Cocos2d-x games use Lua scripting.
package lua

import (
	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

// Lua type constants
const (
	LUA_TNIL           = 0
	LUA_TBOOLEAN       = 1
	LUA_TLIGHTUSERDATA = 2
	LUA_TNUMBER        = 3
	LUA_TSTRING        = 4
	LUA_TTABLE         = 5
	LUA_TFUNCTION      = 6
	LUA_TUSERDATA      = 7
	LUA_TTHREAD        = 8
)

func init() {
	// Stack operations
	stubs.RegisterFunc("lua", "lua_settop", stubLuaSettop)
	stubs.RegisterFunc("lua", "lua_gettop", stubLuaGettop)
	stubs.RegisterFunc("lua", "lua_checkstack", stubLuaCheckstack)
	stubs.RegisterFunc("lua", "lua_pop", stubLuaNoop)
	stubs.RegisterFunc("lua", "lua_remove", stubLuaNoop)
	stubs.RegisterFunc("lua", "lua_insert", stubLuaNoop)
	stubs.RegisterFunc("lua", "lua_replace", stubLuaNoop)
	stubs.RegisterFunc("lua", "lua_copy", stubLuaNoop)

	// Push operations
	stubs.RegisterFunc("lua", "lua_pushnil", stubLuaPushnil)
	stubs.RegisterFunc("lua", "lua_pushnumber", stubLuaPushnumber)
	stubs.RegisterFunc("lua", "lua_pushinteger", stubLuaPushinteger)
	stubs.RegisterFunc("lua", "lua_pushstring", stubLuaPushstring)
	stubs.RegisterFunc("lua", "lua_pushlstring", stubLuaPushlstring)
	stubs.RegisterFunc("lua", "lua_pushboolean", stubLuaPushboolean)
	stubs.RegisterFunc("lua", "lua_pushvalue", stubLuaNoop)
	stubs.RegisterFunc("lua", "lua_pushlightuserdata", stubLuaNoop)
	stubs.RegisterFunc("lua", "lua_pushcclosure", stubLuaPushcclosure)
	stubs.RegisterFunc("lua", "lua_pushcfunction", stubLuaPushcclosure)

	// Get operations
	stubs.RegisterFunc("lua", "lua_gettable", stubLuaNoop)
	stubs.RegisterFunc("lua", "lua_getfield", stubLuaGetfield)
	stubs.RegisterFunc("lua", "lua_getglobal", stubLuaGetglobal)
	stubs.RegisterFunc("lua", "lua_rawget", stubLuaNoop)
	stubs.RegisterFunc("lua", "lua_rawgeti", stubLuaNoop)
	stubs.RegisterFunc("lua", "lua_getmetatable", stubLuaGetmetatable)

	// Set operations
	stubs.RegisterFunc("lua", "lua_settable", stubLuaNoop)
	stubs.RegisterFunc("lua", "lua_setfield", stubLuaSetfield)
	stubs.RegisterFunc("lua", "lua_setglobal", stubLuaSetglobal)
	stubs.RegisterFunc("lua", "lua_rawset", stubLuaRawset)
	stubs.RegisterFunc("lua", "lua_rawseti", stubLuaRawseti)
	stubs.RegisterFunc("lua", "lua_setmetatable", stubLuaSetmetatable)

	// Type checking
	stubs.RegisterFunc("lua", "lua_type", stubLuaType)
	stubs.RegisterFunc("lua", "lua_typename", stubLuaTypename)
	stubs.RegisterFunc("lua", "lua_isnil", stubLuaIsnil)
	stubs.RegisterFunc("lua", "lua_isboolean", stubLuaIsboolean)
	stubs.RegisterFunc("lua", "lua_isnumber", stubLuaIsnumber)
	stubs.RegisterFunc("lua", "lua_isstring", stubLuaIsstring)
	stubs.RegisterFunc("lua", "lua_istable", stubLuaIstable)
	stubs.RegisterFunc("lua", "lua_isfunction", stubLuaIsfunction)
	stubs.RegisterFunc("lua", "lua_iscfunction", stubLuaIscfunction)
	stubs.RegisterFunc("lua", "lua_isuserdata", stubLuaIsuserdata)

	// Conversion
	stubs.RegisterFunc("lua", "lua_tonumber", stubLuaTonumber)
	stubs.RegisterFunc("lua", "lua_tointeger", stubLuaTointeger)
	stubs.RegisterFunc("lua", "lua_toboolean", stubLuaToboolean)
	stubs.RegisterFunc("lua", "lua_tostring", stubLuaTostring)
	stubs.RegisterFunc("lua", "lua_tolstring", stubLuaTolstring)
	stubs.RegisterFunc("lua", "lua_touserdata", stubLuaTouserdata)
	stubs.RegisterFunc("lua", "lua_topointer", stubLuaTopointer)

	// Table operations
	stubs.RegisterFunc("lua", "lua_createtable", stubLuaCreatetable)
	stubs.RegisterFunc("lua", "lua_newtable", stubLuaNewtable)
	stubs.RegisterFunc("lua", "lua_newuserdata", stubLuaNewuserdata)
	stubs.RegisterFunc("lua", "lua_objlen", stubLuaObjlen)
	stubs.RegisterFunc("lua", "lua_next", stubLuaNext)

	// Comparison
	stubs.RegisterFunc("lua", "lua_equal", stubLuaEqual)
	stubs.RegisterFunc("lua", "lua_rawequal", stubLuaRawequal)
	stubs.RegisterFunc("lua", "lua_lessthan", stubLuaLessthan)

	// Call operations
	stubs.RegisterFunc("lua", "lua_call", stubLuaNoop)
	stubs.RegisterFunc("lua", "lua_pcall", stubLuaPcall)
	stubs.RegisterFunc("lua", "lua_cpcall", stubLuaCpcall)

	// Error handling
	stubs.RegisterFunc("lua", "lua_error", stubLuaError)
	stubs.RegisterFunc("lua", "luaL_error", stubLuaLError)

	// State management
	stubs.RegisterFunc("lua", "luaL_newstate", stubLuaLNewstate)
	stubs.RegisterFunc("lua", "lua_newstate", stubLuaNewstate)
	stubs.RegisterFunc("lua", "lua_close", stubLuaNoop)
	stubs.RegisterFunc("lua", "luaL_openlibs", stubLuaLOpenlibs)
	stubs.RegisterFunc("lua", "luaopen_base", stubLuaNoop)
	stubs.RegisterFunc("lua", "luaopen_table", stubLuaNoop)
	stubs.RegisterFunc("lua", "luaopen_string", stubLuaNoop)
	stubs.RegisterFunc("lua", "luaopen_math", stubLuaNoop)
	stubs.RegisterFunc("lua", "luaopen_io", stubLuaNoop)
	stubs.RegisterFunc("lua", "luaopen_os", stubLuaNoop)
	stubs.RegisterFunc("lua", "luaopen_debug", stubLuaNoop)
	stubs.RegisterFunc("lua", "luaopen_package", stubLuaNoop)

	// Auxiliary library
	stubs.RegisterFunc("lua", "luaL_register", stubLuaLRegister)
	stubs.RegisterFunc("lua", "luaL_getmetatable", stubLuaLGetmetatable)
	stubs.RegisterFunc("lua", "luaL_newmetatable", stubLuaLNewmetatable)
	stubs.RegisterFunc("lua", "luaL_checkudata", stubLuaLCheckudata)
	stubs.RegisterFunc("lua", "luaL_checknumber", stubLuaLChecknumber)
	stubs.RegisterFunc("lua", "luaL_checkinteger", stubLuaLCheckinteger)
	stubs.RegisterFunc("lua", "luaL_checkstring", stubLuaLCheckstring)
	stubs.RegisterFunc("lua", "luaL_checklstring", stubLuaLChecklstring)
	stubs.RegisterFunc("lua", "luaL_optstring", stubLuaLOptstring)
	stubs.RegisterFunc("lua", "luaL_optnumber", stubLuaLOptnumber)
	stubs.RegisterFunc("lua", "luaL_optinteger", stubLuaLOptinteger)
	stubs.RegisterFunc("lua", "luaL_ref", stubLuaLRef)
	stubs.RegisterFunc("lua", "luaL_unref", stubLuaNoop)
	stubs.RegisterFunc("lua", "luaL_loadfile", stubLuaLLoadfile)
	stubs.RegisterFunc("lua", "luaL_loadstring", stubLuaLLoadstring)
	stubs.RegisterFunc("lua", "luaL_loadbuffer", stubLuaLLoadbuffer)
	stubs.RegisterFunc("lua", "luaL_dofile", stubLuaLDofile)
	stubs.RegisterFunc("lua", "luaL_dostring", stubLuaLDostring)

	// GC
	stubs.RegisterFunc("lua", "lua_gc", stubLuaGc)
}

// Fake lua_State pointer
var luaStatePtr uint64

// stubLuaNoop is a no-op stub that just returns
func stubLuaNoop(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

// Stack operations

func stubLuaSettop(emu *emulator.Emulator) bool {
	// void lua_settop(lua_State *L, int index)
	// Just return
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaGettop(emu *emulator.Emulator) bool {
	// int lua_gettop(lua_State *L)
	emu.SetX(0, 0) // Return 0 (empty stack)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaCheckstack(emu *emulator.Emulator) bool {
	// int lua_checkstack(lua_State *L, int n)
	emu.SetX(0, 1) // Return 1 (success)
	stubs.ReturnFromStub(emu)
	return false
}

// Push operations

func stubLuaPushnil(emu *emulator.Emulator) bool {
	stubs.DefaultRegistry.Log("lua", "lua_pushnil", "")
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaPushnumber(emu *emulator.Emulator) bool {
	// void lua_pushnumber(lua_State *L, lua_Number n)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaPushinteger(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaPushstring(emu *emulator.Emulator) bool {
	// const char *lua_pushstring(lua_State *L, const char *s)
	sPtr := emu.X(1)
	if sPtr != 0 {
		s, _ := emu.MemReadString(sPtr, 256)
		if len(s) > 0 {
			stubs.DefaultRegistry.Log("lua", "lua_pushstring", s)
		}
	}
	emu.SetX(0, sPtr) // Return the string pointer
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaPushlstring(emu *emulator.Emulator) bool {
	sPtr := emu.X(1)
	emu.SetX(0, sPtr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaPushboolean(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaPushcclosure(emu *emulator.Emulator) bool {
	// void lua_pushcclosure(lua_State *L, lua_CFunction fn, int n)
	stubs.ReturnFromStub(emu)
	return false
}

// Get operations

func stubLuaGetfield(emu *emulator.Emulator) bool {
	// void lua_getfield(lua_State *L, int index, const char *k)
	kPtr := emu.X(2)
	if kPtr != 0 {
		k, _ := emu.MemReadString(kPtr, 128)
		if len(k) > 0 {
			stubs.DefaultRegistry.Log("lua", "lua_getfield", k)
		}
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaGetglobal(emu *emulator.Emulator) bool {
	namePtr := emu.X(1)
	if namePtr != 0 {
		name, _ := emu.MemReadString(namePtr, 128)
		if len(name) > 0 {
			stubs.DefaultRegistry.Log("lua", "lua_getglobal", name)
		}
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaGetmetatable(emu *emulator.Emulator) bool {
	emu.SetX(0, 0) // Return 0 (no metatable)
	stubs.ReturnFromStub(emu)
	return false
}

// Set operations

func stubLuaSetfield(emu *emulator.Emulator) bool {
	// void lua_setfield(lua_State *L, int index, const char *k)
	kPtr := emu.X(2)
	if kPtr != 0 {
		k, _ := emu.MemReadString(kPtr, 128)
		if len(k) > 0 {
			stubs.DefaultRegistry.Log("lua", "lua_setfield", k)
		}
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaSetglobal(emu *emulator.Emulator) bool {
	namePtr := emu.X(1)
	if namePtr != 0 {
		name, _ := emu.MemReadString(namePtr, 128)
		if len(name) > 0 {
			stubs.DefaultRegistry.Log("lua", "lua_setglobal", name)
		}
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaRawset(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaRawseti(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaSetmetatable(emu *emulator.Emulator) bool {
	emu.SetX(0, 1) // Return 1 (success)
	stubs.ReturnFromStub(emu)
	return false
}

// Type checking

func stubLuaType(emu *emulator.Emulator) bool {
	// int lua_type(lua_State *L, int index)
	emu.SetX(0, LUA_TNIL) // Return nil type
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaTypename(emu *emulator.Emulator) bool {
	// const char *lua_typename(lua_State *L, int tp)
	emu.SetX(0, 0) // Return NULL
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaIsnil(emu *emulator.Emulator) bool {
	emu.SetX(0, 1) // Return true
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaIsboolean(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaIsnumber(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaIsstring(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaIstable(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaIsfunction(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaIscfunction(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaIsuserdata(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

// Conversion

func stubLuaTonumber(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaTointeger(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaToboolean(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaTostring(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaTolstring(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaTouserdata(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaTopointer(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

// Table operations

func stubLuaCreatetable(emu *emulator.Emulator) bool {
	// void lua_createtable(lua_State *L, int narr, int nrec)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaNewtable(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaNewuserdata(emu *emulator.Emulator) bool {
	// void *lua_newuserdata(lua_State *L, size_t size)
	size := emu.X(1)
	if size == 0 {
		size = 64
	}
	ptr := emu.Malloc(size)
	emu.SetX(0, ptr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaObjlen(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaNext(emu *emulator.Emulator) bool {
	emu.SetX(0, 0) // Return 0 (no more elements)
	stubs.ReturnFromStub(emu)
	return false
}

// Comparison

func stubLuaEqual(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaRawequal(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLessthan(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

// Call operations

func stubLuaPcall(emu *emulator.Emulator) bool {
	// int lua_pcall(lua_State *L, int nargs, int nresults, int errfunc)
	emu.SetX(0, 0) // LUA_OK
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaCpcall(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

// Error handling

func stubLuaError(emu *emulator.Emulator) bool {
	stubs.DefaultRegistry.Log("lua", "lua_error", "error raised")
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLError(emu *emulator.Emulator) bool {
	stubs.DefaultRegistry.Log("lua", "luaL_error", "error raised")
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

// State management

func stubLuaLNewstate(emu *emulator.Emulator) bool {
	// lua_State *luaL_newstate(void)
	if luaStatePtr == 0 {
		luaStatePtr = emu.Malloc(256)
	}
	stubs.DefaultRegistry.Log("lua", "luaL_newstate", stubs.FormatHex(luaStatePtr))
	emu.SetX(0, luaStatePtr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaNewstate(emu *emulator.Emulator) bool {
	if luaStatePtr == 0 {
		luaStatePtr = emu.Malloc(256)
	}
	emu.SetX(0, luaStatePtr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLOpenlibs(emu *emulator.Emulator) bool {
	stubs.DefaultRegistry.Log("lua", "luaL_openlibs", "")
	stubs.ReturnFromStub(emu)
	return false
}

// Auxiliary library

func stubLuaLRegister(emu *emulator.Emulator) bool {
	// void luaL_register(lua_State *L, const char *libname, const luaL_Reg *l)
	libnamePtr := emu.X(1)
	if libnamePtr != 0 {
		libname, _ := emu.MemReadString(libnamePtr, 128)
		if len(libname) > 0 {
			stubs.DefaultRegistry.Log("lua", "luaL_register", libname)
		}
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLGetmetatable(emu *emulator.Emulator) bool {
	emu.SetX(0, LUA_TNIL)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLNewmetatable(emu *emulator.Emulator) bool {
	emu.SetX(0, 1) // Return 1 (new metatable created)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLCheckudata(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLChecknumber(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLCheckinteger(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLCheckstring(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLChecklstring(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLOptstring(emu *emulator.Emulator) bool {
	// Returns default if nil
	def := emu.X(2)
	emu.SetX(0, def)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLOptnumber(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLOptinteger(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLRef(emu *emulator.Emulator) bool {
	emu.SetX(0, 1) // Return reference ID 1
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLLoadfile(emu *emulator.Emulator) bool {
	emu.SetX(0, 0) // LUA_OK
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLLoadstring(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLLoadbuffer(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLDofile(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubLuaLDostring(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

// GC

func stubLuaGc(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}
