package cache

type Node struct {
	prev *Node
	data []byte
	next *Node
}

type LRUCache struct {
	head        *Node
	urlToAddr   map[string]*Node
	currentSize int
	maxSize     int
}

func NewLRUCache(maxSize int) *LRUCache {
	headNode := &Node{}
	headNode.prev = headNode
	headNode.next = headNode

	return &LRUCache{
		urlToAddr:   make(map[string]*Node),
		maxSize:     maxSize,
		currentSize: 0,
		head:        headNode,
	}
}
func (cache *LRUCache) Put(url string, resp []byte) {

	if cache.currentSize >= cache.maxSize {
		lastNode := cache.head.prev
		lastNode.prev.next = lastNode.next
		lastNode.next.prev = lastNode.prev
		cache.currentSize--
	}
	currNode := &Node{data: resp}

	currNode.next = cache.head.next
	cache.head.next.prev = currNode
	currNode.prev = cache.head
	cache.head = currNode

	cache.urlToAddr[url] = currNode
	cache.currentSize++
}

func (cache *LRUCache) Get(url string) []byte {
	var resp []byte = nil
	val, ok := cache.urlToAddr[url]
	if ok {
		resp = val.data
		if cache.head != val {
			currNode := val

			currNode.prev.next = currNode.next
			currNode.next.prev = currNode.prev

			currNode.next = cache.head.next
			cache.head.next.prev = currNode
			currNode.prev = cache.head
			cache.head = currNode

			cache.urlToAddr[url] = currNode
		}
	}
	return resp
}
