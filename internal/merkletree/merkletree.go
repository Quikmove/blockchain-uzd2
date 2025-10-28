package merkletree

import (
	"crypto/sha256"
	"fmt"
	"io"
)

type Hash32 [32]byte
type MerkleTree struct {
	Root *Node
}
type Node struct {
	Val   Hash32
	Left  *Node
	Right *Node
}

func NewMerkleTree(hashes []Hash32) *MerkleTree {
	if len(hashes) == 0 {
		return &MerkleTree{Root: nil}
	}
	var nodes []*Node
	for _, h := range hashes {
		nodes = append(nodes, &Node{Val: h})
	}
	for len(nodes) > 1 {
		if len(nodes)%2 == 1 {
			nodes = append(nodes, nodes[len(nodes)-1])
		}
		var newLevel []*Node
		for i := 0; i < len(nodes); i += 2 {
			left := nodes[i]
			right := nodes[i+1]
			parentHash := doubleHashPair(left.Val, right.Val)
			parentNode := &Node{
				Val:   parentHash,
				Left:  left,
				Right: right,
			}
			newLevel = append(newLevel, parentNode)
		}
		nodes = newLevel
	}
	return &MerkleTree{Root: nodes[0]}
}
func doubleHashPair(left, right Hash32) Hash32 {
	var buf [64]byte
	copy(buf[:32], left[:])
	copy(buf[32:], right[:])
	data := buf[:]

	hash := sha256.Sum256(data)
	doubleHash := sha256.Sum256(hash[:])
	return doubleHash
}

// print in levels
func (t *MerkleTree) PrintTree(io io.Writer) {
	if t.Root == nil {
		return
	}
	type levelNodes struct {
		level int
		nodes []*Node
	}
	queue := []levelNodes{{level: 0, nodes: []*Node{t.Root}}}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, node := range current.nodes {
			fmt.Fprintf(io, "Level %d: %x\n", current.level, node.Val)
			if node.Left != nil {
				queue = append(queue, levelNodes{level: current.level + 1, nodes: []*Node{node.Left}})
			}
			if node.Right != nil {
				queue = append(queue, levelNodes{level: current.level + 1, nodes: []*Node{node.Right}})
			}
		}
	}
}
