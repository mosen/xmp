package indd

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/mosen/xmp"
	"io"
	"io/ioutil"
	"os"
)

const (
	MASTER_PAGE_LENGTH = 4096
	PAGE_LENGTH        = 4096

	MARKER_LENGTH = 32

	MINIMUM_XMP_LENGTH = 4 + 53 + 19 // x + header + trailer len
	XPACKET_BEGIN      = "<?xpacket begin="
)

var HEADER_MARKER_GUID [16]byte = [16]byte{0xde, 0x39, 0x39, 0x79, 0x51, 0x88, 0x4b, 0x6c, 0x8e, 0x63, 0xee, 0xf8, 0xae, 0xe0, 0xdd, 0x38}
var MASTER_PAGE_GUID [16]byte = [16]byte{0x06, 0x06, 0xed, 0xf5, 0xd8, 0x1d, 0x46, 0xe5, 0xbd, 0x31, 0xef, 0xe7, 0xfe, 0x74, 0xb7, 0x1d}
var TRAILER_MARKER_ZGUID [16]byte = [16]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
var TRAILER_MARKER_GUID [16]byte = [16]byte{0xfd, 0xce, 0xdb, 0x70, 0xf7, 0x86, 0x4b, 0x4f, 0xa4, 0xd3, 0xc7, 0x28, 0xb3, 0x41, 0x71, 0x06}

type MasterPage struct {
	Guid               [16]byte
	Magic              [8]byte
	ObjectStreamEndian uint8
	_                  [239]byte
	SequenceNumber     uint64 // Little endian
	_                  [8]byte
	FilePages          uint32 // Little endian
	_                  [3812]byte
}

func DecodeMasterPage(r io.Reader) (*MasterPage, error) {
	master := MasterPage{}

	if err := binary.Read(r, binary.LittleEndian, &master); err != nil {
		return nil, err
	}

	if master.Guid != MASTER_PAGE_GUID {
		return nil, errors.New("master page does not match GUID. file is not a valid indesign file")
	}

	fmt.Printf("Master page: Guid(%x), SequenceNumber(%d), FilePages(%d)\n", master.Guid, master.SequenceNumber, master.FilePages)

	return &master, nil
}

type ContiguousObjectMarker struct {
	Guid          [16]byte
	ObjectUID     uint32
	ObjectClassID uint32
	StreamLength  uint32
	Checksum      uint32

	IsHeader  bool
	IsTrailer bool
}

func DecodeContiguousObjectMarker(r io.Reader) (*ContiguousObjectMarker, error) {

	var guid [16]byte
	if err := binary.Read(r, binary.LittleEndian, &guid); err != nil {
		return nil, err
	}

	var objectUid uint32
	if err := binary.Read(r, binary.LittleEndian, &objectUid); err != nil {
		return nil, err
	}

	var objectClassId uint32
	if err := binary.Read(r, binary.LittleEndian, &objectClassId); err != nil {
		return nil, err
	}

	var streamLength uint32
	if err := binary.Read(r, binary.LittleEndian, &streamLength); err != nil {
		return nil, err
	}

	var checksum uint32
	if err := binary.Read(r, binary.LittleEndian, &checksum); err != nil {
		return nil, err
	}

	var marker *ContiguousObjectMarker = &ContiguousObjectMarker{
		guid,
		objectUid,
		objectClassId,
		streamLength,
		checksum,
		false,
		false,
	}

	marker.IsHeader = marker.Guid == HEADER_MARKER_GUID
	marker.IsTrailer = marker.Guid == TRAILER_MARKER_GUID

	fmt.Printf("Contiguous Object Marker: Guid(%x), StreamLength(%d), Checksum(%x)\n", marker.Guid, marker.StreamLength, marker.Checksum)

	return marker, nil
}

type InDesignDocument struct {
	masters []MasterPage
	pages   uint32
	endian  uint8
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
	contObjOffset := pages*PAGE_LENGTH - (2 * MARKER_LENGTH)

	fmt.Printf("Seeking to contiguous object section at offset: %d\n", contObjOffset)

	if s, ok := r.(io.Seeker); ok {
		s.Seek(int64(contObjOffset), os.SEEK_SET)
	} else {
		io.CopyN(ioutil.Discard, r, int64(contObjOffset-(2*MASTER_PAGE_LENGTH)))
	}

	indd := &InDesignDocument{
		[]MasterPage{*firstMaster, *secondMaster},
		pages,
		endian,
		[][]byte{},
	}

	for {
		io.CopyN(ioutil.Discard, r, 2*MARKER_LENGTH)

		marker, err := DecodeContiguousObjectMarker(r)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		var innerLength uint32
		if endian == 1 {
			if err := binary.Read(r, binary.LittleEndian, &innerLength); err != nil {
				return nil, err
			}
		} else {
			if err := binary.Read(r, binary.BigEndian, &innerLength); err != nil {
				return nil, err
			}
		}

		fmt.Println("Inner length: ", innerLength)
		if innerLength < MINIMUM_XMP_LENGTH {
			io.CopyN(ioutil.Discard, r, int64(innerLength))
			continue
		}


		if innerLength != marker.StreamLength-4 {
			return nil, errors.New("contiguous object marker inner length does not match its stream length")
		}

		var stream []byte = make([]byte, innerLength)
		if err := binary.Read(r, binary.LittleEndian, &stream); err != nil {
			return nil, errors.New("Could not read stream of contiguous object")
		}

		indd.streams = append(indd.streams, stream)

	}

	return indd, nil
}

func DumpXMPString(r io.Reader) (string, error) {
	indd, err := DecodeInDesignDocument(r)

	if err != nil {
		return "", err
	} else {
		// TODO: Loop through streams to check that they have xpacket begin
		xmpStream := indd.streams[0]
		return string(xmpStream), nil
	}
}

func init() {
	xmp.RegisterReader("indd", DumpXMPString)
}
