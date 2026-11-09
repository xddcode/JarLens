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
	if len(visible) != 6 || visible[2].Node.Name != "Hello.class" || visible[3].Node.Name != "World.class" {
		t.Fatalf("expanded classes are wrong: %+v", visible)
	}
	if visible[2].Node.Entry != 1 || visible[3].Node.Entry != 2 {
		t.Fatal("tree nodes lost their archive entry identities")
	}
}

func TestTreeGroupsNestedClasses(t *testing.T) {
	tree := BuildTree([]Entry{
		{ID: 0, Path: "example/Outer.class", Name: "Outer.class", Kind: EntryClass},
		{ID: 1, Path: "example/Outer$Inner.class", Name: "Outer$Inner.class", Kind: EntryClass},
		{ID: 2, Path: "example/Outer$Inner$Deep.class", Name: "Outer$Inner$Deep.class", Kind: EntryClass},
		{ID: 3, Path: "example/Orphan$Inner.class", Name: "Orphan$Inner.class", Kind: EntryClass},
		{ID: 4, Path: "example/Outer$1.class", Name: "Outer$1.class", Kind: EntryClass},
		{ID: 5, Path: "example/Outer$1Local.class", Name: "Outer$1Local.class", Kind: EntryClass},
	})
	byPath := make(map[string]Node)
	for _, node := range tree.Nodes {
		byPath[node.Path] = node
	}
	outer := byPath["example/Outer.class"]
	inner := byPath["example/Outer$Inner.class"]
	deep := byPath["example/Outer$Inner$Deep.class"]
	orphan := byPath["example/Orphan$Inner.class"]
	if inner.Parent != outer.ID || inner.Name != "Inner.class" || inner.Entry != 1 {
		t.Fatalf("inner class was not grouped under its outer class: %+v", inner)
	}
	if deep.Parent != inner.ID || deep.Name != "Deep.class" || deep.Entry != 2 {
		t.Fatalf("deep class was not grouped under its immediate class: %+v", deep)
	}
	if orphan.Parent == orphan.ID || orphan.Name != "Orphan$Inner.class" || orphan.Entry != 3 {
		t.Fatalf("orphan class was given a synthetic parent: %+v", orphan)
	}
	anonymous := byPath["example/Outer$1.class"]
	if anonymous.Parent != outer.ID || anonymous.Name != "Outer$1.class" {
		t.Fatalf("anonymous class lost its binary name: %+v", anonymous)
	}
	local := byPath["example/Outer$1Local.class"]
	if local.Parent != outer.ID || local.Name != "Outer$1Local.class" {
		t.Fatalf("local class lost its binary name: %+v", local)
	}
	visible := tree.Visible(map[int]bool{outer.ID: true, outer.Parent: true})
	foundInner := false
	for _, item := range visible {
		if item.Node.ID == inner.ID && item.Depth == 2 {
			foundInner = true
		}
	}
	if !foundInner {
		t.Fatalf("expanded outer class does not reveal its inner class: outer=%+v visible=%+v", outer, visible)
	}
}

func TestVisibleQueryKeepsOnlyMatchesAndAncestors(t *testing.T) {
	tree := BuildTree([]Entry{
		{ID: 0, Path: "com/example/Hello.class", Name: "Hello.class", Kind: EntryClass},
		{ID: 1, Path: "org/other/World.class", Name: "World.class", Kind: EntryClass},
	})
	visible := tree.VisibleQuery("hello")
	if len(visible) != 3 {
		t.Fatalf("filtered nodes: %+v", visible)
	}
	for index, want := range []string{"com", "example", "Hello.class"} {
		if visible[index].Node.Name != want || visible[index].Depth != index {
			t.Fatalf("filtered path %d is %+v, want %s", index, visible[index], want)
		}
	}
	if visible := tree.VisibleQuery("missing"); len(visible) != 0 {
		t.Fatalf("unexpected matches: %+v", visible)
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
