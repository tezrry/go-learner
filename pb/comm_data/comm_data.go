package comm_data

type DataTableHeader struct {
	id    uint16
	typ   uint16
	lh    uint32
	dirty uint64
}

type storeType uint8

type FieldHeader struct {
	index  uint16
	dt     uint8
	store  uint8
	length uint32
}

type Field struct {
	FieldHeader
	data []byte
}

type MessageHeader struct {
	id      uint16
	nFields uint16
	length  uint32
}

type Message struct {
	MessageHeader
	fields []Field
}

type Comment struct {
}
