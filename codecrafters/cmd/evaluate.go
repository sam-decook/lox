package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func (ae *AssignmentExpr) Evaluate(env *Environment) Object {
	val := ae.expr.Evaluate(env)
	env.Assign(ae.name, val)
	return val
}

// The logical operators return a value of the proper truthiness
func (loe *LogicOrExpr) Evaluate(env *Environment) Object {
	left := loe.left.Evaluate(env)
	if IsTruthy(left) {
		// Short-circuit
		return left
	}
	return loe.right.Evaluate(env)
}

// The logical operators return a value of the proper truthiness
func (lae *LogicAndExpr) Evaluate(env *Environment) Object {
	left := lae.left.Evaluate(env)
	if !IsTruthy(left) {
		// Short-circuit
		return left
	}
	return lae.right.Evaluate(env)
}

func (ue *UnaryExpr) Evaluate(env *Environment) Object {
	right := ue.right.Evaluate(env)

	switch ue.op.Type {
	case BANG:
		return NewBool(!IsTruthy(right))
	case MINUS:
		n := assertNumber(right)
		return NewNumber(-n)
	}
	panic("unreachable: UnaryExpression.Evaluate(env)")
}

func (ce *CallExpr) Evaluate(env *Environment) Object {
	// Couldn't figure out a cleaner way to bolt on native functions.
	if ie, ok := ce.callee.(*VariableExpr); ok && ie.name.Lexeme == "clock" {
		return NewNumber(float64(time.Now().Unix()))
	}

	callee := ce.callee.Evaluate(env)
	fn, ok := IsFunction(callee)
	if !ok {
		runtimeError("Can only call functions and classes.")
	}

	if len(ce.args) != len(fn.funDecl.params) {
		runtimeError(fmt.Sprintf(
			"Expected %d arguments but got %d.", len(fn.funDecl.params), len(ce.args),
		))
	}

	args := []Object{}
	for _, arg := range ce.args {
		args = append(args, arg.Evaluate(env))
	}

	return fn.Call(args)
}

func (be *BinaryExpr) Evaluate(env *Environment) Object {
	left := be.left.Evaluate(env)
	right := be.right.Evaluate(env)

	switch be.op.Type {
	case PLUS:
		a, aok := IsString(left)
		b, bok := IsString(right)
		if aok && bok {
			return NewString(a + b)
		}

		c, cok := IsNumber(left)
		d, dok := IsNumber(right)
		if cok && dok {
			return NewNumber(c + d)
		}

		runtimeError("Operands must be two numbers or two strings.")

	case MINUS:
		a, b := assertNumbers(left, right)
		return NewNumber(a - b)

	case STAR:
		a, b := assertNumbers(left, right)
		return NewNumber(a * b)

	case SLASH:
		a, b := assertNumbers(left, right)
		return NewNumber(a / b)

	case GREATER:
		a, b := assertNumbers(left, right)
		return NewBool(a > b)

	case GREATER_EQUAL:
		a, b := assertNumbers(left, right)
		return NewBool(a >= b)

	case LESS:
		a, b := assertNumbers(left, right)
		return NewBool(a < b)

	case LESS_EQUAL:
		a, b := assertNumbers(left, right)
		return NewBool(a <= b)

	case EQUAL_EQUAL:
		return NewBool(isEqual(left, right))

	case BANG_EQUAL:
		return NewBool(!isEqual(left, right))
	}

	panic("unreachable: BinaryExpression.Evaluate(env)")
}

func (ge *GroupExpr) Evaluate(env *Environment) Object {
	return ge.group.Evaluate(env)
}

func (le *LiteralExpr) Evaluate(env *Environment) Object {
	switch le.token.Type {
	case TRUE:
		return NewBool(true)
	case FALSE:
		return NewBool(false)
	case NIL:
		return NewNil()
	case STRING:
		return NewString(le.token.Literal)
	case NUMBER:
		n, _ := strconv.ParseFloat(le.token.Literal, 64)
		return NewNumber(n)
	}
	panic("unreachable: LiteralExpression.Evaluate(env)")
}

func (ve *VariableExpr) Evaluate(env *Environment) Object {
	return env.Get(ve.name.Lexeme)
}

// --------------- Helper Functions --------------- //
func assertNumbers(left, right Object) (float64, float64) {
	a, aok := IsNumber(left)
	b, bok := IsNumber(right)

	if !aok || !bok {
		runtimeError("Operands must be a numbers.")
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
