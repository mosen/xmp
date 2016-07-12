package pdf

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/mosen/xmp"
	"io"
	"strconv"
)

func getCrossReferenceTableOffset(r io.Reader) (int, error) {

	xrefScanner := bufio.NewScanner(r)
	// Split func "borrowed" from golang
	xrefScanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {

		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, 0x0d); i >= 0 {
			// We have a full newline-terminated line.
			return i + 1, data[0:i], nil
		}
		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil

	})

	for xrefScanner.Scan() {
		if xrefScanner.Text() == "startxref" {
			xrefScanner.Scan()
			xrefOffset, err := strconv.Atoi(xrefScanner.Text())
			if err != nil {
				return 0, err
			}

			return xrefOffset, nil
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
