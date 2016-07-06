package png

import (
	"testing"
	"io/ioutil"
	"bytes"
	"fmt"
)

func TestDumpXMPString(t *testing.T) {
	data, err := ioutil.ReadFile("../../bluesquare/BlueSquare.png")
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