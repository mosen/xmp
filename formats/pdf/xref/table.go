package xref

import "fmt"

type CrossReferenceObjectFreeStatus []byte

var (
	CrossReferenceObjectInUse CrossReferenceObjectFreeStatus = []byte{'n'}
	CrossReferenceObjectFree                                 = []byte{'f'}
)

type ObjectReference struct {
	Offset     int
	Generation int
	InUse      CrossReferenceObjectFreeStatus
}

type CrossReferenceTable struct {
	// Object number to offset reference
	ObjectStart int
	ObjectCount int

	References []ObjectReference
}

func (cr *CrossReferenceTable) String() string {
	return fmt.Sprintf("CrossReferenceTable with %d References, ID range %d - %d", len(cr.References),
		cr.ObjectStart, cr.ObjectStart+cr.ObjectCount)
}
