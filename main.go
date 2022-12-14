package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var i18n *message.Printer

const (
	I18N_ERR_PROG_TOO_MANY_PARMS = "too many parameters"
	I18N_ERR_PROG_NEED_ASM_EXT   = "the file must have an .asm extension"

	I18N_ERR_OP_ONLY_ONE_LEFT_VAL = "must have only one operation left value, but received"
	I18N_ERR_OP_LEFT_VAL_INVALID  = "invalid operation left value"
	I18N_ERR_OP_RIGHT_VAL_INVALID = "invalid operation first right value"
	I18N_ERR_OP_OP_INVALID        = "invalid operation, expecting (+,-,/,*), but received"
	I18N_ERR_OP_2_VAL_INVALID     = "invalid operation second right valuie"
	I18N_ERR_OP_NOT_ENDED         = "expecting operation to finish, but received"

	I18N_ERR_TO_INVALID_WORD = "expecting valid word, but received"

	I18N_ERR_IF_EXPECT_IF             = "expecting word 'if', but received"
	I18N_ERR_IF_EXPECT_COMP_OP        = "expecting logic operator(==, !=, >=, <=, >, <), but received"
	I18N_ERR_IF_EXPECT_LOGIC_OP       = "expecting comparison(&&, ||), but received"
	I18N_ERR_IF_EXPECT_VALUE          = "expecting a value, but received"
	I18N_ERR_IF_EXPECT_END_WITH_VALUE = "expecting ending with a value, mas recbeu"

	I18N_ERR_WRITE_EXPECT_VALUE   = "expecting a value, but received"
	I18N_ERR_WRITE_ONLY_ONE_PARAM = "expecting only one value as a parameter, but received"

	I18N_ERR_READ_EXPECT_VALUE  = "expecting a value, but received"
	I18N_ERR_READ_INVALID_WORD  = "expecting valid word, but received"
	I18N_ERR_READ_NOTHING       = "trying to read when there is no more input"
	I18N_ERR_READ_NO_ELSE_LABEL = "no else label"

	I18N_COMPILE_ERR_TEMPLATE = "[Compilation error: line %d] %v."

	I18N_COMPILE_ERR_INST_NOT_FOUND  = "instruction not found"
	I18N_COMPILE_ERR_LABEL_NOT_FOUND = "label not defined"

	I18N_EXEC_ERR_INVALID_MEMORY_ACCESS = "invalid memory access"

	I18N_EXEC_ERR_TEMPLATE = "[Execution error: line %d] %v."

	I18N_INPUT_ERR_TEMPLATE = "On file '%s' line %d was not possible to convert '%s' into a number\n"
)

const MEMORY_SIZE = 1024

const (
	USE_HELP = iota
	USE_RUN

	USE_INVALID_FILE
	USE_TOO_MANY_PARAMS
	USE_NONE
)

func formatError(typ string, message string, problem interface{}) error {
	return fmt.Errorf("<%v> %v: %v", typ, message, problem)
}

func getUse() int {
	if len(os.Args) == 1 {
		return USE_HELP
	}
	if len(os.Args) != 2 {
		return USE_TOO_MANY_PARAMS
	}
	if strings.ToLower(os.Args[1]) == "help" {
		return USE_HELP
	}
	if !strings.HasSuffix(os.Args[1], ".asm") {
		return USE_INVALID_FILE
	}

	return USE_RUN
}

func printUsage() {
	fmt.Println("")
}

func read(filepath string) (string, error) {
	dat, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(dat[:]), nil
}

func getTokens(c rune) bool {
	return c == ' ' || c == '\t'
}

const (
	INST_OP = iota
	INST_TO
	INST_WRITE
	INST_READ
)

type Instruction struct {
	typ  int
	val  interface{}
	line int
}

// isCommentInst if first token starts with '#'
func isCommentInst(tokens []string) bool {
	return len(tokens) == 0 || strings.HasPrefix(tokens[0], "#")
}

// isWord if is a valid text that can be used as a Symbol in the compiler
func isWord(code string) bool {
	if len(code) <= 1 {
		return false
	}
	for i, r := range code {
		isNumber := int(r) >= int('0') && int(r) <= int('9')
		hasValidChars := int(r) == int('_') || int(r) >= int('a') && int(r) <= int('z')

		if i == 0 && isNumber {
			return false
		}
		if !hasValidChars && !isNumber {
			return false
		}
	}
	return true
}

