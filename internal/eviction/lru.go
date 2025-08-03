package eviction

import (
	"container/list"
	"github.com/tectix/hpcs/internal/cache"
)

type LRU struct {
	maxSize int64
	list    *list.List
	items   map[string]*list.Element
}

type lruItem struct {
	key   string
	entry *cache.Entry
}

func NewLRU(maxSize int64) *LRU {
	return &LRU{
		maxSize: maxSize,
		list:    list.New(),
		items:   make(map[string]*list.Element),
	}
}

func (l *LRU) OnGet(key string, entry *cache.Entry) {
	if elem, exists := l.items[key]; exists {
		l.list.MoveToFront(elem)
	} else {
		item := &lruItem{key: key, entry: entry}
		elem := l.list.PushFront(item)
		l.items[key] = elem
	}
}

func (l *LRU) OnSet(key string, entry *cache.Entry) {
	if elem, exists := l.items[key]; exists {
		l.list.MoveToFront(elem)
		elem.Value.(*lruItem).entry = entry
	} else {
		item := &lruItem{key: key, entry: entry}
		elem := l.list.PushFront(item)
		l.items[key] = elem
	}
}

func (l *LRU) OnDelete(key string) {
	if elem, exists := l.items[key]; exists {
		l.list.Remove(elem)
		delete(l.items, key)
	}
}

func (l *LRU) GetVictims(c *cache.Cache) []string {
	if c.Size() <= l.maxSize {
		return nil
	}
	
	var victims []string
	currentSize := c.Size()
	
	for elem := l.list.Back(); elem != nil && currentSize > l.maxSize; elem = elem.Prev() {
		item := elem.Value.(*lruItem)
		victims = append(victims, item.key)
		currentSize -= int64(len(item.entry.Value))
	}
	
	return victims
}