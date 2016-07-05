package xmp

import (
	"io"
	"errors"
)

type reader struct {
	extension string
	dump func (io.Reader) (string, error)
}

var readers []reader = []reader{}


// Register an XMP reader for a file extension
// The same reader can be registered multiple times for different extensions
func RegisterReader(extension string, dump func (io.Reader) (string, error)) {
	readers = append(readers, reader{extension, dump})
}

func FindReader(extension string) (*reader, bool) {
	for _, v := range readers {
		if v.extension == extension {
			return &v, true
		}
	}

	return nil, false
}

func DumpXMP(r io.Reader, extension string) (string, error) {
	reader, found := FindReader(extension)
	if !found {
		return "", errors.New("Format is not supported.")
	}

	xmpString, err := reader.dump(r)
	if err != nil {
		return "", err
	}

	return xmpString, nil
}
