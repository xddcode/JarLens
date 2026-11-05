package archive

import (
	"sort"
	"strings"
)

type NodeKind uint8

const (
	NodePackage NodeKind = iota
	NodeClass
	NodeResource
)

type Node struct {
	ID       int
	Parent   int
	Entry    int
	Path     string
	Name     string
	Kind     NodeKind
	Children []int
}

type Tree struct {
	Nodes []Node
	Roots []int
}

type VisibleNode struct {
	Node  Node
	Depth int
}

func BuildTree(entries []Entry) *Tree {
	ordered := append([]Entry(nil), entries...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Path == ordered[j].Path {
			return ordered[i].ID < ordered[j].ID
		}
		return ordered[i].Path < ordered[j].Path
	})

	tree := &Tree{}
	packages := make(map[string]int)
	for _, entry := range ordered {
		parts := strings.Split(entry.Path, "/")
		parent := -1
		prefix := ""
		for _, part := range parts[:len(parts)-1] {
			if prefix == "" {
				prefix = part
			} else {
				prefix += "/" + part
			}
			id, exists := packages[prefix]
			if !exists {
				id = len(tree.Nodes)
				packages[prefix] = id
				tree.Nodes = append(tree.Nodes, Node{ID: id, Parent: parent, Entry: -1, Path: prefix, Name: part, Kind: NodePackage})
				tree.attach(parent, id)
			}
			parent = id
		}
		name := entry.Name
		if entry.Kind == EntryClass {
			name = strings.TrimSuffix(name, ".class")
		}
		id := len(tree.Nodes)
		kind := NodeResource
		if entry.Kind == EntryClass {
			kind = NodeClass
		}
		tree.Nodes = append(tree.Nodes, Node{ID: id, Parent: parent, Entry: entry.ID, Path: entry.Path, Name: name, Kind: kind})
		tree.attach(parent, id)
	}
	tree.sortChildren(tree.Roots)
	for id := range tree.Nodes {
		tree.sortChildren(tree.Nodes[id].Children)
	}
	return tree
}

func (t *Tree) attach(parent, child int) {
	if parent < 0 {
		t.Roots = append(t.Roots, child)
		return
	}
	t.Nodes[parent].Children = append(t.Nodes[parent].Children, child)
}

func (t *Tree) sortChildren(children []int) {
	sort.SliceStable(children, func(i, j int) bool {
		left, right := t.Nodes[children[i]], t.Nodes[children[j]]
		if left.Kind != right.Kind {
			return left.Kind < right.Kind
		}
		if left.Name == right.Name {
			return left.ID < right.ID
		}
		return strings.ToLower(left.Name) < strings.ToLower(right.Name)
	})
}

func (t *Tree) Visible(expanded map[int]bool) []VisibleNode {
	var visible []VisibleNode
	var walk func([]int, int)
	walk = func(children []int, depth int) {
		for _, id := range children {
			node := t.Nodes[id]
			visible = append(visible, VisibleNode{Node: node, Depth: depth})
			if node.Kind == NodePackage && expanded[id] {
				walk(node.Children, depth+1)
			}
		}
	}
	walk(t.Roots, 0)
	return visible
}
