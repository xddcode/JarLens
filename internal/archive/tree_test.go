package archive

import (
	"testing"
)

func TestTreeIsSortedAndLazy(t *testing.T) {
	tree := BuildTree([]Entry{
		{ID: 0, Path: "z/Last.class", Name: "Last.class", Kind: EntryClass},
		{ID: 1, Path: "com/example/Hello.class", Name: "Hello.class", Kind: EntryClass},
		{ID: 2, Path: "com/example/World.class", Name: "World.class", Kind: EntryClass},
		{ID: 3, Path: "META-INF/MANIFEST.MF", Name: "MANIFEST.MF", Kind: EntryResource},
	})
	visible := tree.Visible(nil)
	if len(visible) != 3 {
		t.Fatalf("initial visible nodes: %+v", visible)
	}
	if visible[0].Node.Name != "com" || visible[1].Node.Name != "META-INF" || visible[2].Node.Name != "z" {
		t.Fatalf("roots are not sorted: %+v", visible)
	}

	com := visible[0].Node
	visible = tree.Visible(map[int]bool{com.ID: true})
	if len(visible) != 4 || visible[1].Node.Name != "example" || visible[1].Depth != 1 {
		t.Fatalf("expanded package is wrong: %+v", visible)
	}
	example := visible[1].Node
	visible = tree.Visible(map[int]bool{com.ID: true, example.ID: true})
	if len(visible) != 6 || visible[2].Node.Name != "Hello" || visible[3].Node.Name != "World" {
		t.Fatalf("expanded classes are wrong: %+v", visible)
	}
	if visible[2].Node.Entry != 1 || visible[3].Node.Entry != 2 {
		t.Fatal("tree nodes lost their archive entry identities")
	}
}

func TestTreeSupportsDuplicateEntryNames(t *testing.T) {
	tree := BuildTree([]Entry{
		{ID: 4, Path: "a/Thing.class", Name: "Thing.class", Kind: EntryClass},
		{ID: 7, Path: "a/Thing.class", Name: "Thing.class", Kind: EntryClass},
	})
	expanded := map[int]bool{}
	for _, node := range tree.Nodes {
		if node.Kind == NodePackage {
			expanded[node.ID] = true
		}
	}
	visible := tree.Visible(expanded)
	if len(visible) != 3 || visible[1].Node.Entry != 4 || visible[2].Node.Entry != 7 {
		t.Fatalf("duplicate entries lost identity: %+v", visible)
	}
}

func FuzzBuildTree(f *testing.F) {
	f.Add("a/b/Hello.class", "Hello.class")
	f.Add("../escape.class", "escape.class")
	f.Add("", "")
	f.Fuzz(func(t *testing.T, path, name string) {
		entry := Entry{ID: 0, Path: path, Name: name, Kind: EntryClass}
		tree := BuildTree([]Entry{entry})
		if path != "" {
			tree.Visible(map[int]bool{0: true})
		}
	})
}
