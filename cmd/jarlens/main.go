package main

import (
	"log"
	"os"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/ui"
	jarlensui "github.com/xddcode/jarlens/internal/ui"
)

func main() {
	model := &jarlensui.Model{}
	config := jarlensui.NewConfig(model)
	var canvas ui.Canvas
	err := app.RunApplication(func(application *app.Application) error {
		if _, err := application.OpenWindow(config, &canvas); err != nil {
			return err
		}
		if len(os.Args) > 1 {
			model.OpenPath(os.Args[1])
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
