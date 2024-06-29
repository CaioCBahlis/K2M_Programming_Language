package evaluator

import (
	"MyInterpreter/lexer"
	"MyInterpreter/object"
	"MyInterpreter/parser"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{input: "5", expected: 5},
		{input: "10", expected: 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func testEval(input string) object.Object {
	l := lexer.NewLexer(input)
	p := parser.NewParser(l)
	program := p.ParseProgram()

	return Eval(program)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T", obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("result.Value=%d, expected=%d", result.Value, expected)
		return false
	}

	return true
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input  string
		expect bool
	}{
		{"True", true},
		{"False", false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expect)
	}
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%=v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has the wrong value. got=%t, want=%t", result.Value, expected)
		return false
	}

	return true
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input  string
		expect bool
	}{
		{input: "!True", expect: false},
		{input: "!False", expect: true},
		{input: "!5", expect: false},
		{input: "!!True", expect: true},
		{input: "!!False", expect: false},
		{input: "!!5", expect: true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expect)
	}
}
