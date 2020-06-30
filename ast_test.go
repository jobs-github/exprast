package exprast

import (
	"fmt"
	"testing"
)

type VarInterpreterTable map[string]func() error

func interpretTrue() error {
	return nil
}

func interpretFalse() error {
	return fmt.Errorf("false")
}

func interpretVar(table VarInterpreterTable) BooleanVarInterpreter {
	return func(key string) (error, error) {
		interpreter, ok := table[key]
		if !ok {
			return nil, fmt.Errorf("key `%v` missing", key)
		}
		if nil == interpreter {
			return nil, fmt.Errorf("key `%v` interpreter is nil", key)
		}
		return interpreter(), nil
	}
}

func Test_booleanExprAst_Interpret(t *testing.T) {
	booleanAst := DefaultBooleanExprAst()
	astskipsign, err := booleanAst.Build("($1 || $2 && $3 || $4)", true)
	if nil != err {
		t.Error(err)
		return
	}
	ast, err := booleanAst.Build("($1 || $2 && $3 || $4)", false)
	if nil != err {
		t.Error(err)
		return
	}

	type args struct {
		ast   *ExprAST
		table map[string]func() error
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"TestInterpretExprAST_1", args{astskipsign, map[string]func() error{
			"1": interpretTrue,
			"2": interpretFalse,
			"3": interpretTrue,
			"4": interpretFalse,
		}}, false, false},
		{"TestInterpretExprAST_2", args{astskipsign, map[string]func() error{
			"1": interpretFalse,
			"2": interpretFalse,
			"3": interpretFalse,
			"4": interpretFalse,
		}}, true, false},
		{"TestInterpretExprAST_3", args{astskipsign, map[string]func() error{
			"10": interpretTrue,
			"2":  interpretFalse,
			"3":  interpretTrue,
			"4":  interpretFalse,
		}}, false, true},
		{"TestInterpretExprAST_4", args{ast, map[string]func() error{
			"$1": interpretFalse,
			"$2": interpretFalse,
			"$3": interpretFalse,
			"$4": interpretFalse,
		}}, true, false},
		{"TestInterpretExprAST_5", args{ast, map[string]func() error{
			"$10": interpretTrue,
			"$2":  interpretFalse,
			"$3":  interpretTrue,
			"$4":  interpretFalse,
		}}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := booleanAst.Interpret(tt.args.ast, interpretVar(tt.args.table))
			if (err != nil) != tt.wantErr {
				t.Errorf("InterpretExprAST() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got != nil) != tt.want {
				t.Errorf("InterpretExprAST() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}

func interpreterVar(key string) (int64, error) {
	if "$num1" == key {
		return 920, nil
	}
	if "$num2" == key {
		return 902, nil
	}
	return 0, fmt.Errorf("unknown key: `%v`", key)
}

func Test_arithmeticExprAst_Interpret(t *testing.T) {
	arithmeticAst := DefaultArithmeticExprAst()
	astPureExpr, err := arithmeticAst.Build("920 + 50 + 50 * 27 - (902 - 102) / 40 + 260", false)
	if nil != err {
		t.Error(err)
		return
	}
	ast, err := arithmeticAst.Build("$num1 + 50 + 50 * 27 - ($num2 - 102) / 40 + 260", false)
	if nil != err {
		t.Error(err)
		return
	}

	type args struct {
		node         *ExprAST
		interpretVar ArithmeticVarInterpreter
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{"TestInterpretExprAST_1", args{astPureExpr, nil}, 2560, false},
		{"TestInterpretExprAST_2", args{ast, interpreterVar}, 2560, false},
		{"TestInterpretExprAST_3", args{ast, nil}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := arithmeticAst.Interpret(tt.args.node, tt.args.interpretVar)
			if (err != nil) != tt.wantErr {
				t.Errorf("InterpretExprAST() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("InterpretExprAST() = %v, want %v", got, tt.want)
			}
		})
	}
}
