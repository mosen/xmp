package pdf

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/mosen/xmp"
	"io"
	"strings"
	"strconv"
)

// getCrossReferenceTableOffset locates the xref offset value in the trailer.
func getCrossReferenceTableOffset(r io.Reader) (int, error) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		if xri := strings.Index(scanner.Text(), "startxref"); xri != -1 {
			scanner.Scan()
			xrefOffsetValue, err := strconv.Atoi(scanner.Text())
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

	xref, err := getCrossReferenceTableOffset(rs)
	if err != nil {
		return "", err
	}

	fmt.Printf("XREF table offset: %d", xref)

	return "", nil
}

func init() {
	xmp.RegisterReader("pdf", DumpXMPString)
}
