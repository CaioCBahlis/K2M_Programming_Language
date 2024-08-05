package parser

import (
	"MyInterpreter/ast"
	"MyInterpreter/lexer"
	"MyInterpreter/token"
	"fmt"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	EXPONENT = 6
	PREFIX   = 7
	CALL     = 8
	INDEX
)

var VariablePropagationCache = map[string]int64{}

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.EXPONENT: EXPONENT,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

//Now, we are done with the lexical analysis, all words, symbols and numbers (hopefully) have been turned into a token
// Now, it's time to parse the tokens into relevant structures (semantic analysis)
// The Parser is going to read the token, and try to fit it (parse it) into a function inside the language
// When reading an "if" for example, the parser is going to parse it into an If Statement, which is structured with
// <If Token> <Expression Node> <'{' Token> <BlockStatement Node>  <'}' Token> (it supports 'else' but this is just an example
//of what the parser role is

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token //Token is a struct with Type and Literal
	peekToken token.Token //See Next token
	errors    []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)
	p.registerPrefix(token.WHILE, p.parseWhileLoop)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.EXPONENT, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	p.ShiftToken() //First Next so peekToken is a token[0]
	p.ShiftToken() //Second Next so curToken is a token[0]

	//ShiftToken makes Parser point to the next token
	// NextToken makes lexer read next char, turn it into a token,
	//return it, and point to the next char in l.input
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) PeekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) ShiftToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
	//similar to readChar or NextToken, but for tokens instead of chars
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{} //initiate AST
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF { //While token is not End of Life, parse it
		stmt := p.parseStatement() //Verify which type of statement it is LET, RETURN, FUNC
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		p.ShiftToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	//So far, there's only 3 types of structures, let statements, return statements and expressions
	//other structures like arrays, functions are inside the scope of this other divisions (Ex: Functions are parsed inside Let Statements)
	// Expressions are simply parsed token by token until we find the semicolon

	if p.curToken.Type == token.IDENT {
		if p.PeekTokenIs(token.PE, token.LE, token.ME, token.DE) {
			return p.parseCompoundAssignStatement()
		}
	}

	switch p.curToken.Type {
	case token.LET:
		return p.ParseLetStatement()
	case token.RETURN:
		return p.ParseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}

}

func (p *Parser) ParseLetStatement() *ast.LetStatement {
	//We are parsing a Let statement
	//Let statements have this structure
	// Let IDENTIFIER = or != EXPRESSION
	//p.curToken is already "let" so next token gotta be the IDENTIFIER

	stmt := &ast.LetStatement{Token: p.curToken} //Let Keyword

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal} //Variable Name

	if !p.expectPeek(token.ASSIGN) {
		return nil //Equal sign
	}

	p.ShiftToken()

	stmt.Value = p.parseExpression(LOWEST) //Result of the Parsed Expression Ex: (5 + 5 * 10) -> 55
	if val, err := stmt.Value.(*ast.IntegerLiteral); err {
		VariablePropagationCache[stmt.Name.Value] = val.Value
	}
	for !p.curTokenIs(token.SEMICOLON) {
		p.ShiftToken() //shift forward until semicolon
	}

	return stmt
}

