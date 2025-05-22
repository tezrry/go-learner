package log

import (
	"testing"
)

var m map[string]interface{}

func TestField(t *testing.T) {
	m = map[string]interface{}{
		"a": 1,
		"b": "string",
	}
	Event("test", 1, "string", int64(-2), m)

}
