package cache

import (
	"fmt"
	"sync"
)

// ProxyItem represents a cached HTTP response
type ProxyItem struct {
	Status string
	Header string
	Body   []byte
}

// node represents a node in the doubly-linked list
type node struct {
	key   string
	value *ProxyItem
	next  *node
	prev  *node
}

// LRUCache implements a thread-safe Least Recently Used cache
type LRUCache struct {
	mu       sync.Mutex
	size     int
	capacity int
	head     *node
	tail     *node
	items    map[string]*node
}

// NewLRUCache creates a new LRU cache with the specified capacity
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*node),
	}
}

// Get retrieves an item from the cache and marks it as most recently used
func (c *LRUCache) Get(key string) (bool, *ProxyItem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	n, ok := c.items[key]
	if !ok {
		return false, nil
	}

	c.removeNode(n)
	c.moveToFront(n)

	return true, n.value
}

// Put adds or updates an item in the cache
func (c *LRUCache) Put(key string, value *ProxyItem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key exists, update it
	if existingNode, ok := c.items[key]; ok {
		existingNode.value = value
		c.removeNode(existingNode)
		c.moveToFront(existingNode)
		return
	}

	// Create new node
	newNode := &node{
		key:   key,
		value: value,
	}

	// Add to map and list
	c.items[key] = newNode
	c.moveToFront(newNode)
	c.size++

	// If this is the first item, set tail
	if c.tail == nil {
		c.tail = newNode
	}

	// Evict least recently used item if cache is full
	if c.size > c.capacity {
		c.evictLRU()
	}
}

// removeNode removes a node from the doubly-linked list
func (c *LRUCache) removeNode(n *node) {
	if c.head == n && c.tail == n {
		// Only node in the list
		c.head = nil
		c.tail = nil
	} else if n == c.head {
		// Node is at the head
		c.head = n.next
		if c.head != nil {
			c.head.prev = nil
		}
	} else if n == c.tail {
		// Node is at the tail
		c.tail = n.prev
		if c.tail != nil {
			c.tail.next = nil
		}
	} else {
		// Node is in the middle
		n.prev.next = n.next
		n.next.prev = n.prev
	}

	n.prev = nil
	n.next = nil
}

// moveToFront moves a node to the front (most recently used position)
func (c *LRUCache) moveToFront(n *node) {
	if c.head == n {
		return
	}

	// If this is the first node
	if c.head == nil {
		c.head = n
		c.tail = n
		n.prev = nil
		n.next = nil
		return
	}

	// Insert at the front
	n.next = c.head
	c.head.prev = n
	c.head = n
	n.prev = nil
}

// evictLRU removes the least recently used item from the cache
func (c *LRUCache) evictLRU() {
	if c.tail == nil {
		return
	}

	// Remove from map
	delete(c.items, c.tail.key)
	c.removeNode(c.tail)
	c.size--
}

// Size returns the current number of items in the cache
func (c *LRUCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.size
}

// Capacity returns the maximum capacity of the cache
func (c *LRUCache) Capacity() int {
	return c.capacity
}

// Display prints the cache contents for debugging (not thread-safe, use with caution)
func (c *LRUCache) Display() {
	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Print("Cache: ")
	n := c.head
	for n != nil {
		fmt.Printf("%s -> ", n.key)
		n = n.next
	}
	fmt.Println("END")
}
