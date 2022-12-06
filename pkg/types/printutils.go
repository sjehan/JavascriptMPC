package types

import (
	"fmt"

	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/token"
)

/*
 * This file provides functions used to print the AST and its
 * various components.
 */

// Parameter to decide wether or not we print the declarations of the program
// and of the functions.
var printDeclarations bool = false

// Token2string is used to print operators
var Token2string = [...]string{
	token.ILLEGAL:                     "ILLEGAL",
	token.EOF:                         "EOF",
	token.COMMENT:                     "COMMENT",
	token.KEYWORD:                     "KEYWORD",
	token.STRING:                      "STRING",
	token.BOOLEAN:                     "BOOLEAN",
	token.NULL:                        "NULL",
	token.NUMBER:                      "NUMBER",
	token.IDENTIFIER:                  "IDENTIFIER",
	token.PLUS:                        "+",
	token.MINUS:                       "-",
	token.MULTIPLY:                    "*",
	token.SLASH:                       "/",
	token.REMAINDER:                   "%",
	token.AND:                         "&",
	token.OR:                          "|",
	token.EXCLUSIVE_OR:                "^",
	token.SHIFT_LEFT:                  "<<",
	token.SHIFT_RIGHT:                 ">>",
	token.UNSIGNED_SHIFT_RIGHT:        ">>>",
	token.AND_NOT:                     "&^",
	token.ADD_ASSIGN:                  "+=",
	token.SUBTRACT_ASSIGN:             "-=",
	token.MULTIPLY_ASSIGN:             "*=",
	token.QUOTIENT_ASSIGN:             "/=",
	token.REMAINDER_ASSIGN:            "%=",
	token.AND_ASSIGN:                  "&=",
	token.OR_ASSIGN:                   "|=",
	token.EXCLUSIVE_OR_ASSIGN:         "^=",
	token.SHIFT_LEFT_ASSIGN:           "<<=",
	token.SHIFT_RIGHT_ASSIGN:          ">>=",
	token.UNSIGNED_SHIFT_RIGHT_ASSIGN: ">>>=",
	token.AND_NOT_ASSIGN:              "&^=",
	token.LOGICAL_AND:                 "&&",
	token.LOGICAL_OR:                  "||",
	token.INCREMENT:                   "++",
	token.DECREMENT:                   "--",
	token.EQUAL:                       "==",
	token.STRICT_EQUAL:                "===",
	token.LESS:                        "<",
	token.GREATER:                     ">",
	token.ASSIGN:                      "=",
	token.NOT:                         "!",
	token.BITWISE_NOT:                 "~",
	token.NOT_EQUAL:                   "!=",
	token.STRICT_NOT_EQUAL:            "!==",
	token.LESS_OR_EQUAL:               "<=",
	token.GREATER_OR_EQUAL:            ">=",
	token.LEFT_PARENTHESIS:            "(",
	token.LEFT_BRACKET:                "[",
	token.LEFT_BRACE:                  "{",
	token.COMMA:                       ",",
	token.PERIOD:                      ".",
	token.RIGHT_PARENTHESIS:           ")",
	token.RIGHT_BRACKET:               "]",
	token.RIGHT_BRACE:                 "}",
	token.SEMICOLON:                   ";",
	token.COLON:                       ":",
	token.QUESTION_MARK:               "?",
	token.IF:                          "if",
	token.IN:                          "in",
	token.DO:                          "do",
	token.VAR:                         "var",
	token.FOR:                         "for",
	token.NEW:                         "new",
	token.TRY:                         "try",
	token.THIS:                        "this",
	token.ELSE:                        "else",
	token.CASE:                        "case",
	token.VOID:                        "void",
	token.WITH:                        "with",
	token.WHILE:                       "while",
	token.BREAK:                       "break",
	token.CATCH:                       "catch",
	token.THROW:                       "throw",
	token.RETURN:                      "return",
	token.TYPEOF:                      "typeof",
	token.DELETE:                      "delete",
	token.SWITCH:                      "switch",
	token.DEFAULT:                     "default",
	token.FINALLY:                     "finally",
	token.FUNCTION:                    "function",
	token.CONTINUE:                    "continue",
	token.DEBUGGER:                    "debugger",
	token.INSTANCEOF:                  "instanceof",
}

