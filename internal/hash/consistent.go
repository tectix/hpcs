package hash

import (
	"crypto/sha1"
	"fmt"
	"sort"
	"strconv"
)

type Ring struct {
	nodes        map[uint32]string
	sortedHashes []uint32
	virtualNodes int
}

func NewRing(virtualNodes int) *Ring {
	return &Ring{
		nodes:        make(map[uint32]string),
		virtualNodes: virtualNodes,
	}
}

func (r *Ring) AddNode(node string) {
	for i := 0; i < r.virtualNodes; i++ {
		virtualKey := node + ":" + strconv.Itoa(i)
		hash := r.hashKey(virtualKey)
		r.nodes[hash] = node
		r.sortedHashes = append(r.sortedHashes, hash)
	}
	sort.Slice(r.sortedHashes, func(i, j int) bool {
		return r.sortedHashes[i] < r.sortedHashes[j]
	})
}

func (r *Ring) RemoveNode(node string) {
	for i := 0; i < r.virtualNodes; i++ {
		virtualKey := node + ":" + strconv.Itoa(i)
		hash := r.hashKey(virtualKey)
		delete(r.nodes, hash)
		r.removeFromSorted(hash)
	}
}

func (r *Ring) GetNode(key string) string {
	if len(r.sortedHashes) == 0 {
		return ""
	}
	
	hash := r.hashKey(key)
	idx := sort.Search(len(r.sortedHashes), func(i int) bool {
		return r.sortedHashes[i] >= hash
	})
	
	if idx == len(r.sortedHashes) {
		idx = 0
	}
	
	return r.nodes[r.sortedHashes[idx]]
}

func (r *Ring) GetNodes(key string, count int) []string {
	if len(r.sortedHashes) == 0 || count <= 0 {
		return nil
	}
	
	hash := r.hashKey(key)
	idx := sort.Search(len(r.sortedHashes), func(i int) bool {
		return r.sortedHashes[i] >= hash
	})
	
	if idx == len(r.sortedHashes) {
		idx = 0
	}
	
	seen := make(map[string]bool)
	var result []string
	
	for len(result) < count && len(seen) < len(r.getUniqueNodes()) {
		node := r.nodes[r.sortedHashes[idx]]
		if !seen[node] {
			result = append(result, node)
			seen[node] = true
		}
		idx = (idx + 1) % len(r.sortedHashes)
	}
	
	return result
}

func (r *Ring) Nodes() []string {
	return r.getUniqueNodes()
}

func (r *Ring) hashKey(key string) uint32 {
	h := sha1.New()
	h.Write([]byte(key))
	hashBytes := h.Sum(nil)
	return uint32(hashBytes[0])<<24 | uint32(hashBytes[1])<<16 | uint32(hashBytes[2])<<8 | uint32(hashBytes[3])
}

func (r *Ring) removeFromSorted(hash uint32) {
	idx := sort.Search(len(r.sortedHashes), func(i int) bool {
		return r.sortedHashes[i] >= hash
	})
	
	if idx < len(r.sortedHashes) && r.sortedHashes[idx] == hash {
		r.sortedHashes = append(r.sortedHashes[:idx], r.sortedHashes[idx+1:]...)
	}
}

func (r *Ring) getUniqueNodes() []string {
	unique := make(map[string]bool)
	for _, node := range r.nodes {
		unique[node] = true
	}
	
	var result []string
	for node := range unique {
		result = append(result, node)
	}
	return result
}