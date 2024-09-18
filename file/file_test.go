package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDir(t *testing.T) {
	dirname, err := filepath.Abs("..")
	require.NoError(t, err)

	path, err := os.Open(dirname)
	defer func() {
		_ = path.Close()
	}()

	require.NoError(t, err)
	t.Log(path.Name())
	t.Log(filepath.Dir(path.Name()))
	t.Log(filepath.Base(path.Name()))

	files, err := path.Readdir(-1)
	require.NoError(t, err)
	for _, file := range files {
		t.Log(file.Name())

		t.Log(filepath.Ext(file.Name()))
		t.Log(strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())))

	}
}