func (p *Parser) ParseReturnStatement() *ast.ReturnStatement {
	//Here we are interested in only 2 tokens
	//The first token has to be the return token
	//the second token is an expression that still has to be parsed

	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.ShiftToken() //Next Token Has to be the Expression

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.PeekTokenIs(token.SEMICOLON) {
		p.ShiftToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	//defer untrace(trace("parseExpressionStatement"))
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.PeekTokenIs(token.SEMICOLON) {
		p.ShiftToken()
	}

	return stmt
}

//ParseExpression Implements Pratt's Top Down Parsing Algorithm, which is absolutely gorgeous
// An expression can look something like 5 + 5, e + phi or even pi + tau * gamma | InfixExpression<<IntegerLiteral> <Operator> <Identifier>>|
// Now we need to evaluate the expression correctly so the AST executes is correctly ex: 5 + 2 * 10 --> (5 + (2 * 10))
// For this, we are going to separate our expression token in 2 parts: prefixes (!, Ident, IntegerLiteral, -, ...) and infixes (+, -, *, /, ...)
// first we parse the prefix by identifying it in the prefix map (prefixParseFns), then in case there are infixes, we create an object InfixExpression
// Attaching the prefix to InfixExpression.left. Now, we check the priority of our curtoken operator and call parseExpression recursively for the rest of the expression
// The same checks are going to be run again, in case there are more Infixes Operators with higher priority, another Infix Object is going to be created
// Ex: InfixExpression< Left<<Ident>> <Operator> <Right> > ---> InfixExpression< Left< <Ident> > <Operator> <Right <InfixExpression< <Ident> <Operator> <Right> > > >

func (p *Parser) parseExpression(precedence int) ast.Expression {
	//defer untrace(trace("parseExpression"))
	prefix := p.prefixParseFns[p.curToken.Type]
	//The Object of the Function that parses the cur prefix
	// A prefix of an IDENT, for example, is going to return parseIdentifier

	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	leftExp := prefix()

	for !p.PeekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {

			return leftExp
		}

		p.ShiftToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	// There's two places in this language where we can find lists of expressions: function parameters and lists
	//instead of creating a parsing method for parameters and lists, the author decided to create this method
	//which evaluates expressions and stops based on the last char of the list: ')' (parameters), ']' [list]
	list := []ast.Expression{}

	if p.PeekTokenIs(end) {
		p.ShiftToken()
		return list
	}

	p.ShiftToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.PeekTokenIs(token.COMMA) {
		p.ShiftToken()
		p.ShiftToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	//defer untrace(trace("parsePrefixExpression"))
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.ShiftToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	//defer untrace(trace("parseInfixExpression"))
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.ShiftToken()
	expression.Right = p.parseExpression(precedence)

	switch left.(type) {
	case *ast.Identifier:
		leftvar, _ := left.(*ast.Identifier)
		leftnum := VariablePropagationCache[leftvar.Value]

		if rightnum, ok := expression.Right.(*ast.IntegerLiteral); ok && leftnum != 0 {
			return p.ConstantFolding(leftnum, expression.Operator, rightnum.Value)

		} else {
			rightvar, ok := left.(*ast.Identifier)
			rightnum := VariablePropagationCache[rightvar.Value]
			if ok {
				return p.ConstantFolding(leftnum, expression.Operator, rightnum)
			}
		}

	case *ast.IntegerLiteral:
		if rightInteger, ok := expression.Right.(*ast.IntegerLiteral); ok {
			return p.ConstantFolding(left.(*ast.IntegerLiteral).Value, expression.Operator, rightInteger.Value)
		} else {
			rightvar, ok := expression.Right.(*ast.Identifier)
			rightnum := VariablePropagationCache[rightvar.Value]
			if ok {

				return p.ConstantFolding(left.(*ast.IntegerLiteral).Value, expression.Operator, rightnum)

			}
		}
	}

	return expression
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.ShiftToken()

	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.PeekTokenIs(token.ELSE) {
		p.ShiftToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	Block := &ast.BlockStatement{Token: p.curToken}
	Block.Statements = []ast.Statement{}

	p.ShiftToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			Block.Statements = append(Block.Statements, stmt)
		}
		p.ShiftToken()
	}
	return Block
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	//defer untrace(trace("IntegerLiteral"))
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifier := []*ast.Identifier{}

	if p.PeekTokenIs(token.RPAREN) {
		p.ShiftToken()
		return identifier
	}

	p.ShiftToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifier = append(identifier, ident)

	for p.PeekTokenIs(token.COMMA) {
		p.ShiftToken() //Shifts to comma
		p.ShiftToken() //Shifts to token after comma
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifier = append(identifier, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifier
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.PeekTokenIs(token.RPAREN) {
		p.ShiftToken()
		return args
	}

	p.ShiftToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.PeekTokenIs(token.COMMA) {
		p.ShiftToken()
		p.ShiftToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.ShiftToken()

	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) PeekTokenIs(t ...token.TokenType) bool {
	for _, tokens := range t {
		if p.peekToken.Type == tokens {
			return true
		}
	}
	return false
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.PeekTokenIs(t) {
		p.ShiftToken()
		return true
	} else {
		p.PeekError(t)
		return false
	}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.ShiftToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}

	array.Elements = p.parseExpressionList(token.RBRACKET)

	return array
}

func (p *Parser) parseCompoundAssignStatement() ast.Statement {

	pleql := &ast.CompoundAssignment{Token: p.curToken}
	if p.curToken.Type == token.IDENT {
		pleql.Variable = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	} else {
		msg := fmt.Sprintf("expected identifier, got %s", p.curToken.Type)
		p.errors = append(p.errors, msg)
	}

	p.ShiftToken()

	pleql.Operator = p.curToken.Literal

	p.ShiftToken()

	pleql.Value = p.parseExpression(LOWEST)

	for !p.curTokenIs(token.SEMICOLON) {
		p.ShiftToken()
	}

	return pleql
}
func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken} // "{"
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.PeekTokenIs(token.RBRACE) {
		p.ShiftToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.ShiftToken()
		value := p.parseExpression(LOWEST)

		hash.Pairs[key] = value

		if !p.PeekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hash
}

func (p *Parser) parseWhileLoop() ast.Expression {
	WLoop := &ast.WhileLoop{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.ShiftToken()

	WLoop.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	WLoop.Consequence = p.parseBlockStatement()

	if !p.curTokenIs(token.RBRACE) {
		return nil
	}

	return WLoop
}

func (p *Parser) ConstantFolding(leftnum int64, Operator string, rightnum int64) ast.Expression {
	//Constant Folding is a function added to optimize the interpreter
	//even for simple operations, the input goes through lexing, parsing and evaluation
	//By implementing ConstantFolding, we can skip the entire parsing and evaluation phase and return the integer object straight up
	//In average, Constant Folding saves about 10ms in the entire process

	//Additionally, I'm experimenting with strength reduction alternatives for common arithmetic operations

	switch Operator {
	case "+":
		result := leftnum + rightnum
		return &ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: strconv.FormatInt(result, 10)}, Value: result}
	case "-":
		result := leftnum - rightnum
		return &ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: strconv.FormatInt(result, 10)}, Value: result}

	case "*":
		result := leftnum * rightnum
		return &ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: strconv.FormatInt(result, 10)}, Value: result}

	case "/":
		if rightnum == 0 {
			return nil
		}
		result := leftnum / rightnum
		return &ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: strconv.FormatInt(result, 10)}, Value: result}

	case ">":
		result := leftnum > rightnum
		if result {
			return &ast.Boolean{Token: token.Token{Type: token.TRUE, Literal: "True"}, Value: result}
		} else {
			return &ast.Boolean{Token: token.Token{Type: token.FALSE, Literal: "False"}, Value: result}
		}
	case "<":
		result := leftnum < rightnum
		if result {
			return &ast.Boolean{Token: token.Token{Type: token.TRUE, Literal: "True"}, Value: result}
		} else {
			return &ast.Boolean{Token: token.Token{Type: token.FALSE, Literal: "False"}, Value: result}
		}
	case "==":
		result := leftnum == rightnum
		if result {
			return &ast.Boolean{Token: token.Token{Type: token.TRUE, Literal: "True"}, Value: result}
		} else {
			return &ast.Boolean{Token: token.Token{Type: token.FALSE, Literal: "False"}, Value: result}
		}
	case "!=":
		result := leftnum != rightnum
		if result {
			return &ast.Boolean{Token: token.Token{Type: token.TRUE, Literal: "True"}, Value: result}
		} else {
			return &ast.Boolean{Token: token.Token{Type: token.FALSE, Literal: "False"}, Value: result}
		}
	}
	return nil
}
