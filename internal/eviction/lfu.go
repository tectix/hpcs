package eviction

import (
	"container/heap"
	"github.com/tectix/hpcs/internal/cache"
)

type LFU struct {
	maxSize int64
	items   map[string]*lfuItem
	heap    *lfuHeap
}

type lfuItem struct {
	key      string
	entry    *cache.Entry
	useCount int
	index    int
}

type lfuHeap []*lfuItem

func (h lfuHeap) Len() int           { return len(h) }
func (h lfuHeap) Less(i, j int) bool { return h[i].useCount < h[j].useCount }
func (h lfuHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *lfuHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*lfuItem)
	item.index = n
	*h = append(*h, item)
}

func (h *lfuHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]
	return item
}

func NewLFU(maxSize int64) *LFU {
	h := &lfuHeap{}
	heap.Init(h)
	return &LFU{
		maxSize: maxSize,
		items:   make(map[string]*lfuItem),
		heap:    h,
	}
}

func (l *LFU) OnGet(key string, entry *cache.Entry) {
	if item, exists := l.items[key]; exists {
		item.useCount = entry.UseCount
		heap.Fix(l.heap, item.index)
	} else {
		item := &lfuItem{
			key:      key,
			entry:    entry,
			useCount: entry.UseCount,
		}
		heap.Push(l.heap, item)
		l.items[key] = item
	}
}

func (l *LFU) OnSet(key string, entry *cache.Entry) {
	if item, exists := l.items[key]; exists {
		item.entry = entry
		item.useCount = entry.UseCount
		heap.Fix(l.heap, item.index)
	} else {
		item := &lfuItem{
			key:      key,
			entry:    entry,
			useCount: entry.UseCount,
		}
		heap.Push(l.heap, item)
		l.items[key] = item
	}
}

func (l *LFU) OnDelete(key string) {
	if item, exists := l.items[key]; exists {
		heap.Remove(l.heap, item.index)
		delete(l.items, key)
	}
}

func (l *LFU) GetVictims(c *cache.Cache) []string {
	if c.Size() <= l.maxSize {
		return nil
	}
	
	var victims []string
	currentSize := c.Size()
	
	heapCopy := make(lfuHeap, len(*l.heap))
	copy(heapCopy, *l.heap)
	
	for len(heapCopy) > 0 && currentSize > l.maxSize {
		item := heap.Pop(&heapCopy).(*lfuItem)
		victims = append(victims, item.key)
		currentSize -= int64(len(item.entry.Value))
	}
	
	return victims
}