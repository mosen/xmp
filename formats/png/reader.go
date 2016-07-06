package png

import (
	"io"
	"github.com/mosen/xmp"
	"encoding/binary"
	"strings"
	"fmt"
)

const ITXT_HEADER_LENGTH = 22
//const ITXT_HEADER_DATA_XMP = "XML:com.adobe.xmp\0\0\0\0\0"

const ITXT_HEADER_DATA_XMP = "XML:com.adobe.xmp"


var PNG_MAGIC = []byte{137, 80, 78, 71, 13, 10, 26, 10}

var CHUNK_TYPES = []string{
	"IHDR",
	"PLTE",
	"IDAT",
	"IEND",
	"cHRM",
	"gAMA",
	"iCCP",
	"sBIT",
	"sRGB",
	"bKGD",
	"hIST",
	"tRNS",
	"pHYs",
	"sPLT",
	"tIME",
	"iTXt",
	"tEXt",
	"zTXt",
};

type PNG struct {
	chunks []Chunk
	xmpString string
}

type Chunk struct {
	length uint32
	chunkType []byte
	data []byte
	crc uint32
}

func DecodeChunk(r io.Reader) (*Chunk, error) {
	var length uint32
	binary.Read(r, binary.LittleEndian, &length)

	var chunkType []byte = make([]byte, 4)
	binary.Read(r, binary.LittleEndian, chunkType)

	var chunkData []byte
	if length > 0 {
		chunkData = make([]byte, length)
		binary.Read(r, binary.LittleEndian, chunkData)
	}

	var crc uint32
	binary.Read(r, binary.LittleEndian, &crc)

	return &Chunk{
		length,
		chunkType,
		chunkData,
		crc,
	}, nil
}

func IsXMPChunk(chunk *Chunk) bool {
	chunkType := string(chunk.chunkType)
	fmt.Println(chunkType)

	if chunkType != "iTXt" {
		return false
	}

	if strings.Compare(string(chunk.data[0:ITXT_HEADER_LENGTH]), ITXT_HEADER_DATA_XMP) != 0 {
	 	return false
	}

	return true
}

func IsEndChunk(chunk *Chunk) bool {
	chunkType := string(chunk.chunkType)
	fmt.Println(chunkType)

	return chunkType == "IEND"
}

func DecodePNG(r io.Reader) (*PNG, error) {

	// Magic header
	var magic []byte = make([]byte, 8)
	_, err := r.Read(magic)

	if err != nil {
		return nil, err
	}

	fmt.Println("Magic %x\n", magic)

	//if magic != PNG_MAGIC {
	//	return nil, errors.New("this file does not seem to be a png file")
	//}

	PNG := &PNG{xmpString:""}

	fmt.Println("Decoding chunks")

	for {
		fmt.Println("Next chunk")
		chunk, err := DecodeChunk(r)
		if err != nil {
			return nil, err
		}

		fmt.Printf("Length: %d bytes\n", chunk.length)
		fmt.Printf("Type: %s\n", chunk.chunkType)

		if IsEndChunk(chunk) {
			break
		}

		PNG.chunks = append(PNG.chunks, *chunk)

		if IsXMPChunk(chunk) {
			PNG.xmpString = string(chunk.data)
			fmt.Println(string(chunk.data))
			break
		}
	}

	return PNG, nil
}

// Dump xmp packet from iTXt chunk
func DumpXMPString(r io.Reader) (string, error) {
	_, err := DecodePNG(r)

	if err != nil {
		return "", err
	}

	return "", nil
}


func init() {
	xmp.RegisterReader("png", DumpXMPString)
}

