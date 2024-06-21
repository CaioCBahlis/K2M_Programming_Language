package parser

import (
	"MyInterpreter/ast"
	"MyInterpreter/lexer"
	"MyInterpreter/token"
	"fmt"
	"strconv"
)

const(
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
)

var precedences = map[token.TokenType]int{
	token.EQ: EQUALS,
	token.NOT_EQ: EQUALS,
	token.LT: LESSGREATER,
	token.GT: LESSGREATER,
	token.PLUS: SUM,
	token.MINUS: SUM,
	token.SLASH: PRODUCT,
	token.ASTERISK: PRODUCT,
	}


type (
	prefixParseFn func() ast.Expression
	infixParseFn func(ast.Expression) ast.Expression
)


type Parser struct{
	l *lexer.Lexer

	curToken token.Token //Token is a struct with Type and Literal 
	peekToken token.Token //See Next token
	errors []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns map[token.TokenType]infixParseFn

	//parser job is to get our tokens that were tokenized by our lexer
	// and Divide them into our AST so we can interpret them
	// We are going to divide the tokens into lines of code with context
	// LetStatement, returnStatement, ifStatement
	// Based on that we group our tokens,
	// LetStatement for ex, has the let keyword, an identifier (var), and an expression
	//after grouping the tokens in this context, we arrange them into our ASTTree
}

func NewParser(l *lexer.Lexer) *Parser{
	p := &Parser{l: l, errors: []string{}} 

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

	p.ShiftToken() //First Next so peekToken is a token[0]
	p.ShiftToken() //Second Next so curToken is a token[0]

	//ShiftToken makes Parser point to the next token
	// NextToken makes lexer read next char, turn it into a token,
	//return it, and point to the next char in l.input
	return p
}

func (p *Parser) parsePrefixExpression() ast.Expression{
	expression := &ast.PrefixExpression{
	Token: p.curToken,
	Operator: p.curToken.Literal,
	}

	p.ShiftToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression{
	expression := &ast.InfixExpression{
		Token: p.curToken,
		Operator: p.curToken.Literal,
		Left: left,
	}

	precedence := p.curPrecedence()
	p.ShiftToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}


func (p *Parser) parseIdentifier() ast.Expression{
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) Errors() []string{
	return p.errors
}

func (p *Parser) PeekError(t token.TokenType){
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) ShiftToken(){
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program{
	program := &ast.Program{} //initiate root
	program.Statements = []ast.Statement{} //initiate statement list


	for p.curToken.Type != token.EOF{ //While token is not End of Life, parse it
		stmt := p.parseStatement()	//Verify which type of statement it is LET, RETURN, FUNC
		if stmt != nil{
			program.Statements = append(program.Statements, stmt)
		}
		p.ShiftToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement{
	switch p.curToken.Type{
		case token.LET:
			return p.ParseLetStatement()
		case token.RETURN:
			return p.ParseReturnStatement()
		default:
			return p.parseExpressionStatement()
	}
}


func (p *Parser) ParseLetStatement() *ast.LetStatement{
	//We are parsing a Let statement
	//Let statements have this structure
	// Let IDENTIFIER = or != EXPRESSION
	//p.curToken is already "let" so next token gotta be the IDENTIFIER


	stmt := &ast.LetStatement{Token: p.curToken} //Let Keyword

	if !p.expectPeek(token.IDENT){
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal } //Variable Name

	if !p.expectPeek(token.ASSIGN){
		return nil								//Equal sign
	}

	for !p.curTokenIs(token.SEMICOLON){
		p.ShiftToken()					//forward until semicolon
	}

	return stmt

}

func (p *Parser) curTokenIs(t token.TokenType) bool{
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool{
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool{
	if p.peekTokenIs(t){
		p.ShiftToken()
		return true
	}else{
		p.PeekError(t)
		return false
	}
}


func (p *Parser) ParseReturnStatement() *ast.ReturnStatement{
	//Here we are intersted in only 2 tokens
	//The first token has to be the return token
	//the second token is an expression that still has to be parsed

	stmt := &ast.ReturnStatement{Token: p.curToken}

	//TODO We have to parse the expression and add it to the returnstmt object

	p.ShiftToken() //Next Token Has to be the Expression

	for !p.curTokenIs(token.SEMICOLON){
		p.ShiftToken()
	}

	return stmt
}



func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn){
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn){
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement{
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON){
		p.ShiftToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression{
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil{
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence(){
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil{
			return leftExp
		}
	}

	p.ShiftToken()

	return leftExp
}


func (p *Parser) parseIntegerLiteral() ast.Expression{
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil{
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int{
	if p, ok := precedences[p.peekToken.Type]; ok{
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int{
	if p, ok := precedences[p.curToken.Type]; ok{
		return p
	}
	return LOWEST
}