package pdf

import (
	"io"
	"github.com/mosen/xmp"
)

func DumpXMPString(r io.Reader) (string, error) {

}

func init() {
	xmp.RegisterReader("pdf", DumpXMPString)
}

