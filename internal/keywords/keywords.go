package keywords

import (
	"strings"
)

type Node struct {
	children map[string]*Node
	value    string
	final    bool
}

func NewNode() *Node {
	node := Node{
		children: make(map[string]*Node),
	}
	return &node
}

type Trie struct {
	root *Node
}

func NewTrie() *Trie {
	trie := Trie{
		root: NewNode(),
	}
	return &trie
}

func NewTrieFrom(words [][]string) *Trie {
	t := NewTrie()
	for _, w := range words {
		t.Insert(w)
	}
	return t
}

func (t *Trie) Insert(words []string) {
	node := t.root
	for _, w := range words {
		w = strings.ToLower(w)
		if _, ok := node.children[w]; !ok {
			n := NewNode()
			n.value = w
			node.children[w] = n
		}
		node = node.children[w]
	}
	node.final = true
}

func (t *Trie) Search(words []string) (int, bool) {
	node := t.root
	for _, w := range words {
		w = strings.ToLower(w)
		if next, ok := node.children[w]; ok {
			node = next
		} else {
			return 0, false
		}
	}
	return len(node.children), node.final
}
