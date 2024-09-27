package plugin

import (
	"testing"

	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/stretchr/testify/require"
)

type testGenerator struct{}

func (inst *testGenerator) Generate(p *Plugin, file *generator.FileDescriptor) {
	p.P("// Code generated by protoc-gen-go, DO NOT EDIT.")
}

func TestDryRun(t *testing.T) {
	p := NewPlugin("test", &testGenerator{})
	err := p.DryRun("../common.proto")
	require.NoError(t, err)
}