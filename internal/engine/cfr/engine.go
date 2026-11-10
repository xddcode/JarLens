package cfr

import (
	_ "embed"
	"regexp"
	"strings"

	"github.com/xddcode/jarlens/internal/decompile"
	"github.com/xddcode/jarlens/internal/engine/javaengine"
)

const (
	engineVersion = "0.152"
	artifactSHA   = "f686e8f3ded377d7bc87d216a90e9e9512df4156e75b06c655a16648ae8765b2"
)

//go:embed assets/cfr-0.152.jar
var artifact []byte

func New() *javaengine.Engine {
	return javaengine.New(javaengine.Config{
		Name: "CFR", Version: engineVersion, ArtifactName: "cfr-" + engineVersion + ".jar",
		SHA256: artifactSHA, Artifact: artifact, UseArchive: true,
		Arguments: func(request decompile.Request, input, output string) []string {
			arguments := []string{input, "--outputdir", output, "--silent", "true", "--comments", "false"}
			if input == request.ArchivePath {
				className := strings.TrimSuffix(strings.ReplaceAll(request.ClassName, "/", "."), ".class")
				arguments = append(arguments, "--jarfilter", "^"+regexp.QuoteMeta(className)+"$")
			}
			return arguments
		},
	})
}
