package pdf

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestDumpXMPString(t *testing.T) {
	data, err := ioutil.ReadFile("../../bluesquare/BlueSquare.pdf")
	reader := bytes.NewReader(data)

	if err != nil {
		t.Error(err)
	}

	xmp, err := DumpXMPString(reader)

	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%v\n", xmp)
}
