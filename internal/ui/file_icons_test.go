package ui

import (
	"testing"

	"github.com/xddcode/jarlens/internal/archive"
)

func TestFileIconIDForUsesOpenIDEADefaults(t *testing.T) {
	tests := []struct {
		name string
		path string
		kind archive.NodeKind
		want fileIconID
	}{
		{name: "package", path: "example/service", kind: archive.NodePackage, want: fileIconPackage},
		{name: "class", path: "example/Service.class", kind: archive.NodeClass, want: fileIconClass},
		{name: "java", path: "Service.JAVA", kind: archive.NodeResource, want: fileIconJava},
		{name: "json", path: "application.json", kind: archive.NodeResource, want: fileIconJSON},
		{name: "yaml", path: "application.yaml", kind: archive.NodeResource, want: fileIconYAML},
		{name: "yml", path: "application.yml", kind: archive.NodeResource, want: fileIconYAML},
		{name: "ini", path: "application.ini", kind: archive.NodeResource, want: fileIconConfig},
		{name: "conf", path: "application.conf", kind: archive.NodeResource, want: fileIconConfig},
		{name: "cfg", path: "application.cfg", kind: archive.NodeResource, want: fileIconConfig},
		{name: "properties", path: "application.properties", kind: archive.NodeResource, want: fileIconProperties},
		{name: "xml", path: "application.xml", kind: archive.NodeResource, want: fileIconXML},
		{name: "html", path: "index.html", kind: archive.NodeResource, want: fileIconHTML},
		{name: "htm", path: "index.htm", kind: archive.NodeResource, want: fileIconHTML},
		{name: "xsd", path: "schema.xsd", kind: archive.NodeResource, want: fileIconXSD},
		{name: "text", path: "notes.txt", kind: archive.NodeResource, want: fileIconText},
		{name: "unknown", path: "archive/data.bin", kind: archive.NodeResource, want: fileIconText},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := fileIconIDFor(test.path, test.kind); got != test.want {
				t.Fatalf("fileIconIDFor(%q) = %v, want %v", test.path, got, test.want)
			}
		})
	}
}

func TestOpenIDEAFileIconsHaveLightAndDarkPalettes(t *testing.T) {
	for id := fileIconPackage; id <= fileIconText; id++ {
		light := openIDEAFileIcon(id, false)
		dark := openIDEAFileIcon(id, true)
		if light.ViewBox != 16 || dark.ViewBox != 16 {
			t.Fatalf("icon %v does not use the OpenIDEA 16-pixel view box", id)
		}
		if len(light.Layers) == 0 || len(dark.Layers) == 0 {
			t.Fatalf("icon %v is missing filled vector layers", id)
		}
		if light.Layers[0].Fill == dark.Layers[0].Fill {
			t.Errorf("icon %v uses the same primary color in both themes", id)
		}
	}
}

func TestOpenIDEAJavaIconPaintsBothStrokes(t *testing.T) {
	for _, dark := range []bool{false, true} {
		icon := openIDEAFileIcon(fileIconJava, dark)
		if icon.Stroke != 1 {
			t.Errorf("Java icon stroke = %v, want the SVG's 1-pixel stroke", icon.Stroke)
		}
		if len(icon.Commands) < 2 {
			t.Errorf("Java icon is missing its cup or handle stroke")
		}
	}
}