// hasLabel if first token ends with ':'
func hasLabel(tokens []string) (string, bool) {
	exists := isWord(tokens[0][:len(tokens[0])-1]) && strings.HasSuffix(tokens[0], ":") && isCommentInst(tokens[1:])
	if exists {
		return tokens[0][:len(tokens[0])-1], true
	}
	return "", false
}

func isRegister(code string) (int, bool) {
	i, err := strconv.Atoi(code[0:])
	if err != nil {
		return 0, false
	}
	if i < 0 || i > 1023 {
		return 0, false
	}

	return i, true
}

func isVariable(code string) (int, bool) {
	if strings.HasPrefix(code, "$") {
		return isRegister(code[1:])
	}
	return 0, false
}

func isReference(code string) (int, bool) {
	if strings.HasPrefix(code, "&") {
		return isRegister(code[1:])
	}
	return 0, false
}

const (
	VAL_CONST = iota
	VAL_VAR
	VAL_REF
)

type InstValue struct {
	typ int
	val int64
}

// hasValue get value stored from constant|variable|reference
func hasValue(code string) *InstValue {
	v, exists := isVariable(code)
	if exists {
		return &InstValue{typ: VAL_VAR, val: int64(v)}
	}
	r, exists := isReference(code)
	if exists {
		return &InstValue{typ: VAL_REF, val: int64(r)}
	}
	c, err := strconv.ParseInt(code, 10, 64)
	if err == nil {
		return &InstValue{typ: VAL_CONST, val: c}
	}

	return nil
}

const (
	OP_UNI = iota
	OP_SUB
	OP_ADD
	OP_MUL
	OP_DIV
)

func isOperator(code string) (int, bool) {
	switch code {
	case "-":
		return OP_SUB, true
	case "+":
		return OP_ADD, true
	case "*":
		return OP_MUL, true
	case "/":
		return OP_DIV, true
	default:
		return -1, false
	}
}

type Operation struct {
	v  InstValue
	v1 InstValue
	v2 InstValue
	op int
}

// hasOperationInst if follow this pattern `$v = $1 {-, +, *, /} $2`
func hasOperationInst(tokens []string) (*Instruction, error) {
	if tokens[1] != "=" {
		for i, token := range tokens[1:] {
			if token == "=" {
				return nil, formatError("op", I18N_ERR_OP_ONLY_ONE_LEFT_VAL, tokens[:i])
			}
		}

		return nil, nil
	}

	v := hasValue(tokens[0])
	if v == nil || v.typ == VAL_CONST {
		return nil, formatError("op", I18N_ERR_OP_LEFT_VAL_INVALID, tokens[0])
	}

	v1 := hasValue(tokens[2])
	if v1 == nil {
		return nil, formatError("op", I18N_ERR_OP_RIGHT_VAL_INVALID, tokens[2])
	}

	if isCommentInst(tokens[3:]) {
		return &Instruction{typ: INST_OP, val: Operation{v: *v, v1: *v1, op: OP_UNI}}, nil
	}

	op, exists := isOperator(tokens[3])
	if !exists {
		return nil, formatError("op", I18N_ERR_OP_OP_INVALID, tokens[3])
	}

	v2 := hasValue(tokens[4])
	if v2 == nil {
		return nil, formatError("op", I18N_ERR_OP_2_VAL_INVALID, tokens[4])
	}

	if !isCommentInst(tokens[5:]) {
		return nil, formatError("op", I18N_ERR_OP_NOT_ENDED, tokens[5:])
	}

	return &Instruction{typ: INST_OP, val: Operation{v: *v, v1: *v1, v2: *v2, op: op}}, nil
}

type IfInst struct {
	target string
	moveIf []interface{}
}

// hasToInst if first token is a 'to' and second is a label
func hasToInst(tokens []string) (*Instruction, error) {
	if tokens[0] != "to" {
		return nil, nil
	}
	if !isWord(tokens[1]) {
		return nil, formatError("to", I18N_ERR_TO_INVALID_WORD, tokens[1])
	}

	var moveIf []interface{} = nil
	var err error
	if len(tokens) > 2 {
		if tokens[2] == "if" {
			moveIf, err = compileIf(tokens[2:])
			if err != nil {
				return nil, err
			}
		} else if !isCommentInst(tokens[2:]) {
			return nil, formatError("to", I18N_ERR_TO_INVALID_WORD, tokens[2:])
		}
	}
	return &Instruction{typ: INST_TO, val: IfInst{target: tokens[1], moveIf: moveIf}}, nil
}

