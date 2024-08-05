package evaluator

import (
	"MyInterpreter/lexer"
	"MyInterpreter/object"
	"MyInterpreter/parser"
	"encoding/csv"
	"os"
	"testing"
	"time"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
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
	env := object.NewEnvironment()

	return Eval(program, env)
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
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"True == True", true},
		{"False == False", true},
		{"True == False", false},
		{"True != False", true},
		{"False != True", true},
		{"(1 < 2) == True", true},
		{"(1 < 2) == False", false},
		{"(1 > 2) == True", false},
		{"(1 > 2) == False", true},
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

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (True) {10}", 10},
		{"if (False) {10}", nil},
		{"if (1) {10}", 10},
		{"if (1 < 2) {10}", 10},
		{"if (1 > 2) {10}", nil},
		{"if (1 > 2) {10} else {20}", 20},
		{"if (1 < 2) {10} else {20}", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not NULL, got=%T (%+v)", obj, obj)
		return false
	}
	return true
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9", 10},
		{"9; return 2 * 5; 9;", 10},
		{` 
		if (10 > 1) {
			if (10 > 1) {
       			return 10;
				}
			return 1;}
		`, 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{"5 + True;",
			"type mismatch: INTEGER + BOOLEAN"},
		{"5 + True; 5", "type mismatch: INTEGER + BOOLEAN"},
		{"-True", "unknown operator: -BOOLEAN"},
		{"True + False", "unknown operator: BOOLEAN + BOOLEAN"},
		{"5; True + False", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if (10 > 1) {True + False}", "unknown operator: BOOLEAN + BOOLEAN"},
		{`
			if (10 > 1){
				if (10 > 1){
				return True + False}
			return 1};
				`, "unknown operator: BOOLEAN + BOOLEAN"},
		{"foobar", "identifier not found: foobar"},
		{
			`{"name": "Monkey"}[fn(x) {x}];`,
			"unusable as hash key: FUNCTION",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got %T(%v)", evaluated, evaluated)
			continue
		}
		if errObj.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expectedMessage, errObj.Message)
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := "fn(x) {x + 2}"

	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Patameters=%+v", fn.Parameters)
	}
	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameters it not 'x' got %q", fn.Parameters[0])
	}

	expectedBody := "(x + 2)"

	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q got %q instead", expectedBody, fn.Body.String())
	}

}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) {x;}; identity(5);", 5},
		{"let identity = fn(x) {return x;}; identity(5);", 5},
		{"let double = fn(x) {x * 2;}; double(5);", 10},
		{"let add = fn(x,y) {x + y;}; add(5, 5);", 10},
		{"let add = fn(x,y) {x + y;}; add(5 + 5, add(5, 5));", 20},
		{"fn(x) {x;}(5)", 5},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `
	let newAdder = fn(x){
		fn(y) {x + y};
	};
	
	let addTwo = newAdder(2);
	addTwo(2);
	`

	testIntegerObject(t, testEval(input), 4)
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+V)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("wrong string value. got=%q instead of 'Hello,World!'", str.Value)
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (=%v)", evaluated, evaluated)
				continue
			}

			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q", expected, errObj.Message)
			}
		}
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 +3]"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not Array, got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d", len(result.Elements))
	}
	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], 6)
}

func TestArrayIndexExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"[1,2,3][2]",
			3,
		},
		{
			"let i = 0; [1][i];",
			1,
		},
		{
			"[1,2,3][1 + 1];",
			3,
		},
		{
			"let myArray = [1,2,3]; myArray[2];",
			3,
		},
		{
			"let myArray = [1,2,3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]",
			2,
		},
		{
			"[1, 2, 3][3]",
			nil,
		},
		{
			"[1,2,3][-1]",
			3,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}

}

func TestHashLiterals(t *testing.T) {
	input := `let two = "two";
{"one": 10-9, two: 1 + 1, "thr" + "ee": 6/2, 4:4, True:5, False:6}`

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Hash)
	if !ok {
		t.Fatalf("Eval didn't return Hash. got=%T (%+v)", evaluated, evaluated)
	}

	expected := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
		TRUE.HashKey():                             5,
		FALSE.HashKey():                            6,
	}

	if len(result.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong num of pairs. got=%d", len(result.Pairs))
	}

	for expectedKey, expectedValue := range expected {
		pair, ok := result.Pairs[expectedKey]
		if !ok {
			t.Errorf("no pair for given key in pairs")
		}

		testIntegerObject(t, pair.Value, expectedValue)
	}
}

func TestHashIndexExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}["bar"]`,
			nil,
		},
		{
			`let key = "foo"; {"foo": 5}[key]`,
			5,
		},
		{
			`{}["foo"]`,
			nil,
		},
		{
			`{5: 5}[5]`,
			5,
		},
		{
			`{True: 5}[True]`,
			5,
		},
		{
			`{False: 5}[False]`,
			5,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func BenchmarkEvaluator(b *testing.B) {
	b.StartTimer()
	for i := 0; i < 1000; i++ {
		TestReturnStatements(&testing.T{})
		TestErrorHandling(&testing.T{})
		TestLetStatements(&testing.T{})
		TestArrayIndexExpression(&testing.T{})
		TestArrayLiterals(&testing.T{})
		TestBangOperator(&testing.T{})
		TestBuiltinFunctions(&testing.T{})
		TestClosures(&testing.T{})
		TestEvalBooleanExpression(&testing.T{})
		TestEvalIntegerExpression(&testing.T{})
		TestFunctionApplication(&testing.T{})
		TestFunctionObject(&testing.T{})
		TestHashIndexExpression(&testing.T{})
		TestHashLiterals(&testing.T{})
		TestIfElseExpressions(&testing.T{})
		TestStringConcatenation(&testing.T{})
	}
	b.StopTimer()

	file, err := os.OpenFile("/Users/caiobahlis/GolandProjects/MyInterpreter/K2M_Programming_Language-main/BenchData.csv", os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModeAppend)
	if err != nil {
		b.Fatalf("Error appending BenchMark results to .csv file. WOMP WOMP")
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	date := time.Now().String()
	Duration := b.Elapsed().String()
	//Machine_Data, _ := exec.Command("lscpu").Output()
	Machine_Data := "Apple M2/ARM"
	Process := "Evaluator Benchmark - 1000 iterations - " + string(Machine_Data)

	writer.Write([]string{date + " | " + Process + " | " + "Duration: " + Duration})
}
