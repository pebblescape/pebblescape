// Package table allows creating properly indented tables for CLIs.
package table

import (
	"bytes"
	"strconv"
)

// Align specifies table cell alignment.
type Align int

// Table cell alignments
const (
	Left Align = iota
	Right
)

// CellConf specifices table cell configuration.
type CellConf struct {
	Align    Align
	PadRight []byte
	PadLeft  []byte
}

// Table holds all the data needed to build a table.
type Table struct {
	CellConf []CellConf
	Data     [][]string
}

// Add adds a row of cells to the table.
func (t *Table) Add(cells ...string) *Table {
	if len(cells) != len(t.CellConf) {
		panic("expected " + strconv.Itoa(len(t.CellConf)) + " got " + strconv.Itoa(len(cells)))
	}
	t.Data = append(t.Data, cells)
	return t
}

// New creates a new table of specified columns.
func New(size int) *Table {
	conf := make([]CellConf, size)
	dflt := CellConf{Left, []byte{' '}, []byte{}}
	for i := 0; i < size; i++ {
		conf[i] = dflt
	}
	conf[size-1].PadRight = []byte{}
	return NewWithConf(conf)
}

// NewWithConf creates a new Table with specified cell configuration.
func NewWithConf(conf []CellConf) *Table {
	return &Table{conf, nil}
}

// String prints current table to string.
func (t *Table) String() string {
	if len(t.Data) == 0 {
		return "\n"
	}
	b := new(bytes.Buffer)
	max := func(p *int, v int) {
		if *p < v {
			*p = v
		}
	}
	lengths := make([]int, len(t.Data[0]))
	for _, v := range t.Data {
		for i, cell := range v {
			max(&lengths[i], len(cell))
		}
	}
	spc := []byte{' '}
	for _, v := range t.Data {
		for i, cell := range v {
			b.Write(t.CellConf[i].PadLeft)
			if t.CellConf[i].Align == Right {
				b.Write(bytes.Repeat(spc, lengths[i]-len(cell)))
			}
			b.WriteString(cell)
			if t.CellConf[i].Align == Left && i < len(v)-1 {
				b.Write(bytes.Repeat(spc, lengths[i]-len(cell)))
			}
			b.Write(t.CellConf[i].PadRight)
		}
		b.WriteString("\n")
	}
	return b.String()
}
