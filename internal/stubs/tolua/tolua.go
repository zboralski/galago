// Package tolua provides stub implementations for toLua++ binding functions.
// toLua++ is commonly used with Cocos2d-x to expose C++ classes to Lua.
package tolua

import (
	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

func init() {
	// Module management
	stubs.RegisterFunc("tolua", "tolua_open", stubToluaOpen)
	stubs.RegisterFunc("tolua", "tolua_module", stubToluaModule)
	stubs.RegisterFunc("tolua", "tolua_beginmodule", stubToluaBeginmodule)
	stubs.RegisterFunc("tolua", "tolua_endmodule", stubToluaNoop)

	// Class binding
	stubs.RegisterFunc("tolua", "tolua_cclass", stubToluaCclass)
	stubs.RegisterFunc("tolua", "tolua_function", stubToluaFunction)
	stubs.RegisterFunc("tolua", "tolua_constant", stubToluaConstant)
	stubs.RegisterFunc("tolua", "tolua_variable", stubToluaVariable)

	// User type management
	stubs.RegisterFunc("tolua", "tolua_usertype", stubToluaUsertype)
	stubs.RegisterFunc("tolua", "tolua_isusertable", stubToluaIsusertable)
	stubs.RegisterFunc("tolua", "tolua_isusertype", stubToluaIsusertype)

	// Value conversion
	stubs.RegisterFunc("tolua", "tolua_pushusertype", stubToluaPushusertype)
	stubs.RegisterFunc("tolua", "tolua_tousertype", stubToluaTousertype)
	stubs.RegisterFunc("tolua", "tolua_pushboolean", stubToluaNoop)
	stubs.RegisterFunc("tolua", "tolua_toboolean", stubToluaToboolean)
	stubs.RegisterFunc("tolua", "tolua_pushnumber", stubToluaNoop)
	stubs.RegisterFunc("tolua", "tolua_tonumber", stubToluaTonumber)
	stubs.RegisterFunc("tolua", "tolua_pushstring", stubToluaNoop)
	stubs.RegisterFunc("tolua", "tolua_tostring", stubToluaTostring)

	// Field access
	stubs.RegisterFunc("tolua", "tolua_pushfieldboolean", stubToluaNoop)
	stubs.RegisterFunc("tolua", "tolua_pushfieldnumber", stubToluaNoop)
	stubs.RegisterFunc("tolua", "tolua_pushfieldstring", stubToluaNoop)
	stubs.RegisterFunc("tolua", "tolua_pushfieldusertype", stubToluaNoop)

	// Type checking
	stubs.RegisterFunc("tolua", "tolua_isnoobj", stubToluaIsnoobj)
	stubs.RegisterFunc("tolua", "tolua_isboolean", stubToluaIsboolean)
	stubs.RegisterFunc("tolua", "tolua_isnumber", stubToluaIsnumber)
	stubs.RegisterFunc("tolua", "tolua_isstring", stubToluaIsstring)
	stubs.RegisterFunc("tolua", "tolua_istable", stubToluaIstable)

	// Error handling
	stubs.RegisterFunc("tolua", "tolua_error", stubToluaError)

	// Object management
	stubs.RegisterFunc("tolua", "tolua_register_gc", stubToluaNoop)
	stubs.RegisterFunc("tolua", "tolua_newmetatable", stubToluaNewmetatable)
	stubs.RegisterFunc("tolua", "tolua_getmetatable", stubToluaGetmetatable)

	// Cocos2d-x specific extensions
	stubs.RegisterFunc("tolua", "tolua_fix_function", stubToluaFixFunction)
	stubs.RegisterFunc("tolua", "toluafix_pushusertype_ccobject", stubToluafixPushusertypeCcobject)
}

// stubToluaNoop is a no-op stub
func stubToluaNoop(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

// Module management

func stubToluaOpen(emu *emulator.Emulator) bool {
	// void tolua_open(lua_State* L)
	stubs.DefaultRegistry.Log("tolua", "tolua_open", "")
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluaModule(emu *emulator.Emulator) bool {
	// void tolua_module(lua_State* L, const char* name, int hasvar)
	namePtr := emu.X(1)
	if namePtr != 0 {
		name, _ := emu.MemReadString(namePtr, 128)
		if len(name) > 0 {
			stubs.DefaultRegistry.Log("tolua", "tolua_module", name)
		}
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluaBeginmodule(emu *emulator.Emulator) bool {
	// void tolua_beginmodule(lua_State* L, const char* name)
	namePtr := emu.X(1)
	if namePtr != 0 {
		name, _ := emu.MemReadString(namePtr, 128)
		if len(name) > 0 {
			stubs.DefaultRegistry.Log("tolua", "tolua_beginmodule", name)
		}
	}
	stubs.ReturnFromStub(emu)
	return false
}

// Class binding

func stubToluaCclass(emu *emulator.Emulator) bool {
	// void tolua_cclass(lua_State* L, const char* lname, const char* name, const char* base, lua_CFunction col)
	lnamePtr := emu.X(1)
	namePtr := emu.X(2)
	if lnamePtr != 0 && namePtr != 0 {
		lname, _ := emu.MemReadString(lnamePtr, 128)
		name, _ := emu.MemReadString(namePtr, 128)
		if len(name) > 0 {
			stubs.DefaultRegistry.Log("tolua", "tolua_cclass", lname+" -> "+name)
		}
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluaFunction(emu *emulator.Emulator) bool {
	// void tolua_function(lua_State* L, const char* name, lua_CFunction func)
	namePtr := emu.X(1)
	if namePtr != 0 {
		name, _ := emu.MemReadString(namePtr, 128)
		if len(name) > 0 {
			stubs.DefaultRegistry.Log("tolua", "tolua_function", name)
		}
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluaConstant(emu *emulator.Emulator) bool {
	// void tolua_constant(lua_State* L, const char* name, lua_Number value)
	namePtr := emu.X(1)
	if namePtr != 0 {
		name, _ := emu.MemReadString(namePtr, 128)
		if len(name) > 0 {
			stubs.DefaultRegistry.Log("tolua", "tolua_constant", name)
		}
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluaVariable(emu *emulator.Emulator) bool {
	// void tolua_variable(lua_State* L, const char* name, lua_CFunction get, lua_CFunction set)
	namePtr := emu.X(1)
	if namePtr != 0 {
		name, _ := emu.MemReadString(namePtr, 128)
		if len(name) > 0 {
			stubs.DefaultRegistry.Log("tolua", "tolua_variable", name)
		}
	}
	stubs.ReturnFromStub(emu)
	return false
}

// User type management

func stubToluaUsertype(emu *emulator.Emulator) bool {
	// void tolua_usertype(lua_State* L, const char* type)
	typePtr := emu.X(1)
	if typePtr != 0 {
		typeName, _ := emu.MemReadString(typePtr, 128)
		if len(typeName) > 0 {
			stubs.DefaultRegistry.Log("tolua", "tolua_usertype", typeName)
		}
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluaIsusertable(emu *emulator.Emulator) bool {
	// int tolua_isusertable(lua_State* L, int lo, const char* type, int def, tolua_Error* err)
	emu.SetX(0, 1) // Return true
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluaIsusertype(emu *emulator.Emulator) bool {
	// int tolua_isusertype(lua_State* L, int lo, const char* type, int def, tolua_Error* err)
	emu.SetX(0, 1)
	stubs.ReturnFromStub(emu)
	return false
}

// Value conversion

func stubToluaPushusertype(emu *emulator.Emulator) bool {
	// void tolua_pushusertype(lua_State* L, void* value, const char* type)
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluaTousertype(emu *emulator.Emulator) bool {
	// void* tolua_tousertype(lua_State* L, int narg, void* def)
	def := emu.X(2)
	emu.SetX(0, def) // Return default
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluaToboolean(emu *emulator.Emulator) bool {
	// int tolua_toboolean(lua_State* L, int narg, int def)
	def := emu.X(2)
	emu.SetX(0, def)
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluaTonumber(emu *emulator.Emulator) bool {
	// lua_Number tolua_tonumber(lua_State* L, int narg, lua_Number def)
	def := emu.X(2)
	emu.SetX(0, def)
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluaTostring(emu *emulator.Emulator) bool {
	// const char* tolua_tostring(lua_State* L, int narg, const char* def)
	def := emu.X(2)
	emu.SetX(0, def)
	stubs.ReturnFromStub(emu)
	return false
}

// Type checking

func stubToluaIsnoobj(emu *emulator.Emulator) bool {
	// int tolua_isnoobj(lua_State* L, int lo, tolua_Error* err)
	emu.SetX(0, 1) // Return true (no object)
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluaIsboolean(emu *emulator.Emulator) bool {
	// int tolua_isboolean(lua_State* L, int lo, int def, tolua_Error* err)
	emu.SetX(0, 1)
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluaIsnumber(emu *emulator.Emulator) bool {
	// int tolua_isnumber(lua_State* L, int lo, int def, tolua_Error* err)
	emu.SetX(0, 1)
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluaIsstring(emu *emulator.Emulator) bool {
	// int tolua_isstring(lua_State* L, int lo, int def, tolua_Error* err)
	emu.SetX(0, 1)
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluaIstable(emu *emulator.Emulator) bool {
	// int tolua_istable(lua_State* L, int lo, int def, tolua_Error* err)
	emu.SetX(0, 1)
	stubs.ReturnFromStub(emu)
	return false
}

// Error handling

func stubToluaError(emu *emulator.Emulator) bool {
	// void tolua_error(lua_State* L, const char* msg, tolua_Error* err)
	msgPtr := emu.X(1)
	if msgPtr != 0 {
		msg, _ := emu.MemReadString(msgPtr, 256)
		if len(msg) > 0 {
			stubs.DefaultRegistry.Log("tolua", "tolua_error", msg)
		}
	}
	stubs.ReturnFromStub(emu)
	return false
}

// Object management

func stubToluaNewmetatable(emu *emulator.Emulator) bool {
	// int tolua_newmetatable(lua_State* L, const char* name)
	emu.SetX(0, 1) // Return 1 (new metatable)
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluaGetmetatable(emu *emulator.Emulator) bool {
	// void tolua_getmetatable(lua_State* L, const char* name)
	stubs.ReturnFromStub(emu)
	return false
}

// Cocos2d-x specific extensions

func stubToluaFixFunction(emu *emulator.Emulator) bool {
	// int tolua_fix_function(lua_State* L, int lo, int def)
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubToluafixPushusertypeCcobject(emu *emulator.Emulator) bool {
	// void toluafix_pushusertype_ccobject(lua_State* L, int refid, int* p_refid, void* ptr, const char* type)
	stubs.ReturnFromStub(emu)
	return false
}