const (
	COMP_EQ = iota
	COMP_DF
	COMP_GT
	COMP_LT
	COMP_GE
	COMP_LE
)

// hasComparison try to get the comparison from code
func hasComparison(code string) (int64, bool) {
	switch code {
	case "==":
		return COMP_EQ, true
	case "!=":
		return COMP_DF, true
	case ">":
		return COMP_GT, true
	case "<":
		return COMP_LT, true
	case ">=":
		return COMP_GE, true
	case "<=":
		return COMP_LE, true
	default:
		return -1, false
	}
}

const (
	LOP_AND = iota
	LOP_OR
)

// hasLogicOperator try to get the logic operator from code
func hasLogicOperator(code string) (int64, bool) {
	switch code {
	case "&&":
		return LOP_AND, true
	case "||":
		return LOP_OR, true
	default:
		return -1, false
	}
}

const (
	IFO_LOP = iota
	IFO_VAL
	IFO_CMP
)

// ifInstOrder try to get the if instruction order from code
func ifInstOrder(i int) int {
	if i > 0 && i%4 == 0 {
		return IFO_LOP
	} else if i%2 == 0 {
		return IFO_VAL
	}
	return IFO_CMP
}

// compileIf if follow this pattern `if $1 {==, !=, >, <, >=, <=} $2 {&&, ||} ... then $n`
func compileIf(tokens []string) ([]interface{}, error) {
	if tokens[0] != "if" {
		return nil, formatError("if", I18N_ERR_IF_EXPECT_IF, tokens[0])
	}

	var params []interface{}
	for i, token := range tokens[1:] {
		if isCommentInst(tokens[i:]) {
			if ifInstOrder(i-1) != IFO_VAL {
				return nil, formatError("if", I18N_ERR_IF_EXPECT_END_WITH_VALUE, tokens[i-1])
			}
			break
		}

		var p interface{}
		var exists bool
		var err error

		switch ifInstOrder(i) {
		case IFO_LOP:
			p, exists = hasLogicOperator(token)
			if !exists {
				err = formatError("if", I18N_ERR_IF_EXPECT_LOGIC_OP, token)
			}
			break
		case IFO_VAL:
			v := hasValue(token)
			if v != nil {
				exists = true
				p = *v
			} else {
				err = formatError("if", I18N_ERR_IF_EXPECT_VALUE, token)
			}
			break
		case IFO_CMP:
			p, exists = hasComparison(token)
			if !exists {
				err = formatError("if", I18N_ERR_IF_EXPECT_COMP_OP, token)
			}
		}

		if err != nil {
			return nil, err
		}
		params = append(params, p)
	}

	if ifInstOrder(len(tokens)-2) != IFO_VAL {
		return nil, formatError("if", I18N_ERR_IF_EXPECT_END_WITH_VALUE, tokens[len(tokens)-1])
	}

	return params, nil
}

func hasWriteInst(tokens []string) (*Instruction, error) {
	if tokens[0] != "write" {
		return nil, nil
	}
	v1 := hasValue(tokens[1])
	if v1 == nil {
		return nil, formatError("write", I18N_ERR_WRITE_EXPECT_VALUE, tokens[1])
	}
	if !isCommentInst(tokens[2:]) {
		return nil, formatError("write", I18N_ERR_WRITE_ONLY_ONE_PARAM, tokens[2:])
	}

	return &Instruction{typ: INST_WRITE, val: *v1}, nil
}

type ReadInst struct {
	target    InstValue
	elseLabel string
}

// hasReadInst will follow the pattern `read $ label?`
func hasReadInst(tokens []string) (*Instruction, error) {
	if tokens[0] != "read" {
		return nil, nil
	}

	t := hasValue(tokens[1])
	if t == nil {
		return nil, formatError("read", I18N_ERR_READ_EXPECT_VALUE, tokens[1])
	}
	label := ""
	if len(tokens) > 2 {
		if !isWord(tokens[2]) {
			return nil, formatError("to", I18N_ERR_READ_INVALID_WORD, tokens[1])
		}
		label = tokens[2]
	}

	return &Instruction{typ: INST_READ, val: ReadInst{target: *t, elseLabel: label}}, nil
}

type InstFunc func(tokens []string) (*Instruction, error)

