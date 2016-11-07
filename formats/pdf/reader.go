package pdf

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/mosen/xmp"
	"github.com/mosen/xmp/formats/pdf/xref"
	"io"
	"strings"
	"strconv"
	"bytes"
)

// getCrossReferenceTableOffset locates the xref offset value in the trailer.
func getCrossReferenceTableOffset(r io.Reader) (int64, error) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		if xri := strings.Index(scanner.Text(), "startxref"); xri != -1 {
			scanner.Scan()
			xrefOffsetValue, err := strconv.ParseInt(scanner.Text(), 10, 0)
			if err != nil {
				return 0, fmt.Errorf("converting xref offset to integer: %s", err)
			}
			return xrefOffsetValue, nil
		}
	}

	return 0, errors.New("XREF table not found")
}

func DumpXMPString(r io.Reader) (string, error) {

	rs, ok := r.(io.ReadSeeker)
	if !ok {
		return "", errors.New("Expected a seekable reader")
	}

	xrefOffset, err := getCrossReferenceTableOffset(rs)
	if err != nil {
		return "", err
	}

	decoder := xref.NewDecoder(rs, xrefOffset)

	xrefTable := &xref.CrossReferenceTable{}
	if err := decoder.Decode(xrefTable); err != nil {
		return "", fmt.Errorf("decoding xref table: %s", err)
	}

	var buf []byte = make([]byte, 128)
	var xpacketObjectReference *xref.ObjectReference
	for _, ref := range xrefTable.References {
		rs.Seek(ref.Offset, 0)
		rs.Read(buf)

		if strings.Contains(string(buf), "<?xpacket begin") {
			xpacketObjectReference = &ref
			break
		}
	}

	if xpacketObjectReference == nil {
		return "", errors.New("no xpacket found")
	}

	rs.Seek(xpacketObjectReference.Offset, 0)
	var nextObjectReference *xref.ObjectReference = xrefTable.FindObjectId(xpacketObjectReference.Id + 1)
	if nextObjectReference == nil {
		return "", errors.New("unhandled case")
	}

	var xpacketLength int64 = nextObjectReference.Offset - xpacketObjectReference.Offset

	var xbuf []byte = make([]byte, xpacketLength)
	rs.Read(xbuf)

	var beginOffset, endOffset int
	beginOffset = bytes.Index(xbuf, []byte("<?xpacket begin"))
	endOffset = bytes.Index(xbuf, []byte("<?xpacket end"))

	var xmpPacket []byte = xbuf[beginOffset:endOffset]

	return string(xmpPacket), nil
}

func init() {
	xmp.RegisterReader("pdf", DumpXMPString)
}
