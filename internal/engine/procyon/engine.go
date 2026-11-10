package procyon

import (
	_ "embed"

	"github.com/xddcode/jarlens/internal/decompile"
	"github.com/xddcode/jarlens/internal/engine/javaengine"
)

const (
	engineVersion = "0.6.0"
	artifactSHA   = "821da96012fc69244fa1ea298c90455ee4e021434bc796d3b9546ab24601b779"
)

//go:embed assets/procyon-decompiler-0.6.0.jar
var artifact []byte

func New() *javaengine.Engine {
	return javaengine.New(javaengine.Config{
		Name: "Procyon", Version: engineVersion, ArtifactName: "procyon-decompiler-" + engineVersion + ".jar",
		SHA256: artifactSHA, Artifact: artifact,
		Arguments: func(_ decompile.Request, input, output string) []string {
			return []string{"--output-directory", output, "--unicode", input}
		},
	})
}
