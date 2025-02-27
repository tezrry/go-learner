package generator

import (
	"fmt"
	"reflect"
	"strings"

	"csv/gameconfig/builder/lexer"
)

const (
	LF = '\n'
)

type Generator struct {
	tb          *lexer.Table
	importCache map[string]reflect.Type
	sb          strings.Builder
}

func NewGenerator(tb *lexer.Table) *Generator {
	inst := &Generator{
		tb:          tb,
		importCache: make(map[string]reflect.Type, 16),
	}

	for _, goType := range lexer.SimpleTypeMapping {
		inst.ImportPkg(goType.typ)
	}

	return inst
}

func (inst *Generator) Save() {
	inst.writeLine("package " + inst.tb.name)
	inst.writeImport()
	inst.writeTableDef()
	fmt.Println(inst.sb.String())
}

func (inst *Generator) ImportPkg(t reflect.Type) {
	path := t.PkgPath()
	if path != "" {
		inst.importCache[path] = t
	}
}

func (inst *Generator) writeTableDef() {
	inst.writeLine("type Table struct {")
	for _, col := range inst.tb.cs {
		inst.sb.WriteString(fmt.Sprintf("%s %s%c", col.name, col.typ.typ.String(), LF))
	}
	inst.sb.WriteString("}")
}

func (inst *Generator) writeImport() {
	inst.writeLine("import (")
	for path, _ := range inst.importCache {
		inst.writeLine(path)
	}
	inst.writeLine(")")
}

func (inst *Generator) writeLine(s string) {
	inst.sb.WriteString(s)
	inst.sb.WriteByte(LF)
}
