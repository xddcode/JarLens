package archive

import (
	"path"
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
		id := len(tree.Nodes)
		kind := NodeResource
		if entry.Kind == EntryClass {
			kind = NodeClass
		}
		tree.Nodes = append(tree.Nodes, Node{ID: id, Parent: parent, Entry: entry.ID, Path: entry.Path, Name: entry.Name, Kind: kind})
		tree.attach(parent, id)
	}
	tree.groupNestedClasses()
	tree.sortChildren(tree.Roots)
	for id := range tree.Nodes {
		tree.sortChildren(tree.Nodes[id].Children)
	}
	return tree
}

func (t *Tree) groupNestedClasses() {
	classes := make(map[string]int)
	for _, node := range t.Nodes {
		if node.Kind == NodeClass {
			classes[node.Path] = node.ID
		}
	}
	for _, node := range t.Nodes {
		if node.Kind != NodeClass {
			continue
		}
		extension := path.Ext(node.Name)
		name := strings.TrimSuffix(node.Name, extension)
		for index := strings.LastIndex(name, "$"); index > 0; index = strings.LastIndex(name[:index], "$") {
			parentPath := path.Join(path.Dir(node.Path), name[:index]+extension)
			parentID, exists := classes[parentPath]
			if !exists || parentID == node.ID {
				continue
			}
			t.detach(node.Parent, node.ID)
			t.Nodes[node.ID].Parent = parentID
			t.Nodes[node.ID].Name = name[index+1:] + extension
			t.attach(parentID, node.ID)
			break
		}
	}
}

func (t *Tree) attach(parent, child int) {
	if parent < 0 {
		t.Roots = append(t.Roots, child)
		return
	}
	t.Nodes[parent].Children = append(t.Nodes[parent].Children, child)
}

func (t *Tree) detach(parent, child int) {
	children := t.Roots
	if parent >= 0 {
		children = t.Nodes[parent].Children
	}
	for index, id := range children {
		if id == child {
			children = append(children[:index], children[index+1:]...)
			break
		}
	}
	if parent < 0 {
		t.Roots = children
	} else {
		t.Nodes[parent].Children = children
	}
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

func (t *Tree) VisibleQuery(query string) []VisibleNode {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}
	matched := make(map[int]bool)
	var matches func([]int) bool
	matches = func(children []int) bool {
		found := false
		for _, id := range children {
			node := t.Nodes[id]
			childMatch := matches(node.Children)
			if strings.Contains(strings.ToLower(node.Name), query) || strings.Contains(strings.ToLower(node.Path), query) || childMatch {
				matched[id] = true
				found = true
			}
		}
		return found
	}
	matches(t.Roots)

	var visible []VisibleNode
	var walk func([]int, int)
	walk = func(children []int, depth int) {
		for _, id := range children {
			if !matched[id] {
				continue
			}
			node := t.Nodes[id]
			visible = append(visible, VisibleNode{Node: node, Depth: depth})
			walk(node.Children, depth+1)
		}
	}
	walk(t.Roots, 0)
	return visible
}

func (t *Tree) Visible(expanded map[int]bool) []VisibleNode {
	var visible []VisibleNode
	var walk func([]int, int)
	walk = func(children []int, depth int) {
		for _, id := range children {
			node := t.Nodes[id]
			visible = append(visible, VisibleNode{Node: node, Depth: depth})
			if len(node.Children) > 0 && expanded[id] {
				walk(node.Children, depth+1)
			}
		}
	}
	walk(t.Roots, 0)
	return visible
}
