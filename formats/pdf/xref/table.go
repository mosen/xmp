package xref

import "fmt"

type CrossReferenceObjectFreeStatus []byte

var (
	CrossReferenceObjectInUse CrossReferenceObjectFreeStatus = []byte{'n'}
	CrossReferenceObjectFree                                 = []byte{'f'}
)

type ObjectReference struct {
	Offset     int64 // int64 more convenient for io.ReadSeeker
	Generation int
	InUse      CrossReferenceObjectFreeStatus
	Id	   int
}

type CrossReferenceTable struct {
	// Object number to offset reference
	ObjectStart int
	ObjectCount int

	References []ObjectReference
}

func (cr *CrossReferenceTable) FindObjectId(objectId int) *ObjectReference {
	for _, ref := range cr.References {
		if ref.Id == objectId {
			return &ref
		}
	}

	return nil
}

func (cr *CrossReferenceTable) String() string {
	return fmt.Sprintf("CrossReferenceTable with %d References, ID range %d - %d", len(cr.References),
		cr.ObjectStart, cr.ObjectStart+cr.ObjectCount)
}
