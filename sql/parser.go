package sql

import (
	"fmt"
	"go-learner/slice"
	"strconv"
	"strings"
	"time"
)

const (
	WordIN  = "IN"
	WordAND = "AND"
	WordOR  = "OR"
)

type SelfConditionFunc func(value string, expected interface{}, ct ColumnType) bool
type RightConditionFunc func(left bool, right bool) bool

var SelfConditionFuncMap map[string]SelfConditionFunc
var RightConditionFuncMap map[string]RightConditionFunc

func init() {
	SelfConditionFuncMap = map[string]SelfConditionFunc{
		">":    greater,
		">=":   greaterOrEqual,
		"<":    less,
		"<=":   lessOrEqual,
		"=":    equal,
		"!=":   notEqual,
		WordIN: in,
	}

	RightConditionFuncMap = map[string]RightConditionFunc{
		WordAND: and,
		WordOR:  or,
	}
}

type Parser struct {
	hasShard bool
	schema   *TableSchema
	where    string
	params   []string
	paramIdx int
	numTypes []bool
	shardKey string
	cn       *conditionNode
}

type conditionNode struct {
	columnType ColumnType
	columnIdx  int
	expected   interface{}
	level      int
	selfFunc   SelfConditionFunc
	rightFunc  RightConditionFunc

	right *conditionNode
	child *conditionNode
}

func CreateParser(schema *TableSchema, shardKey string, where string, params []string) (*Parser, error) {
	if where == "" {
		return nil, nil
	}

	p := &Parser{
		schema:   schema,
		where:    where,
		params:   params,
		paramIdx: 0,
		shardKey: shardKey,
		numTypes: make([]bool, len(params)),
	}

	i := 0
	cn, err := p.parseCondition(p.where, &i, 0)
	if err != nil {
		return nil, err
	}

	if "" != nextWordBySpace(p.where, &i) {
		return nil, fmt.Errorf("invalid condition: %s", p.where)
	}

	if p.paramIdx != len(p.params) {
		return nil, fmt.Errorf("condition mismatch parameter num")
	}

	p.cn = cn
	return p, nil
}

func (p *Parser) Check(rowData []byte) bool {
	if rowData == nil {
		return false
	}

	return p.cn.check(p.schema, rowData)
}

func (p *Parser) ValidShard() bool {
	return p.hasShard
}

func (cn *conditionNode) setRight(operator string, right *conditionNode) bool {
	rightFunc, ok := RightConditionFuncMap[operator]
	if !ok {
		return false
	}

	cn.right = right
	cn.rightFunc = rightFunc

	return true
}

func (cn *conditionNode) set(operator string, columnIdx int, columnType ColumnType, expected interface{}) bool {
	selfFunc, ok := SelfConditionFuncMap[operator]
	if !ok {
		return false
	}

	cn.columnIdx = columnIdx
	cn.columnType = columnType
	cn.selfFunc = selfFunc
	cn.expected = expected

	return true
}

func (cn *conditionNode) check(schema *TableSchema, rowData []byte) bool {
	var left bool
	if cn.child != nil {
		left = cn.child.check(schema, rowData)

	} else {
		value := GetValueByIndex(schema, rowData, cn.columnIdx)
		left = cn.selfFunc(value, cn.expected, cn.columnType)
	}

	if cn.right == nil {
		return left
	}

	right := cn.right.check(schema, rowData)
	return cn.rightFunc(left, right)
}

func (p *Parser) replaceParam(word string, isNum bool) string {
	if word != "?" {
		return word
	}

	if p.paramIdx >= len(p.params) {
		return ""
	}

	ret := p.params[p.paramIdx]
	p.numTypes[p.paramIdx] = isNum
	p.paramIdx++

	return ret
}

