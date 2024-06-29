package repl

import (
	"MyInterpreter/evaluator"
	"MyInterpreter/lexer"
	"MyInterpreter/parser"
	"bufio"
	"fmt"
	"io"
)

const PROMPT = ">>"

const MONKEY_FACE = `
						 __,__
					.--. .-" "-. .--.
					/ .. \/ .-. .-. \/ .. \
					| | '| / Y \ |' | |
					| \ \ \ 0 | 0 / / / |
					\ '- ,\.-"""""""-./, -' /
					''-' /_ ^ ^ _\ '-''
					| \._ _./ |
					\ \ '~' / /
					'._ '-=-' _.'
					'-----'
 `

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.NewLexer(line)
		p := parser.NewParser(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}
		evaluated := evaluator.Eval(program)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}

	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "WOMP WOMP! PARSER IS NOT HAPPY")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
