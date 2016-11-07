package xref

import (
	"io"
	"fmt"
	"bufio"
	"strings"
	"strconv"
)

const XREF_START string = "xref"


type Decoder struct {
	r io.ReadSeeker
	offset int64
}

func (dec *Decoder) Decode(v *CrossReferenceTable) error {
	dec.r.Seek(dec.offset, 0) // Seek to XREF table offset
	scanner := bufio.NewScanner(dec.r)
	scanner.Scan()

	header := scanner.Text()

	if !strings.HasPrefix(header, XREF_START) {
		return fmt.Errorf("start of cross reference table did not match expected value: %s", header)
	}

	scanner.Scan()
	objectcounts := strings.Split(scanner.Text(), " ")
	if len(objectcounts) != 2 {
		return fmt.Errorf("expected object start, object count in xref table, got: %d object(s)", len(objectcounts))
	}

	objectStart, err := strconv.Atoi(objectcounts[0])
	if err != nil {
		return fmt.Errorf("converting object start to integer: %s", err)
	}

	objectCount, err := strconv.Atoi(objectcounts[1])
	if err != nil {
		return fmt.Errorf("converting object count to integer: %s", err)
	}

	v.ObjectStart = objectStart
	v.ObjectCount = objectCount
	v.References = make([]ObjectReference, v.ObjectCount)

	for i := 0; i < v.ObjectCount; i++ {
		scanner.Scan()
		reference := strings.Split(scanner.Text(), " ")
		objectOffset, err := strconv.Atoi(reference[0])
		if err != nil {
			return fmt.Errorf("converting object reference offset to int: %s", err)
		}

		objectGeneration, err := strconv.Atoi(reference[1])
		if err != nil {
			return fmt.Errorf("converting object reference generation to int: %s", err)
		}

		v.References[i] = ObjectReference{
			Offset: objectOffset,
			Generation: objectGeneration,
		}
	}

	return nil
}

func NewDecoder(r io.ReadSeeker, offset int64) *Decoder {
	return &Decoder{r: r, offset: offset}
}

