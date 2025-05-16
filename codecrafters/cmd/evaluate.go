package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func (ae *AssignmentExpr) Evaluate(lox *Interpreter) Object {
	obj := ae.expr.Evaluate(lox)

	distance, isLocal := lox.locals[ae]
	if isLocal {
		lox.AssignAt(distance, ae.name, obj)
	} else {
		lox.globals.Assign(ae.name, obj)
	}
	return obj
}

func (se *SetExpr) Evaluate(lox *Interpreter) Object {
	obj := se.object.Evaluate(lox)
	inst, ok := IsInstance(obj)
	if !ok {
		runtimeError("Only instances have fields.")
	}

	val := se.value.Evaluate(lox)
	inst.Set(se.name, val)
	return val
}

// The logical operators return a value of the proper truthiness
func (loe *LogicOrExpr) Evaluate(lox *Interpreter) Object {
	left := loe.left.Evaluate(lox)
	if IsTruthy(left) {
		// Short-circuit
		return left
	}
	return loe.right.Evaluate(lox)
}

// The logical operators return a value of the proper truthiness
func (lae *LogicAndExpr) Evaluate(lox *Interpreter) Object {
	left := lae.left.Evaluate(lox)
	if !IsTruthy(left) {
		// Short-circuit
		return left
	}
	return lae.right.Evaluate(lox)
}

func (ue *UnaryExpr) Evaluate(lox *Interpreter) Object {
	right := ue.right.Evaluate(lox)

	switch ue.op.Type {
	case BANG:
		return &LoxBool{!IsTruthy(right)}
	case MINUS:
		n := assertNumber(right)
		return &LoxNumber{-n}
	}
	panic("unreachable: UnaryExpression.Evaluate(lox)")
}

func (ce *CallExpr) Evaluate(lox *Interpreter) Object {
	// Couldn't figure out a cleaner way to bolt on native functions.
	if ie, ok := ce.callee.(*VariableExpr); ok && ie.name.Lexeme == "clock" {
		return &LoxNumber{float64(time.Now().Unix())}
	}

	callee := ce.callee.Evaluate(lox)

	var callable Callable
	switch callee.(type) {
	case *LoxFunction:
		callable = callee.(*LoxFunction)
	case *LoxClass:
		callable = callee.(*LoxClass)
	default:
		runtimeError("Can only call functions and classes.")
	}

	if len(ce.args) != callable.Arity() {
		runtimeError(fmt.Sprintf(
			"Expected %d arguments but got %d.", callable.Arity(), len(ce.args),
		))
	}

	args := []Object{}
	for _, arg := range ce.args {
		args = append(args, arg.Evaluate(lox))
	}

	return callable.Call(lox, args)
}

func (ge *GetExpr) Evaluate(lox *Interpreter) Object {
	obj := ge.object.Evaluate(lox)

	inst, ok := IsInstance(obj)
	if !ok {
		runtimeError("Only instances have properties.")
	}

	return inst.Get(ge.name.Lexeme)
}

func (te *ThisExpr) Evaluate(lox *Interpreter) Object {
	return lox.LookUpVariable(te, te.keyword.Lexeme)
}

func (be *BinaryExpr) Evaluate(lox *Interpreter) Object {
	left := be.left.Evaluate(lox)
	right := be.right.Evaluate(lox)

	switch be.op.Type {
	case PLUS:
		a, aok := IsString(left)
		b, bok := IsString(right)
		if aok && bok {
			return &LoxString{a + b}
		}

		c, cok := IsNumber(left)
		d, dok := IsNumber(right)
		if cok && dok {
			return &LoxNumber{c + d}
		}

		runtimeError("Operands must be two numbers or two strings.")

	case MINUS:
		a, b := assertNumbers(left, right)
		return &LoxNumber{a - b}

	case STAR:
		a, b := assertNumbers(left, right)
		return &LoxNumber{a * b}

	case SLASH:
		a, b := assertNumbers(left, right)
		return &LoxNumber{a / b}

	case GREATER:
		a, b := assertNumbers(left, right)
		return &LoxBool{a > b}

	case GREATER_EQUAL:
		a, b := assertNumbers(left, right)
		return &LoxBool{a >= b}

	case LESS:
		a, b := assertNumbers(left, right)
		return &LoxBool{a < b}

	case LESS_EQUAL:
		a, b := assertNumbers(left, right)
		return &LoxBool{a <= b}

	case EQUAL_EQUAL:
		return &LoxBool{isEqual(left, right)}

	case BANG_EQUAL:
		return &LoxBool{!isEqual(left, right)}
	}

	panic("unreachable: BinaryExpression.Evaluate(lox)")
}

func (ge *GroupExpr) Evaluate(lox *Interpreter) Object {
	return ge.group.Evaluate(lox)
}

func (le *LiteralExpr) Evaluate(lox *Interpreter) Object {
	switch le.token.Type {
	case TRUE:
		return &LoxBool{true}
	case FALSE:
		return &LoxBool{false}
	case NIL:
		return &LoxNil{}
	case STRING:
		return &LoxString{le.token.Literal}
	case NUMBER:
		n, _ := strconv.ParseFloat(le.token.Literal, 64)
		return &LoxNumber{n}
	}
	panic("unreachable: LiteralExpression.Evaluate(lox)")
}

func (ve *VariableExpr) Evaluate(lox *Interpreter) Object {
	return lox.LookUpVariable(ve, ve.name.Lexeme)
}

func (se *SuperExpr) Evaluate(lox *Interpreter) Object {
	distance := lox.locals[se]
	superclass := lox.GetAt(distance, "super").(*LoxClass)
	instance := lox.GetAt(distance-1, "this").(*LoxInstance) //look an environment nearer for this

	method := superclass.FindMethod(se.method.Lexeme)
	if method == nil {
		runtimeError("Undefined property: " + se.method.Lexeme)
	}
	return method.bind(instance)
}

// --------------- Helper Functions --------------- //
func assertNumbers(left, right Object) (float64, float64) {
	a, aok := IsNumber(left)
	b, bok := IsNumber(right)

	if !aok || !bok {
		runtimeError("Operands must be numbers.")
	}

	return a, b
}

func isEqual(left, right Object) bool {
	leftNil := IsNil(left)
	rightNil := IsNil(right)

	if leftNil && rightNil {
		return true
	}
	if leftNil || rightNil {
		return false
	}

	n1, leftNumber := IsNumber(left)
	n2, rightNumber := IsNumber(right)
	if leftNumber && rightNumber {
		return n1 == n2
	}

	s1, leftString := IsString(left)
	s2, rightString := IsString(right)
	if leftString && rightString {
		return s1 == s2
	}

	b1, leftBool := IsBool(left)
	b2, rightBool := IsBool(right)
	if leftBool && rightBool {
		return b1 == b2
	}

	return false
}

func assertNumber(obj Object) float64 {
	n, ok := IsNumber(obj)
	if !ok {
		runtimeError("Operand must be a number.")
	}
	return n
}

func runtimeError(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(70)
}