// printExpression is used to print every kind of node implementing
// the ast.Expression interface.
func printExpression(E ast.Expression, indent string) {
	switch exp := E.(type) {
	case *ast.ArrayLiteral:
		if exp != nil {
			fmt.Println(indent, "ArrayLiteral")
			for i, val := range exp.Value {
				fmt.Println(indent, "[", i, "]")
				printExpression(val, indent+"\t")
			}
		}
	case *ast.AssignExpression:
		if exp != nil {
			fmt.Println(indent, "AssignExpression")
			fmt.Println(indent, "Left:")
			printExpression(exp.Left, indent+"\t")
			fmt.Println(indent, "Right:")
			printExpression(exp.Right, indent+"\t")
		}
	case *ast.BadExpression:
		if exp != nil {
			fmt.Println(indent, "BadExpression")
		}
	case *ast.BinaryExpression:
		if exp != nil {
			fmt.Println(indent, "BinaryExpression")
			fmt.Println(indent, "Left:")
			printExpression(exp.Left, indent+"\t")
			fmt.Println(indent, "Operator:", Token2string[exp.Operator])
			fmt.Println(indent, "Right:")
			printExpression(exp.Right, indent+"\t")
			fmt.Println(indent, "Comparison:", exp.Comparison)
		}
	case *ast.BooleanLiteral:
		if exp != nil {
			fmt.Println(indent, "BooleanLiteral ", exp.Value)
		}
	case *ast.BracketExpression:
		if exp != nil {
			fmt.Println(indent, "BracketExpression")
			fmt.Println(indent, "Left:")
			printExpression(exp.Left, indent+"\t")
			fmt.Println(indent, "Member:")
			printExpression(exp.Member, indent+"\t")
		}
	case *ast.CallExpression:
		if exp != nil {
			fmt.Println(indent, "CallExpression")
			fmt.Println(indent, "Callee:")
			printExpression(exp.Callee, indent+"\t")
			fmt.Println(indent, "Arguments:")
			for _, arg := range exp.ArgumentList {
				printExpression(arg, indent+"\t")
				fmt.Println()
			}
		}
	case *ast.ConditionalExpression:
		if exp != nil {
			fmt.Println(indent, "ConditionalExpression")
			fmt.Println(indent, "Test:")
			printExpression(exp.Test, indent+"\t")
			fmt.Println(indent, "Consequent:")
			printExpression(exp.Consequent, indent+"\t")
			fmt.Println(indent, "Alternate:")
			printExpression(exp.Alternate, indent+"\t")
		}
	case *ast.DotExpression:
		if exp != nil {
			fmt.Println(indent, "DotExpression")
			fmt.Println(indent, "Left:")
			printExpression(exp.Left, indent+"\t")
			fmt.Println(indent, "Identifier:")
			fmt.Println(indent+"\t", "Name:", exp.Identifier.Name)
		}
	case *ast.EmptyExpression:
		if exp != nil {
			fmt.Println(indent, "EmptyExpression")
		}
	case *ast.FunctionLiteral:
		if exp != nil {
			fmt.Println(indent, "FunctionLiteral")
			fmt.Println(indent, "Name:", exp.Name.Name)
			fmt.Println(indent, "Parameter list:")
			for _, p := range exp.ParameterList.List {
				fmt.Println(indent+"\t", p.Name)
			}
			fmt.Println(indent, "Body:")
			printStatement(exp.Body, indent+"\t")
			if printDeclarations {
				fmt.Println(indent, "Declaration list:")
				printDeclarationList(exp.DeclarationList, indent+"\t")
			}
		}
	case *ast.Identifier:
		if exp != nil {
			fmt.Println(indent, "Identifier", exp.Name)
		}
	case *ast.NewExpression:
		if exp != nil {
			fmt.Println(indent, "NewExpression")
			fmt.Println(indent, "Callee:")
			printExpression(exp.Callee, indent+"\t")
			fmt.Println(indent, "Argument list:")
			for _, arg := range exp.ArgumentList {
				printExpression(arg, indent+"\t")
			}
		}
	case *ast.NullLiteral:
		if exp != nil {
			fmt.Println(indent, "NullLiteral")
		}
	case *ast.NumberLiteral:
		if exp != nil {
			fmt.Println(indent, "NumberLiteral ", exp.Value)
		}
	case *ast.ObjectLiteral:
		if exp != nil {
			fmt.Println(indent, "ObjectLiteral")
			fmt.Println(indent, "Properties:")
			for _, prop := range exp.Value {
				fmt.Println(indent+"\t"+"Key: ", prop.Key)
				fmt.Println(indent+"\t"+"Kind: ", prop.Kind)
				fmt.Println(indent + "\t" + "Value: ")
				printExpression(prop.Value, indent+"\t\t")
			}
		}
	case *ast.RegExpLiteral:
		if exp != nil {
			// TODO but should not be used in Freegates
			fmt.Println(indent, "RegExpLiteral")
		}
	case *ast.SequenceExpression:
		if exp != nil {
			fmt.Println(indent, "SequenceExpression")
			for _, xexp := range exp.Sequence {
				printExpression(xexp, indent+"\t")
			}
		}
	case *ast.StringLiteral:
		if exp != nil {
			fmt.Println(indent, "StringLiteral")
			fmt.Println(indent, "Value:", exp.Value)
		}
	case *ast.ThisExpression:
		if exp != nil {
			fmt.Println(indent, "Identifier")
		}
	case *ast.UnaryExpression:
		if exp != nil {
			fmt.Println(indent, "UnaryExpression")
			fmt.Println(indent, "Operator:", Token2string[exp.Operator])
			fmt.Println(indent, "Operand:")
			printExpression(exp.Operand, indent+"\t")
			fmt.Println(indent, "Postfix:", exp.Postfix)

		}
	case *ast.VariableExpression:
		if exp != nil {
			fmt.Println(indent, "VariableExpression of name:", exp.Name)
			if exp.Initializer != nil {
				fmt.Println(indent, "Initialized to:")
				printExpression(exp.Initializer, indent+"\t")
			}
		}
	}
}

