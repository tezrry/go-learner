package generator

import "testing"

func TestPkgName(t *testing.T) {
	g := NewGenerator(nil)
	g.writeImport()
	t.Log(g.sb.String())
}
