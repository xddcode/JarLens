package ui

import (
	"sync"

	"github.com/vivaui/viva/app/iconpng"
	vivaui "github.com/vivaui/viva/ui"
	jarlensassets "github.com/xddcode/jarlens/assets"
)

var (
	appIconOnce sync.Once
	appIcon     vivaui.Bitmap
)

func applicationIcon() vivaui.Bitmap {
	appIconOnce.Do(func() {
		icon, err := iconpng.DecodeFit(jarlensassets.AppIconPNG, 512)
		if err != nil {
			return
		}
		pixels := make([]byte, len(icon.RGBA))
		for index := 0; index < len(pixels); index += 4 {
			red, green, blue, alpha := icon.RGBA[index], icon.RGBA[index+1], icon.RGBA[index+2], icon.RGBA[index+3]
			pixels[index] = uint8((uint16(blue)*uint16(alpha) + 127) / 255)
			pixels[index+1] = uint8((uint16(green)*uint16(alpha) + 127) / 255)
			pixels[index+2] = uint8((uint16(red)*uint16(alpha) + 127) / 255)
			pixels[index+3] = alpha
		}
		appIcon = vivaui.Bitmap{
			Pix: pixels, Width: icon.Size, Height: icon.Size, Stride: icon.Size * 4,
			ID: 0x4a4152, Version: 1,
		}
	})
	return appIcon
}