// printStatement is used to print every kind of node implementing
// the ast.Statement interface.
func printStatement(S ast.Statement, indent string) {
	switch st := S.(type) {
	case *ast.BadStatement:
		if st != nil {
			fmt.Println(indent, "Bad statement :-(")
		}
	case *ast.BlockStatement:
		if st != nil {
			fmt.Println(indent, "Block statement")
			printStatementList(st.List, indent+"")
		}
	case *ast.BranchStatement:
		if st != nil {
			fmt.Println(indent, "Branch statement")
		}
	case *ast.CaseStatement:
		if st != nil {
			fmt.Println(indent, "Case statement")
		}
	case *ast.CatchStatement:
		if st != nil {
			fmt.Println(indent, "Catch statement")
		}
	case *ast.DoWhileStatement:
		if st != nil {
			fmt.Println(indent, "DoWhile statement")
		}
	case *ast.DebuggerStatement:
		if st != nil {
			fmt.Println(indent, "Debugger statement")
		}
	case *ast.EmptyStatement:
		if st != nil {
			fmt.Println(indent, "Empty statement")
		}
	case *ast.ExpressionStatement:
		if st != nil {
			fmt.Println(indent, "Expression statement")
			printExpression(st.Expression, indent+"\t")
		}
	case *ast.ForInStatement:
		if st != nil {
			fmt.Println(indent, "ForIn statement")
		}
	case *ast.ForStatement:
		if st != nil {
			fmt.Println(indent, "For statement")
			fmt.Println(indent, "Initializer:")
			printExpression(st.Initializer, indent+"\t")
			fmt.Println(indent, "Test:")
			printExpression(st.Test, indent+"\t")
			fmt.Println(indent, "Update:")
			printExpression(st.Update, indent+"\t")
			fmt.Println(indent, "Body:")
			printStatement(st.Body, indent+"\t")
			fmt.Println()
		}
	case *ast.FunctionStatement:
		if st != nil {
			fmt.Println(indent, "Function statement")
			printExpression(st.Function, indent)
		}
	case *ast.IfStatement:
		if st != nil {
			fmt.Println(indent, "If statement")
			fmt.Println(indent, "Test:")
			printExpression(st.Test, indent+"\t")
			fmt.Println(indent, "Consequent:")
			printStatement(st.Consequent, indent+"\t")
			fmt.Println(indent, "Alternate:")
			printStatement(st.Alternate, indent+"\t")
			fmt.Println()
		}
	case *ast.LabelledStatement:
		if st != nil {
			fmt.Println(indent, "Labelled statement")
		}
	case *ast.ReturnStatement:
		if st != nil {
			fmt.Println(indent, "Return statement")
			printExpression(st.Argument, indent+"\t")
		}
	case *ast.SwitchStatement:
		if st != nil {
			fmt.Println(indent, "Switch statement")
		}
	case *ast.ThrowStatement:
		if st != nil {
			fmt.Println(indent, "Throw statement")
		}
	case *ast.TryStatement:
		if st != nil {
			fmt.Println(indent, "Try statement")
		}
	case *ast.VariableStatement:
		if st != nil {
			fmt.Println(indent, "Variable statement")
			for _, v := range st.List {
				printExpression(v, indent+"\t")
			}
			fmt.Println()
		}
	case *ast.WhileStatement:
		if st != nil {
			fmt.Println(indent, "While statement")
			fmt.Println(indent, "Test:")
			printExpression(st.Test, indent+"\t")
			fmt.Println(indent, "Body:")
			printStatement(st.Body, indent+"\t")
		}
	case *ast.WithStatement:
		if st != nil {
			fmt.Println(indent, "With statement")
		}
	}
}

// printDeclarationList is used to print the declarations of the
// program or of a function.
func printDeclarationList(L []ast.Declaration, indent string) {
	for _, dec := range L {
		switch d := dec.(type) {
		case *ast.FunctionDeclaration:
			if d != nil {
				printExpression(d.Function, indent)
			}
		case *ast.VariableDeclaration:
			if d != nil {
				for j, v := range d.List {
					fmt.Println(indent, "Variable expression", j)
					printExpression(v, indent+"\t")
				}
			}
		}
	}
	fmt.Println()
}

// printStatementList is used to print a list of statement nodes
func printStatementList(L []ast.Statement, indent string) {
	for _, st := range L {
		printStatement(st, indent+"\t")
	}
	fmt.Println()
}

// PrintAST prints the whole ast from a program node
func PrintAST(prog *ast.Program, withDeclarations bool) {
	printDeclarations = withDeclarations

	if printDeclarations {
		fmt.Println("Number of declarations :", len(prog.DeclarationList))
		printDeclarationList(prog.DeclarationList, "")
	}
	fmt.Println("\n\nNumber of statements :", len(prog.Body))
	printStatementList(prog.Body, "")
	fmt.Println()
}
