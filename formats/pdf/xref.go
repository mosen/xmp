package pdf

import "io"

type ObjectReference struct {
	Offset int
	Generation int
}

type CrossReferenceTable struct {
	// Object number to offset reference
	References map[int]ObjectReference
}

func NewCrossReferenceTable() *CrossReferenceTable {
	return &CrossReferenceTable{
		References: map[int]ObjectReference{},
	}
}


func DecodeCrossReferenceTable (r io.Reader) (*CrossReferenceTable, error) {
    return &CrossReferenceTable{}, nil
}
