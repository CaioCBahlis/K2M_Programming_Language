package parser

import (
	"MyInterpreter/ast"
	"MyInterpreter/lexer"
	"testing"
	"fmt"
)


func TestLetStatements(t *testing.T){
	input :=  `
	let x = 5;
	let y = 10;
	let foobar = 838383;
	 `

	 l := lexer.NewLexer(input) //create new lexer
	 p := NewParser(l) //create new parser


	 program := p.ParseProgram()
	 checkParseErrors(t, p)

	 if program == nil{ //no root node, how are we going to parse a tree without roots????
		t.Fatalf("ParseProgram() returned nil")
	 }


	if len(program.Statements) != 3{ 
		t.Fatalf("program.Statements does not contain 3 statements, got %d", len(program.Statements))
		//Every Let Statement is composed of 3 pieces
		//The let keyword, which is represented as a lexer.Token.Literal
		// the Identifier, which is the name of the variable and represented by p.LetStatement.Name
		// and finally the expression contained in the let statement, represented by p.LetStatement.Expression
	}

	tests := []struct {
		expectedIdentifier string
		}{
			{"x"},
			{"y"},
			{"foobar"},
		}

	for i, tt := range tests{
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.expectedIdentifier){
			return
		}
	}
}

func checkParseErrors(t *testing.T, p *Parser){
	errors := p.Errors()

	if len(errors) == 0{
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors{
		t.Errorf("parse error %q", msg)
	}
	t.FailNow()

 }
	
func testLetStatement(t *testing.T, s ast.Statement, name string) bool{
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral != 'let'. got=%q", s.TokenLiteral())
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok{
		t.Errorf("s not *ast.LetStatement. got=%T", s)
	}

	if letStmt.Name.Value != name{
		t.Errorf("letStmt.Name.Value not %s. got %s instead", name, letStmt.Name.Value)
	}

	if letStmt.Name.TokenLiteral() != name{
		t.Errorf("s.Name not '%s'. got=%s instead", name, letStmt.Name)
	}

	return true

}

func TestReturnStatements(t *testing.T){
	input := `
	return = 5;
	return = 10;
	return = 99322;"
	`

	l := lexer.NewLexer(input)
	p := NewParser(l)

	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 3{
		t.Fatalf("program.Statements does not contain 3 statements, got %d", len(program.Statements))
	}

	for _, stmt := range program.Statements{
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok{
			t.Errorf("stmt not *ast.returnStatement, got=%T instead", stmt)
			continue
		}

		if returnStmt.TokenLiteral() != "return"{
			t.Errorf("returnStmt.TokenLiteral not return, got %q instead", returnStmt.TokenLiteral())
		}
	}

}

func TestIdentifierExpression(t *testing.T){
	input := "foobar;"

	l := lexer.NewLexer(input)
	p := NewParser(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)
	
	if len(program.Statements) != 1{
		t.Fatalf("program has not enough statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok{
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	ident, ok := stmt.Expression.(*ast.Identifier)
		if !ok{
			t.Fatalf("exp not *ast.Identifier, got=%T", stmt.Expression)
		}
	

	if ident.Value != "foobar"{
		t.Errorf("ident.Value not %s, got=%s instead", "foobar", ident.Value)
	}

	if ident.TokenLiteral() != "foobar"{
		t.Errorf("ident.TokenLiteral not %s, got=%s instead", "foobar", ident.TokenLiteral())
	}

}	

func TestIntegerLiteralExpression(t *testing.T){
	input := "5;"

	l := lexer.NewLexer(input)
	p := NewParser(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 1{
		t.Fatalf("program has not enough statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok{
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
		if !ok{
			t.Errorf("exp not *ast.IntegerLiteral. got%T", stmt.Expression)
		}

	if literal.TokenLiteral() != "5"{
		t.Errorf("literal.TokenLiteral not %s, got %s", "5", literal.TokenLiteral())
	}

}

func TestParsingPrefixExpressions(t *testing.T){
	prefixTests := []struct{
		input string
		operator string
		integerValue int64
	}{
		{"!5", "!", 5},
		{"-15", "-", 15},
	}

	for _, tt := range prefixTests{
		l := lexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		if len(program.Statements) != 1{
			t.Fatalf("program.Statements does not contain %d statements", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok{
		t.Fatalf("program.Statements does not contain %d statements. got %d instead", 1, len(program.Statements))
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok{
			t.Fatalf("stmt is not ast.PrefixExpression. got=%T", stmt.Expression)
		}

		if exp.Operator != tt.operator{
			t.Fatalf("exp.Operator is not %s, got=%s",exp.Operator, tt.operator)
		}
		
		if !testIntegerLiteral(t, exp.Right, tt.integerValue){
			return
		}
	}
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool{
		integ, ok := il.(*ast.IntegerLiteral)
		if !ok{
			t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
			return false
		}

		if integ.Value != value{
			t.Errorf("integ.Value got %d, got %d instead", value, integ.Value)
			return false

		}

		if integ.TokenLiteral() != fmt.Sprintf("%d", value){
			t.Errorf("integ.TokenLiteral not %d. got=%s", value, integ.TokenLiteral())
			return false
		}
	
		return true
}

		
func TestParsingInfixExpressions(t *testing.T){
	infixTests := []struct{
		input string
		leftValue int64
		operator string
		rightvalue int64
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		}

	for _, tt := range infixTests{
		l := lexer.NewLexer(tt.input)
		p := NewParser(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		if len(program.Statements) != 1{
			t.Fatalf("program Statements does not contain %d statements. got %d instead", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok{
			t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
		}
		exp, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok{
			t.Fatalf("program Statements[0] is not ast.ExpressionStatement. got%T", program.Statements[0])	
		}

		if !testIntegerLiteral(t, exp.Left, tt.leftValue){
			return
		}

		if exp.Operator != tt.operator{
			t.Fatalf("exp Operator is not %s got %s", exp.Operator, tt.operator)
		}
		
		if !testIntegerLiteral(t, exp.Right, tt.rightvalue){
			return
		}
	}	
}

 