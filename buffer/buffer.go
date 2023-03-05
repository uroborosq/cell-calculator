package buffer

type CellBuffer struct {
	buffer      map[string]map[string]string
	threshold   uint64
	currentSize uint64
}

func NewCellBuffer(threshold uint64) *CellBuffer {
	return &CellBuffer{
		buffer:    make(map[string]map[string]string),
		threshold: threshold,
	}
}

func (b *CellBuffer) Add(column string, row string, value string) {
	if b.currentSize >= b.threshold {
		for column := range b.buffer {
			for row := range b.buffer {
				delete(b.buffer[column], row)
			}
		}
		return
	}
	if b.buffer[column] == nil {
		b.buffer[column] = make(map[string]string, 1)
	}
	b.buffer[column][row] = value
	b.currentSize++
}

func (b *CellBuffer) Get(column string, row string) (string, bool) {
	if b.buffer[column] == nil {
		return "", false
	}
	value, ok := b.buffer[column][row]
	return value, ok
}

func (b *CellBuffer) Update(column string, row string, value string) bool {
	if _, ok := b.buffer[column][row]; ok {
		b.buffer[column][row] = value
		return true
	}
	return false
}