func (p *Parser) parseCondition(sql string, start *int, level int) (*conditionNode, error) {
	word, sep := nextWord(sql, start, "`=!>< ")
	// check `name`=? like
	if sep == "`" {
		if word != "" {
			return nil, fmt.Errorf("sign '`' mismatched")
		}

		word, sep = nextWord(sql, start, "`=!>< ")
		if word == "" || sep == "" {
			return nil, fmt.Errorf("sign '`' mismatched")
		}

		if sep[0] != '`' {
			return nil, fmt.Errorf("sign '`' mismatched")
		}

		sep = sep[1:]

	} else if word == "" {
		return nil, fmt.Errorf("empty condition")
	}

	node := &conditionNode{columnIdx: -1, level: level}
	schema := p.schema
	if sep == "" {
		if word[0] == '(' {
			i := 1
			child, err := p.parseCondition(word, &i, level+1)
			if err != nil {
				return nil, err
			}

			node.child = child

		} else {
			//特殊条件，仅支持：x IN (a, b)
			column := schema.GetColumnSchema(word)
			if column == nil {
				return nil, fmt.Errorf("invalid column: %s", word)
			}

			word = nextWordBySpace(sql, start)
			word = strings.ToUpper(word)
			if word != WordIN {
				return nil, fmt.Errorf("invalid IN condition: %s", sql)
			}

			i := nextNonEmptyIndex(sql, *start)
			if i < 0 || sql[i] != '(' {
				return nil, fmt.Errorf("invalid IN condition: %s", sql)
			}

			i++
			v := make([]string, 0, 4)
			for {
				word, sep = nextWord(sql, &i, ", ")
				if word == "" {
					return nil, fmt.Errorf("invalid IN condition: %s", sql)
				}

				word = p.replaceParam(word, column.IsNumber)
				if word == "" {
					return nil, fmt.Errorf("condition mismatch parameter num")
				}

				v = append(v, word)
				if sep != "," {
					break
				}
			}

			cv := castValueSlice(v, column.Type)
			if cv == nil {
				return nil, fmt.Errorf("invalid column type: %s", sql)
			}

			i++
			*start = i

			if !node.set(WordIN, column.Index, column.Type, cv) {
				return nil, fmt.Errorf("nonsupport operator: %s", WordIN)
			}
		}

	} else {
		column := schema.GetColumnSchema(word)
		if column == nil {
			return nil, fmt.Errorf("invalid column: %s", word)
		}

		word = nextWordBySpace(sql, start)
		if word == "" {
			return nil, fmt.Errorf("invalid condition: %s", sql)
		}

		word = p.replaceParam(word, column.IsNumber)
		if word == "" {
			return nil, fmt.Errorf("condition mismatch parameter num")
		}

		cv := castValue(word, column.Type)
		if cv == nil {
			return nil, fmt.Errorf("invalid column type: %s", sql)
		}

		if !node.set(sep, column.Index, column.Type, cv) {
			return nil, fmt.Errorf("nonsupport operator: %s", sep)
		}

		if column.Index == schema.ShardIndex {
			if word != p.shardKey || sep != "=" {
				return nil, fmt.Errorf("invalid shard condition: %s", word)
			}

			if level != 0 {
				return nil, fmt.Errorf("shard key MUST be level 0: %s", word)
			}

			p.hasShard = true
		}
	}

	i := *start
	word = nextWordBySpace(sql, start)
	word = strings.ToUpper(word)
	switch word {
	case WordAND:
		right, err := p.parseCondition(sql, start, level)
		if err != nil {
			return nil, err
		}

		if !node.setRight(word, right) {
			return nil, fmt.Errorf("nonsupport operator: %s", word)
		}

	case WordOR:
		if level < 1 {
			return nil, fmt.Errorf("OR can not be in level 0: %s", sql)
		}

		//用child来保证运算优先级，OR < AND
		rightChild, err := p.parseCondition(sql, start, level+1)
		if err != nil {
			return nil, err
		}

		if !node.setRight(word, &conditionNode{columnIdx: -1, level: level, child: rightChild}) {
			return nil, fmt.Errorf("nonsupport operator: %s", word)
		}

	default:
		*start = i
	}

	return node, nil
}

func nextWord(value string, start *int, separators string) (string, string) {
	s := nextNonEmptyIndex(value, *start)
	if s < 0 {
		return "", ""
	}

	n := len(value)
	e := n

	var sep []byte
	nBracket := 0
	quote := byte(0)

	var i int
	for i = s; i < n; i++ {
		c := value[i]

		if sep != nil {
			if strings.IndexByte(separators, c) > -1 {
				if c != ' ' {
					sep = append(sep, c)
				}

			} else {
				break
			}

		} else if quote == 0 {
			if nBracket == 0 && strings.IndexByte(separators, c) > -1 {
				e = i

				sep = make([]byte, 0, 2)
				if c != ' ' {
					sep = append(sep, c)
				}

			} else if c == '(' {
				nBracket++

			} else if c == ')' {
				nBracket--
				if nBracket < 0 {
					e = i
					break
				}

			} else if c == '\'' {
				quote = c
			}

		} else if c == quote {
			quote = 0
		}
	}

	if quote != 0 {
		return "", ""
	}

	*start = i
	for i = e - 1; i >= s; i-- {
		if value[i] != ' ' {
			e = i + 1
			break
		}
	}

	if value[s] == '\'' && value[e-1] == '\'' {
		s++
		e--
	}

	return value[s:e], slice.ByteSlice2String(sep)
}

