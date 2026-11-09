package ui

import (
	"errors"
	"sync"

	"github.com/vivaui/viva/app"
	vivaui "github.com/vivaui/viva/ui"
	"github.com/xddcode/jarlens/assets"
)

var (
	fontOnce        sync.Once
	monoFace        vivaui.MonoFace
	interfaceFamily string
	jetBrainsFamily string
	fontErr         error
)

func configureFonts() {
	fontOnce.Do(func() {
		inter, err := assets.Files.ReadFile("fonts/InterVariable.ttf")
		if err != nil {
			fontErr = err
		} else {
			family, _, err := app.RegisterFont(inter)
			if err != nil {
				fontErr = err
			} else if !app.SetInterfaceFontFamily(family) {
				fontErr = errors.New("the interface font is unavailable")
			} else {
				interfaceFamily = family
			}
		}

		mono, err := assets.Files.ReadFile("fonts/JetBrainsMono-Variable.ttf")
		if err != nil {
			fontErr = errors.Join(fontErr, err)
			monoFace = vivaui.MonoSystem
			return
		}
		family, _, err := app.RegisterFont(mono)
		if err != nil {
			fontErr = errors.Join(fontErr, err)
			monoFace = vivaui.MonoSystem
			return
		}
		jetBrainsFamily = family
		monoFace = vivaui.MonoFamily(family)
	})
}
