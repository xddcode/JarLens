package ui

import (
	"testing"

	"github.com/vivaui/viva/app/iconpng"
	jarlensassets "github.com/xddcode/jarlens/assets"
)

func TestApplicationIconConvertsRGBAToPremultipliedBGRA(t *testing.T) {
	source, err := iconpng.DecodeFit(jarlensassets.AppIconPNG, 512)
	if err != nil {
		t.Fatal(err)
	}
	bitmap := applicationIcon()
	for index := 0; index < len(source.RGBA); index += 4 {
		red, green, blue, alpha := source.RGBA[index], source.RGBA[index+1], source.RGBA[index+2], source.RGBA[index+3]
		if alpha == 0 || red == blue {
			continue
		}
		wantBlue := uint8((uint16(blue)*uint16(alpha) + 127) / 255)
		wantGreen := uint8((uint16(green)*uint16(alpha) + 127) / 255)
		wantRed := uint8((uint16(red)*uint16(alpha) + 127) / 255)
		if bitmap.Pix[index] != wantBlue || bitmap.Pix[index+1] != wantGreen || bitmap.Pix[index+2] != wantRed || bitmap.Pix[index+3] != alpha {
			t.Fatalf("icon pixel %d has the wrong channel order or alpha", index/4)
		}
		return
	}
	t.Fatal("icon has no opaque sample with distinct red and blue channels")
}