func nextWordBySpace(value string, start *int) string {
	s := nextNonEmptyIndex(value, *start)
	if s < 0 {
		return ""
	}

	n := len(value)
	e := n

	nBracket := 0
	quote := byte(0)

	var i int
	for i = s; i < n; i++ {
		c := value[i]

		if quote == 0 {
			if nBracket == 0 && c == ' ' {
				e = i
				break

			} else if c == '(' {
				nBracket++

			} else if c == ')' {
				nBracket--
				if nBracket < 0 {
					e = i
					break
				}

			} else if c == '\'' {
				quote = c
			}

		} else if c == quote {
			quote = 0
		}
	}

	*start = i
	//for i = e - 1; i > s; i-- {
	//	if value[i] != ' ' {
	//		e = i + 1
	//		break
	//	}
	//}

	if value[s] == '\'' && value[e-1] == '\'' {
		s++
		e--
	}

	return value[s:e]
}

func nextNonEmptyIndex(value string, start int) int {
	n := len(value)
	for i := start; i < n; i++ {
		if value[i] != ' ' {
			return i
		}
	}

	return -1
}

func greater(value string, expected interface{}, ct ColumnType) bool {
	actual := castValue(value, ct)
	if actual == nil {
		return false
	}

	switch ct {
	case ColumnTypeInt:
		return actual.(int64) > expected.(int64)

	case ColumnTypeFloat:
		return actual.(float64) > expected.(float64)

	case ColumnTypeTime:
		return actual.(int64) > expected.(int64)

	default:
		return actual.(string) > expected.(string)
	}
}

func greaterOrEqual(value string, expected interface{}, ct ColumnType) bool {
	actual := castValue(value, ct)
	if actual == nil {
		return false
	}

	switch ct {
	case ColumnTypeInt:
		return actual.(int64) >= expected.(int64)

	case ColumnTypeFloat:
		return actual.(float64) >= expected.(float64)

	case ColumnTypeTime:
		return actual.(int64) >= expected.(int64)

	default:
		return actual.(string) >= expected.(string)
	}
}

func less(value string, expected interface{}, ct ColumnType) bool {
	actual := castValue(value, ct)
	if actual == nil {
		return false
	}

	switch ct {
	case ColumnTypeInt:
		return actual.(int64) < expected.(int64)

	case ColumnTypeFloat:
		return actual.(float64) < expected.(float64)

	case ColumnTypeTime:
		return actual.(int64) < expected.(int64)

	default:
		return actual.(string) < expected.(string)
	}
}

func lessOrEqual(value string, expected interface{}, ct ColumnType) bool {
	actual := castValue(value, ct)
	if actual == nil {
		return false
	}

	switch ct {
	case ColumnTypeInt:
		return actual.(int64) <= expected.(int64)

	case ColumnTypeFloat:
		return actual.(float64) <= expected.(float64)

	case ColumnTypeTime:
		return actual.(int64) <= expected.(int64)

	default:
		return actual.(string) <= expected.(string)
	}
}

func equal(value string, expected interface{}, ct ColumnType) bool {
	actual := castValue(value, ct)
	if actual == nil {
		return false
	}

	switch ct {
	case ColumnTypeInt:
		return actual.(int64) == expected.(int64)

	case ColumnTypeFloat:
		return actual.(float64) == expected.(float64)

	case ColumnTypeTime:
		return actual.(int64) == expected.(int64)

	default:
		return actual.(string) == expected.(string)
	}
}

func notEqual(value string, expected interface{}, ct ColumnType) bool {
	actual := castValue(value, ct)
	if actual == nil {
		return false
	}

	return actual != expected
}

func in(value string, expected interface{}, ct ColumnType) bool {
	actual := castValue(value, ct)
	if actual == nil {
		return false
	}

	for _, v := range expected.([]interface{}) {
		if v == actual {
			return true
		}
	}

	return false
}

func and(left bool, right bool) bool {
	return left && right
}

func or(left bool, right bool) bool {
	return left || right
}

func castValue(value string, ct ColumnType) interface{} {
	switch ct {
	case ColumnTypeInt:
		iValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil
		}

		return iValue

	case ColumnTypeFloat:
		fValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil
		}

		return fValue

	case ColumnTypeTime:
		tValue, err := time.Parse("2006-01-02 15:04:05", value)
		if err != nil {
			return nil
		}

		return tValue.Unix()

	default:
		return value
	}
}

func castValueSlice(vs []string, ct ColumnType) []interface{} {
	n := len(vs)
	ret := make([]interface{}, n)
	for i := 0; i < n; i++ {
		ret[i] = castValue(vs[i], ct)
	}

	return ret
}
