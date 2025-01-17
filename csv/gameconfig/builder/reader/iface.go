package reader

type IReader interface {
	Version(filename string) string
	Read(filename string) (rows [][]string, err error)
}