// WARN: the order here matters, check the first error for `hasOperationInst` and `hasToInst` to understand why.
var INSTRUCTIONS = []InstFunc{hasToInst, hasWriteInst, hasOperationInst, hasReadInst}

type Program struct {
	labels       map[string]int
	instructions []Instruction
}

func compilationError(line int, err error) error {
	return fmt.Errorf(I18N_COMPILE_ERR_TEMPLATE, line, err)
}

func compile(code string) (*Program, error) {
	var instructions []Instruction
	labels := make(map[string]int)
	lines := strings.Split(code, "\n")

	for iline, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		tokens := strings.FieldsFunc(line, getTokens)
		if isCommentInst(tokens) {
			continue
		}
		if label, exists := hasLabel(tokens); exists {
			labels[label] = len(instructions)
		} else {
			hasError := true
			for _, f := range INSTRUCTIONS {
				inst, err := f(tokens)
				if err != nil {
					return nil, compilationError(iline, err)
				}
				if inst != nil {
					hasError = false
					inst.line = iline + 1
					instructions = append(instructions, *inst)
					break
				}
			}
			if hasError {
				return nil, compilationError(iline, formatError("?", I18N_COMPILE_ERR_INST_NOT_FOUND, tokens))
			}
		}
	}

	for _, inst := range instructions {
		if inst.typ == INST_TO {
			k := inst.val.(IfInst).target
			if _, ok := labels[k]; !ok {
				return nil, compilationError(inst.line, formatError("label", I18N_COMPILE_ERR_LABEL_NOT_FOUND, k))
			}
		} else if inst.typ == INST_READ {
			k := inst.val.(ReadInst).elseLabel
			if k != "" {
				if _, ok := labels[k]; !ok {
					return nil, compilationError(inst.line, formatError("label", I18N_COMPILE_ERR_LABEL_NOT_FOUND, k))
				}
			}
		}
	}

	return &Program{instructions: instructions, labels: labels}, nil
}

func valueFromMem(mem []int64, val InstValue) (int64, error) {
	switch val.typ {
	case VAL_CONST:
		return val.val, nil
	case VAL_VAR:
		return mem[val.val], nil
	case VAL_REF:
		if mem[val.val] < 0 || mem[val.val] > 1023 {
			return 0, formatError("[memory]", I18N_EXEC_ERR_INVALID_MEMORY_ACCESS, val.val)
		}
		return mem[mem[val.val]], nil
	}
	panic("IMPOSSIBLE")
}

func executionError(line int, err error) error {
	return fmt.Errorf(I18N_EXEC_ERR_TEMPLATE, line, err)
}

func executeIf(mem []int64, inst Instruction) (res bool, err error) {
	moveIf := inst.val.(IfInst).moveIf
	if moveIf == nil {
		return true, nil
	}
	for i := 0; i < len(moveIf); i += 4 {
		v1, err := valueFromMem(mem, moveIf[i].(InstValue))
		if err != nil {
			return false, executionError(inst.line, err)
		}
		cp := moveIf[i+1].(int64)
		v2, err := valueFromMem(mem, moveIf[i+2].(InstValue))
		if err != nil {
			return false, executionError(inst.line, err)
		}

		var r bool
		switch cp {
		case COMP_EQ:
			r = v1 == v2
			break
		case COMP_DF:
			r = v1 != v2
			break
		case COMP_GT:
			r = v1 > v2
			break
		case COMP_LT:
			r = v1 < v2
			break
		case COMP_GE:
			r = v1 >= v2
			break
		case COMP_LE:
			r = v1 <= v2
			break
		}

		if i > 0 {
			switch moveIf[i-1].(int64) {
			case LOP_AND:
				res = res && r
				break
			case LOP_OR:
				res = res || r
				break
			}
		} else {
			res = r
		}
	}

	return res, nil
}

type WriteResult struct {
	val InstValue
	ref int64
	res int64
}

func (w WriteResult) ToString() string {
	switch w.val.typ {
	case VAL_CONST:
		return fmt.Sprintf("$ %d", w.val.val)
	case VAL_VAR:
		return fmt.Sprintf("$ [ %d ] %d", w.val.val, w.res)
	case VAL_REF:
		return fmt.Sprintf("$ [ %d -> %d ] %d", w.val.val, w.ref, w.res)
	}
	panic("IMPOSSIBLE")
}

