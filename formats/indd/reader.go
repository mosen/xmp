package indd

import (
	"io"
	"github.com/mosen/xmp"
	"encoding/binary"
	"errors"
	"os"
	"io/ioutil"
	"fmt"
)

const (
	MASTER_PAGE_LENGTH = 4096
	PAGE_LENGTH = 4096

	MARKER_LENGTH = 32

	MINIMUM_XMP_LENGTH = 4 + 53 + 19 // x + header + trailer len
	XPACKET_BEGIN = "<?xpacket begin="
)

var HEADER_MARKER_GUID [16]byte = [16]byte{0xde, 0x39, 0x39, 0x79, 0x51, 0x88, 0x4b, 0x6c, 0x8e, 0x63, 0xee, 0xf8, 0xae, 0xe0, 0xdd, 0x38}
var MASTER_PAGE_GUID [16]byte = [16]byte{0x06, 0x06, 0xed, 0xf5, 0xd8, 0x1d, 0x46, 0xe5, 0xbd, 0x31, 0xef, 0xe7, 0xfe, 0x74, 0xb7, 0x1d}
var TRAILER_MARKER_GUID [16]byte = [16]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}


type MasterPage struct {
	Guid [16]byte
	Magic [8]byte
	ObjectStreamEndian uint8
	_ [239]byte
	SequenceNumber uint64 // Little endian
	_ [8]byte
	FilePages uint32 // Little endian
	_ [3812]byte
}

func DecodeMasterPage(r io.Reader) (*MasterPage, error) {
	master := MasterPage{}

  	if err := binary.Read(r, binary.LittleEndian, &master); err != nil {
		return nil, err
	}

	if master.Guid != MASTER_PAGE_GUID {
		return nil, errors.New("master page does not match GUID. file is not a valid indesign file")
	}

	return &master, nil
}

type ContiguousObjectMarker struct {
 	Guid [16]byte
	ObjectUID uint32
	ObjectClassID uint32
	StreamLength uint32
	Checksum uint32

	IsHeader bool
	IsTrailer bool
}

func DecodeContiguousObjectMarker(r io.Reader) (*ContiguousObjectMarker, error) {
	marker := ContiguousObjectMarker{}

	if err := binary.Read(r, binary.LittleEndian, &marker); err != nil {
		return nil, err
	}

	marker.IsHeader = marker.Guid == HEADER_MARKER_GUID
	marker.IsTrailer = marker.Guid == TRAILER_MARKER_GUID

	return &marker, nil
}

type InDesignDocument struct {
	masters []MasterPage
	pages uint32
	endian uint8
	streams [][]byte
}

func DecodeInDesignDocument(r io.Reader) (*InDesignDocument, error) {
	firstMaster, err := DecodeMasterPage(r)
	if err != nil {
		return nil, err
	}

	secondMaster, err := DecodeMasterPage(r)
	if err != nil {
		return nil, err
	}

	var pages uint32
	var endian uint8

	if secondMaster.SequenceNumber > firstMaster.SequenceNumber {
		pages = secondMaster.FilePages
		endian = secondMaster.ObjectStreamEndian
	} else {
		pages = firstMaster.FilePages
		endian = firstMaster.ObjectStreamEndian
	}

	// Skip past the data pages indicated by FilePages to get to the contiguous object section.
	contObjOffset := pages * PAGE_LENGTH - (2 * MARKER_LENGTH)

	if s, ok := r.(io.Seeker); ok {
		s.Seek(int64(contObjOffset), os.SEEK_SET)
	} else {
		io.CopyN(ioutil.Discard, r, int64(contObjOffset - (2 * MASTER_PAGE_LENGTH)))
	}

	var streams [][]byte = make([][]byte, 20)

	for {
		io.CopyN(ioutil.Discard, r, 2 * MARKER_LENGTH)

		marker, err := DecodeContiguousObjectMarker(r)
		if err != nil {
			break
		}

		var innerLength uint32
		if endian == 1 {
			if err :=binary.Read(r, binary.LittleEndian, &innerLength); err != nil {
				return nil, err
			}
		} else {
			if err :=binary.Read(r, binary.BigEndian, &innerLength); err != nil {
				return nil, err
			}
		}

		fmt.Println("Inner length: ", innerLength)

		if innerLength != marker.StreamLength - 4 {
			return nil, errors.New("continuous object marker inner length does not match its stream length")
		}

		var stream []byte = make([]byte, innerLength)
		if err := binary.Read(r, binary.LittleEndian, &stream); err != nil {
			return nil, errors.New("Could not read stream of continuous object")
		}

		streams = append(streams, stream)

	}

	return &InDesignDocument{
		[]MasterPage{*firstMaster, *secondMaster},
		pages,
		endian,
		streams,
	}, nil
}

func DumpXMPString(r io.Reader) (string, error) {
	indd, err := DecodeInDesignDocument(r)

	fmt.Printf("%v\n", indd)

	if err != nil {
		return "", err
	} else {
		return "", nil
	}
}

func init() {
	xmp.RegisterReader("indd", DumpXMPString)
}
