package formats

import (
	"github.com/mosen/psd"
	"io"
	"github.com/mosen/xmp"
)

func DumpXMPString(r io.Reader) (string, error) {
	psdFile, err := psd.Decode(r)
	if err != nil {
		return "", err
	}

	xmpString, err := psdFile.XMPString()
	if err != nil {
		return "", err
	}

	return xmpString, nil
}


func init() {
	xmp.RegisterReader("psd", DumpXMPString)
	xmp.RegisterReader("psb", DumpXMPString)
}
