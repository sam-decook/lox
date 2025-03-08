package main

import "fmt"

type ObjectType int

const (
	Nil ObjectType = iota
	Bool
	Number
	String
	Function
)

// Object represents any object in Lox
type Object interface {
	Type() ObjectType
	String() string
}

type LoxNil struct{}

func (n *LoxNil) Type() ObjectType { return Nil }
func (n *LoxNil) String() string   { return "nil" }

type LoxBool struct {
	value bool
}

func (b *LoxBool) Type() ObjectType { return Bool }
func (b *LoxBool) String() string   { return fmt.Sprintf("%t", b.value) }

type LoxNumber struct {
	num float64
}

func (n *LoxNumber) Type() ObjectType { return Number }
func (n *LoxNumber) String() string   { return fmt.Sprintf("%.10g", n.num) }

type LoxString struct {
	str string
}

func (s *LoxString) Type() ObjectType { return String }
func (s *LoxString) String() string   { return s.str }

type LoxFunction struct {
	name string
	// TODO: this should be the other way around. A function should have the
	// parameter and body, and the FunDecl should point to it instead
	funDecl *FunDecl
	closure *Environment
}

func (f *LoxFunction) Type() ObjectType { return Function }
func (f *LoxFunction) String() string   { return fmt.Sprintf("<fn %s>", f.name) }
func (f *LoxFunction) Call(args []Object) (ret Object) {
	newEnv := NewEnvironment(f.closure)

	for i, arg := range args {
		newEnv.Define(f.funDecl.params[i].Lexeme, arg)
	}

	for _, stmt := range f.funDecl.body {
		retVal, ret := stmt.Run(&newEnv)
		if ret {
			return retVal
		}
	}
	return NewNil()
}

// Helper functions to create objects
// TODO: just use the struct literals, this isn't Java after all
func NewNumber(n float64) Object { return &LoxNumber{num: n} }
func NewString(s string) Object  { return &LoxString{str: s} }
func NewBool(b bool) Object      { return &LoxBool{value: b} }
func NewNil() Object             { return &LoxNil{} }

// Helper functions to extract objects
// TODO: split out concerns: check type and extract value
func IsNumber(v Object) (float64, bool) {
	if n, ok := v.(*LoxNumber); ok {
		return n.num, true
	}
	return 0, false
}

func IsString(v Object) (string, bool) {
	if s, ok := v.(*LoxString); ok {
		return s.str, true
	}
	return "", false
}

func IsBool(v Object) (bool, bool) {
	if b, ok := v.(*LoxBool); ok {
		return b.value, true
	}
	return false, false
}

func IsNil(v Object) bool {
	_, ok := v.(*LoxNil)
	return ok
}

func IsFunction(v Object) (*LoxFunction, bool) {
	if f, ok := v.(*LoxFunction); ok {
		return f, true
	}
	return nil, false
}

// Only false and nil are falsy
func IsTruthy(v Object) bool {
	switch val := v.(type) {
	case *LoxNil:
		return false
	case *LoxBool:
		return val.value
	default:
		return true
	}
}
