package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const MEMORY_SIZE = 1024

const (
	USE_HELP = iota
	USE_RUN

	USE_INVALID_FILE
	USE_TOO_MANY_PARAMS
	USE_NONE
)

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

func read(filepath string) string {
	dat, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return string(dat[:])
}

func getTokens(c rune) bool {
	return c == ' ' || c == '\t'
}

const (
	INST_OP = iota
	INST_TO
	INST_HALT
	INST_PRINT
)

type Instruction struct {
	typ int
	val interface{}
}

// isCommentInst if first token starts with '#'
func isCommentInst(tokens []string) bool {
	return len(tokens) == 0 || strings.HasPrefix(tokens[0], "#")
}

// isWord if is a valid text that can be used as a Symbol in the compiler
func isWord(code string) bool {
	for i, r := range code {
		isNumber := int(r) >= int('0') && int(r) <= int('9')
		hasValidChars := int(r) == int('_') && int(r) >= int('a') && int(r) <= int('z')

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
	OP_SUB = iota
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
		return OP_ADD, true
	case "/":
		return OP_ADD, true
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
func hasOperationInst(tokens []string) *Instruction {
	v := hasValue(tokens[0])
	if v == nil || v.typ != VAL_CONST {
		return nil
	}

	if tokens[1] != "=" {
		return nil
	}

	v1 := hasValue(tokens[2])
	if v1 == nil {
		return nil
	}

	op, exists := isOperator(tokens[3])
	if !exists {
		return nil
	}

	v2 := hasValue(tokens[4])
	if v2 == nil {
		return nil
	}

	if !isCommentInst(tokens[5:]) {
		return nil
	}

	return &Instruction{typ: INST_OP, val: Operation{v: *v, v1: *v1, v2: *v2, op: op}}
}

type IfInst struct {
	target string
	moveIf []interface{}
}

// hasToInst if first token is a 'to' and second is a label
func hasToInst(tokens []string) *Instruction {
	if tokens[0] != "to" {
		return nil
	}
	if !isWord(tokens[1]) {
		return nil
	}
	moveIf := compileIf(tokens[2:])
	if moveIf == nil {
		return nil
	}
	return &Instruction{typ: INST_TO, val: IfInst{target: tokens[1], moveIf: *moveIf}}
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
func compileIf(tokens []string) *[]interface{} {
	if tokens[0] != "if" {
		return nil
	}

	var params []interface{}
	for i, token := range tokens[1:] {
		if isCommentInst(tokens[i:]) {
			if ifInstOrder(i-1) != IFO_VAL {
				return nil
			}
			break
		}

		var p interface{}
		exists := false

		switch ifInstOrder(i) {
		case IFO_LOP:
			p, exists = hasLogicOperator(token)
			break
		case IFO_VAL:
			v := hasValue(token)
			if p != nil {
				exists = true
				p = *v
			}
			break
		case IFO_CMP:
			p, exists = hasComparison(token)
		}

		if !exists {
			return nil
		}
		params = append(params, p)
	}

	if ifInstOrder(len(tokens)-2) != IFO_VAL {
		return nil
	}

	return &params
}

func hasHaltInst(tokens []string) *Instruction {
	if tokens[0] != "halt" {
		return nil
	}
	if !isCommentInst(tokens[1:]) {
		return nil
	}
	return &Instruction{typ: INST_HALT}
}

func hasPrintInst(tokens []string) *Instruction {
	if tokens[0] != "print" {
		return nil
	}
	v1 := hasValue(tokens[1])
	if v1 == nil {
		return nil
	}

	return &Instruction{typ: INST_PRINT, val: *v1}
}

type InstFunc func(tokens []string) *Instruction

var INSTRUCTIONS = []InstFunc{hasOperationInst, hasToInst, hasHaltInst, hasPrintInst}

type Program struct {
	labels       map[string]int
	instructions []Instruction
}

func compilationError(line int, msg string) {
	fmt.Printf("[Erro de compilação : linha %d] %s.", line, msg)
	os.Exit(1)
}

// TODO: add better error message?
// TODO: add error line
func compile(code string) Program {
	var instructions []Instruction
	var labels map[string]int
	lines := strings.Split(code, "\n")

	for _, line := range lines {
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
				inst := f(tokens)
				if inst != nil {
					hasError = false
					instructions = append(instructions, *inst)
					break
				}
			}
			if hasError {
				compilationError(-1, "não foi possivel compilar essa instrução")
			}
		}
	}

	for _, inst := range instructions {
		if inst.typ == INST_TO {
			k := inst.val.(IfInst).target
			if _, ok := labels[k]; !ok {
				compilationError(-1, fmt.Sprintf("a label '%v' não foi definida", k))
			}
		}
	}

	return Program{instructions: instructions, labels: labels}
}

func valueFromMem(mem []int64, val InstValue) int64 {
	switch val.typ {
	case VAL_CONST:
		return val.val
	case VAL_VAR:
		return mem[val.val]
	case VAL_REF:
		if mem[val.val] < 0 || mem[val.val] > 1023 {
			executionError(-1, fmt.Sprintf("acesso de memória inválida: %d", val.val))
		}
		return mem[mem[val.val]]
	}
	// WARN: Must never reach this!!!
	return 0
}

func executionError(line int, msg string) {
	fmt.Printf("[Erro de execução : linha %d] %s.", line, msg)
	os.Exit(1)
}

func executeIf(mem []int64, inst IfInst) (res bool) {
	for i := 0; i < len(inst.moveIf); i += 4 {
		v1 := valueFromMem(mem, inst.moveIf[i].(InstValue))
		cp := inst.moveIf[i+1].(int64)
		v2 := valueFromMem(mem, inst.moveIf[i+2].(InstValue))

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
			switch inst.moveIf[i-1].(int64) {
			case LOP_AND:
				res = res && r
				break
			case LOP_OR:
				res = res || r
				break
			}
		}
	}

	return res
}

func execute(prog Program) {
	pc := 0
	mem := make([]int64, 1024)
	for pc < len(prog.instructions) {
		switch prog.instructions[pc].typ {
		case INST_OP:
			op := prog.instructions[pc].val.(Operation)
			v1 := valueFromMem(mem, op.v1)
			v2 := valueFromMem(mem, op.v2)
			switch op.op {
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
			if executeIf(mem, i) {
				pc = prog.labels[i.target]
			}
			break
		case INST_PRINT:
			val := prog.instructions[pc].val.(InstValue)
			switch val.typ {
			case VAL_CONST:
				fmt.Printf("$ %d\n", val.val)
				break
			case VAL_VAR:
				fmt.Printf("$ [ %d ] %d\n", val.val, valueFromMem(mem, val))
				break
			case VAL_REF:
				fmt.Printf("$ [ %d -> %d ] %d\n", val.val, mem[val.val], valueFromMem(mem, val))
				break
			}
			pc += 1
			break
		case INST_HALT:
			return
		}
	}
}

func run(filepath string) {
	execute(compile(read(filepath)))
}

func main() {
	switch getUse() {
	case USE_HELP:
	case USE_NONE:
		printUsage()
		break
	case USE_TOO_MANY_PARAMS:
		fmt.Println("Muitos parametros")
		os.Exit(1)
		break
	case USE_INVALID_FILE:
		fmt.Println("Arquivo deve ter a extensão .asm")
		os.Exit(1)
		break
	case USE_RUN:
		run(os.Args[1])
		break
	}
}
