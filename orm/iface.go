package orm

type IDriver interface {
	Get(rt *RecordType, keys ...any) [][]byte
}