func execute(prog Program, input []int64) ([]WriteResult, error) {
	var results []WriteResult
	pc := 0
	rc := 0
	mem := make([]int64, 1024)

	for pc < len(prog.instructions) {
		switch prog.instructions[pc].typ {
		case INST_OP:
			op := prog.instructions[pc].val.(Operation)
			v1, err := valueFromMem(mem, op.v1)
			if err != nil {
				return results, executionError(prog.instructions[pc].line, err)
			}
			v2, err := valueFromMem(mem, op.v2)
			if err != nil {
				return results, executionError(prog.instructions[pc].line, err)
			}
			switch op.op {
			case OP_UNI:
				if op.v.typ == VAL_REF {
					mem[mem[op.v.val]] = v1
				} else {
					mem[op.v.val] = v1
				}
				break
			case OP_ADD:
				if op.v.typ == VAL_REF {
					mem[mem[op.v.val]] = v1 + v2
				} else {
					mem[op.v.val] = v1 + v2
				}
				break
			case OP_SUB:
				if op.v.typ == VAL_REF {
					mem[mem[op.v.val]] = v1 - v2
				} else {
					mem[op.v.val] = v1 - v2
				}
				break
			case OP_MUL:
				if op.v.typ == VAL_REF {
					mem[mem[op.v.val]] = v1 * v2
				} else {
					mem[op.v.val] = v1 * v2
				}
				break
			case OP_DIV:
				if op.v.typ == VAL_REF {
					mem[mem[op.v.val]] = v1 / v2
				} else {
					mem[op.v.val] = v1 / v2
				}
				break
			}
			pc += 1
		case INST_TO:
			i := prog.instructions[pc].val.(IfInst)
			c, err := executeIf(mem, prog.instructions[pc])
			if err != nil {
				return results, err
			}
			if c {
				pc = prog.labels[i.target]
			} else {
				pc += 1
			}
			break
		case INST_WRITE:
			val := prog.instructions[pc].val.(InstValue)
			switch val.typ {
			case VAL_CONST:
				results = append(results, WriteResult{val: val})
				break
			case VAL_VAR:
				v, err := valueFromMem(mem, val)
				if err != nil {
					return results, executionError(prog.instructions[pc].line, err)
				} else {
					results = append(results, WriteResult{val: val, res: v})
				}
				break
			case VAL_REF:
				v, err := valueFromMem(mem, val)
				if err != nil {
					return results, executionError(prog.instructions[pc].line, err)
				} else {
					results = append(results, WriteResult{val: val, ref: mem[val.val], res: v})
				}
				break
			}
			pc += 1
			break
		case INST_READ:
			in := prog.instructions[pc].val.(ReadInst)
			if rc < len(input) {
				if in.target.typ == VAL_REF {
					mem[mem[in.target.val]] = input[rc]
				} else {
					mem[in.target.val] = input[rc]
				}
				rc += 1
				pc += 1
			} else {
				if in.elseLabel == "" {
					return results, executionError(prog.instructions[pc].line, formatError("[read]", I18N_ERR_READ_NOTHING, I18N_ERR_READ_NO_ELSE_LABEL))
				}
				pc = prog.labels[in.elseLabel]
			}
			break
		}
	}

	return results, nil
}

func Run(source string, input string) ([]WriteResult, error) {
	sdat, err := read(source)
	if err != nil {
		return []WriteResult{}, err
	}

	var ivalues []int64
	if input != "" {
		idat, err := read(input)
		if err != nil {
			return []WriteResult{}, err
		}
		lines := strings.Split(idat, "\n")
		ivalues = make([]int64, len(lines))
		for i, v := range lines {
			v = strings.TrimSpace(v)
			ivalues[i], err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				i18n.Printf(I18N_INPUT_ERR_TEMPLATE, input, i+1, v)
				os.Exit(1)
			}
		}
	}
	prog, err := compile(sdat)
	if err != nil {
		return []WriteResult{}, err
	}
	return execute(*prog, ivalues)
}

