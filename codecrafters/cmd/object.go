package main

import "fmt"

type ObjectType int

const (
	Nil ObjectType = iota
	Bool
	Number
	String
	Function
	Class
	Instance
)

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
	funDecl *FunDecl
	closure *Environment
	isInit  bool
}

func (f *LoxFunction) Type() ObjectType { return Function }
func (f *LoxFunction) String() string   { return fmt.Sprintf("<fn %s>", f.funDecl.name) }

type LoxClass struct {
	name       string
	superclass *LoxClass
	methods    map[string]*LoxFunction
}

func (c *LoxClass) Type() ObjectType { return Class }
func (c *LoxClass) String() string   { return c.name }

type LoxInstance struct {
	loxClass LoxClass
	fields   map[string]Object
}

func (i *LoxInstance) Type() ObjectType { return Instance }
func (i *LoxInstance) String() string   { return i.loxClass.name + " instance" }

// Helper functions to extract objects
func IsNumber(obj Object) (float64, bool) {
	if n, ok := obj.(*LoxNumber); ok {
		return n.num, true
	}
	return 0, false
}

func IsString(obj Object) (string, bool) {
	if s, ok := obj.(*LoxString); ok {
		return s.str, true
	}
	return "", false
}

func IsBool(obj Object) (bool, bool) {
	if b, ok := obj.(*LoxBool); ok {
		return b.value, true
	}
	return false, false
}

func IsNil(obj Object) bool {
	_, ok := obj.(*LoxNil)
	return ok
}

func IsFunction(obj Object) (*LoxFunction, bool) {
	if f, ok := obj.(*LoxFunction); ok {
		return f, true
	}
	return nil, false
}

func IsClass(obj Object) (*LoxClass, bool) {
	if c, ok := obj.(*LoxClass); ok {
		return c, true
	}
	return nil, false
}

func IsInstance(obj Object) (*LoxInstance, bool) {
	if i, ok := obj.(*LoxInstance); ok {
		return i, true
	}
	return nil, false
}

// Only false and nil are falsy
func IsTruthy(obj Object) bool {
	switch val := obj.(type) {
	case *LoxNil:
		return false
	case *LoxBool:
		return val.value
	default:
		return true
	}
}
