package metafile

import (
	"testing"

	"csv/gameconfig/infra/ctype"
	"go-learner/slice"

	"github.com/stretchr/testify/require"
)

func TestCreateFile(t *testing.T) {
	b := make([]byte, VersionLen)
	b[0] = 'a'
	b[VersionLen-1] = 'a'
	mf, err := CreateFile("test", slice.ByteSlice2String(b), "shop", 111)
	require.NoError(t, err)
	require.NotNil(t, mf)

	id1 := "abc"
	gid1 := ctype.GlobalId(mf.TableId(), 1)
	err = mf.AddId(id1)
	require.NoError(t, err)
	require.Equal(t, gid1, mf.NameToGlobalID(id1))
	err = mf.AddId(id1)
	require.NoError(t, err)
	require.Equal(t, gid1, mf.NameToGlobalID(id1))

	id2 := "def"
	gid2 := ctype.GlobalId(mf.TableId(), 2)
	err = mf.AddId(id2)
	require.NoError(t, err)
	require.Equal(t, gid2, mf.NameToGlobalID(id2))

	mf.Close()
}

func TestLoadFile(t *testing.T) {
	mf, err := LoadFile("test", "shop")
	require.NoError(t, err)
	require.NotNil(t, mf)
	t.Log(mf.Version())
	t.Log(mf.TableId())

	err = mf.LoadAll()
	require.NoError(t, err)

	require.Equal(t, ctype.GlobalId(mf.TableId(), 2), mf.NameToGlobalID("def"))
}

func TestSetVersion(t *testing.T) {
	mf, err := LoadFile("test", "shop")
	require.NoError(t, err)
	require.NotNil(t, mf)

	b := make([]byte, VersionLen)
	b[0] = 'b'
	b[VersionLen-1] = 'b'
	err = mf.Update(slice.ByteSlice2String(b))
	require.NoError(t, err)

	err = mf.AddId("fgh")
	require.NoError(t, err)
	require.Equal(t, ctype.GlobalId(mf.TableId(), 3), mf.NameToGlobalID("fgh"))
}