func init() {
	message.SetString(language.BrazilianPortuguese, I18N_ERR_PROG_TOO_MANY_PARMS, "Muitos parametros")
	message.SetString(language.BrazilianPortuguese, I18N_ERR_PROG_NEED_ASM_EXT, "Arquivo deve ter a extens??o .asm")

	message.SetString(language.BrazilianPortuguese, I18N_ERR_OP_ONLY_ONE_LEFT_VAL, "deve ter apenas um valor no lado esquerdo da opera????o, mas recebeu")
	message.SetString(language.BrazilianPortuguese, I18N_ERR_OP_LEFT_VAL_INVALID, "valor esquerdo da opera????o inv??lido")
	message.SetString(language.BrazilianPortuguese, I18N_ERR_OP_RIGHT_VAL_INVALID, "primeiro valor direito da opera????o inv??lido")
	message.SetString(language.BrazilianPortuguese, I18N_ERR_OP_OP_INVALID, "opera????o inv??lida, esperando(+,-,/,*), mas recebeu")
	message.SetString(language.BrazilianPortuguese, I18N_ERR_OP_2_VAL_INVALID, "segundo valor direito da opera????o inv??lido")
	message.SetString(language.BrazilianPortuguese, I18N_ERR_OP_NOT_ENDED, "esperando finalizar opera????o, mas recebeu")

	message.SetString(language.BrazilianPortuguese, I18N_ERR_TO_INVALID_WORD, "esperando uma palavra v??lida, mas recebeu")

	message.SetString(language.BrazilianPortuguese, I18N_ERR_IF_EXPECT_IF, "esperando a palavra 'if', mas recebeu")
	message.SetString(language.BrazilianPortuguese, I18N_ERR_IF_EXPECT_COMP_OP, "esperando um operador l??gico(==, !=, >=, <=, >, <), mas recebeu")
	message.SetString(language.BrazilianPortuguese, I18N_ERR_IF_EXPECT_LOGIC_OP, "esperando uma compara????o(&&, ||), mas recebeu")
	message.SetString(language.BrazilianPortuguese, I18N_ERR_IF_EXPECT_VALUE, "esperando um valor, mas recebeu")
	message.SetString(language.BrazilianPortuguese, I18N_ERR_IF_EXPECT_END_WITH_VALUE, "esperando terminar com um valor, mas recebeu")

	message.SetString(language.BrazilianPortuguese, I18N_ERR_WRITE_EXPECT_VALUE, "espera um valor, mas recebeu")
	message.SetString(language.BrazilianPortuguese, I18N_ERR_WRITE_ONLY_ONE_PARAM, "recebe apenas um valor como parametro, mas recebeu")

	message.SetString(language.BrazilianPortuguese, I18N_ERR_READ_EXPECT_VALUE, "espera um valor, mas recebeu")
	message.SetString(language.BrazilianPortuguese, I18N_ERR_READ_INVALID_WORD, "esperando uma palavra v??lida, mas recebeu")
	message.SetString(language.BrazilianPortuguese, I18N_ERR_READ_NOTHING, "tentando ler um arquivo que j?? acabou")
	message.SetString(language.BrazilianPortuguese, I18N_ERR_READ_NO_ELSE_LABEL, "sem uma label de sa??da")

	message.SetString(language.BrazilianPortuguese, I18N_COMPILE_ERR_TEMPLATE, "[Erro de compila????o : linha %d] %v.")

	message.SetString(language.BrazilianPortuguese, I18N_COMPILE_ERR_INST_NOT_FOUND, "instru????o n??o identificada")
	message.SetString(language.BrazilianPortuguese, I18N_COMPILE_ERR_LABEL_NOT_FOUND, "label n??o foi definida")

	message.SetString(language.BrazilianPortuguese, I18N_EXEC_ERR_INVALID_MEMORY_ACCESS, "acesso de mem??ria inv??lido")

	message.SetString(language.BrazilianPortuguese, I18N_EXEC_ERR_TEMPLATE, "[Erro de execu????o : linha %d] %v.")

	message.SetString(language.BrazilianPortuguese, I18N_INPUT_ERR_TEMPLATE, "No arquivo de entrada '%s' na linha %d n??o foi possivel converter '%s' em um n??mero\n")

	i18n = message.NewPrinter(language.BrazilianPortuguese)
}

func print(r WriteResult) {
	fmt.Println(r.ToString())
}

func main() {
	switch getUse() {
	case USE_HELP:
	case USE_NONE:
		printUsage()
		break
	case USE_TOO_MANY_PARAMS:
		i18n.Println(I18N_ERR_PROG_TOO_MANY_PARMS)
		os.Exit(1)
		break
	case USE_INVALID_FILE:
		i18n.Println(I18N_ERR_PROG_NEED_ASM_EXT)
		os.Exit(1)
		break
	case USE_RUN:
		input := ""
		if len(os.Args) > 2 {
			input = os.Args[2]
		}
		res, err := Run(os.Args[1], input)
		for _, r := range res {
			print(r)
		}
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		break
	}
}
