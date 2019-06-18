package strictus

import (
	. "bamboo-runtime/execution/strictus/ast"
	"github.com/antlr/antlr4/runtime/Go/antlr"
	. "github.com/onsi/gomega"
	"testing"
)

func TestParseComplexFunction(t *testing.T) {

	input := antlr.NewInputStream(`
		pub fun sum(a: i32, b: i32[2], c: i32[][3]): i64 {
            const x = 1
            var y: i32 = 2
            y = (3)
            x.foo.bar[0][1].baz
            z = sum(0o3, 0x2, 0b1) % 42
            return a
            while x < 2 {
                x = x + 1
            }
            if true {
                return 1
            } else if false {
                return 2 > 3 ? 4 : 5
            } else {
                return [2, true]
            }
        }
	`)

	lexer := NewStrictusLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	parser := NewStrictusParser(stream)
	// diagnostics, for debugging only:
	// parser.AddErrorListener(antlr.NewDiagnosticErrorListener(true))
	parser.AddErrorListener(antlr.NewConsoleErrorListener())
	actual := parser.Program().Accept(&ProgramVisitor{}).(Program)

	sum := FunctionDeclaration{
		IsPublic:   true,
		Identifier: "sum",
		Parameters: []Parameter{
			{Identifier: "a", Type: Int32Type{}},
			{Identifier: "b", Type: FixedType{Type: Int32Type{}, Size: 2}},
			{Identifier: "c", Type: DynamicType{Type: FixedType{Type: Int32Type{}, Size: 3}}},
		},
		ReturnType: Int64Type{},
		Block: Block{
			Statements: []Statement{
				VariableDeclaration{
					IsConst:    true,
					Identifier: "x",
					Type:       nil,
					Value:      UInt64Expression(1),
				},
				VariableDeclaration{
					IsConst:    false,
					Identifier: "y",
					Type:       Int32Type{},
					Value:      UInt64Expression(2),
				},
				Assignment{
					Target: IdentifierExpression{Identifier: "y"},
					Value:  UInt64Expression(3),
				},
				ExpressionStatement{
					Expression: MemberExpression{
						Expression: IndexExpression{
							Expression: IndexExpression{
								Expression: MemberExpression{
									Expression: MemberExpression{
										Expression: IdentifierExpression{Identifier: "x"},
										Identifier: "foo",
									},
									Identifier: "bar",
								},
								Index: UInt64Expression(0),
							},
							Index: UInt64Expression(1),
						},
						Identifier: "baz",
					},
				},
				Assignment{
					Target: IdentifierExpression{Identifier: "z"},
					Value: BinaryExpression{
						Operation: OperationMod,
						Left: InvocationExpression{
							Expression: IdentifierExpression{Identifier: "sum"},
							Arguments: []Expression{
								UInt64Expression(3),
								UInt64Expression(2),
								UInt64Expression(1),
							},
						},
						Right: UInt64Expression(42),
					},
				},
				ReturnStatement{Expression: IdentifierExpression{Identifier: "a"}},
				WhileStatement{
					Test: BinaryExpression{
						Operation: OperationLess,
						Left:      IdentifierExpression{Identifier: "x"},
						Right:     UInt64Expression(2),
					},
					Block: Block{
						Statements: []Statement{
							Assignment{
								Target: IdentifierExpression{Identifier: "x"},
								Value: BinaryExpression{
									Operation: OperationPlus,
									Left:      IdentifierExpression{Identifier: "x"},
									Right:     UInt64Expression(1),
								},
							},
						},
					},
				},
				IfStatement{
					Test: BoolExpression(true),
					Then: Block{
						Statements: []Statement{
							ReturnStatement{Expression: UInt64Expression(1)},
						},
					},
					Else: Block{
						Statements: []Statement{
							IfStatement{
								Test: BoolExpression(false),
								Then: Block{
									Statements: []Statement{
										ReturnStatement{
											Expression: ConditionalExpression{
												Test: BinaryExpression{
													Operation: OperationGreater,
													Left:      UInt64Expression(2),
													Right:     UInt64Expression(3),
												},
												Then: UInt64Expression(4),
												Else: UInt64Expression(5),
											},
										},
									},
								},
								Else: Block{
									Statements: []Statement{
										ReturnStatement{
											Expression: ArrayExpression{
												Values: []Expression{
													UInt64Expression(2),
													BoolExpression(true),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	expected := Program{
		AllDeclarations: []Declaration{sum},
		Declarations:    map[string]Declaration{"sum": sum},
	}

	NewWithT(t).Expect(actual).Should(Equal(expected))
}

func TestParseIntegerTypes(t *testing.T) {

	input := antlr.NewInputStream(`
		const a: i8 = 1
		const b: i16 = 2
		const c: i32 = 3
		const d: i64 = 4
		const e: u8 = 5
		const f: u16 = 6
		const g: u32 = 7
		const h: u64 = 8
	`)

	lexer := NewStrictusLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	parser := NewStrictusParser(stream)
	// diagnostics, for debugging only:
	// parser.AddErrorListener(antlr.NewDiagnosticErrorListener(true))
	parser.AddErrorListener(antlr.NewConsoleErrorListener())
	actual := parser.Program().Accept(&ProgramVisitor{}).(Program)

	a := VariableDeclaration{Identifier: "a", IsConst: true, Type: Int8Type{}, Value: UInt64Expression(1)}
	b := VariableDeclaration{Identifier: "b", IsConst: true, Type: Int16Type{}, Value: UInt64Expression(2)}
	c := VariableDeclaration{Identifier: "c", IsConst: true, Type: Int32Type{}, Value: UInt64Expression(3)}
	d := VariableDeclaration{Identifier: "d", IsConst: true, Type: Int64Type{}, Value: UInt64Expression(4)}
	e := VariableDeclaration{Identifier: "e", IsConst: true, Type: UInt8Type{}, Value: UInt64Expression(5)}
	f := VariableDeclaration{Identifier: "f", IsConst: true, Type: UInt16Type{}, Value: UInt64Expression(6)}
	g := VariableDeclaration{Identifier: "g", IsConst: true, Type: UInt32Type{}, Value: UInt64Expression(7)}
	h := VariableDeclaration{Identifier: "h", IsConst: true, Type: UInt64Type{}, Value: UInt64Expression(8)}

	expected := Program{
		AllDeclarations: []Declaration{a, b, c, d, e, f, g, h},
		Declarations:    map[string]Declaration{"a": a, "b": b, "c": c, "d": d, "e": e, "f": f, "g": g, "h": h},
	}

	NewWithT(t).Expect(actual).Should(Equal(expected))
}
