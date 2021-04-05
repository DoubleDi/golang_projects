package main

import "fmt"

type Item struct {
	next *Item
	key  string
	prev *Item
}

type Data struct {
	data  interface{}
	index *Item
}

type Cache struct {
	m     map[string]*Data
	start *Item
	end   *Item
	cap   int
}

func NewCache(cap int) *Cache {
	return &Cache{
		cap: cap,
		m:   make(map[string]*Data, cap),
	}
}

func (c *Cache) Add(key string, value interface{}) {
	if _, ok := c.m[key]; !ok {
		if len(c.m) >= c.cap {
			oldKey := c.start.key
			c.start.next.prev = nil
			next := c.start.next
			c.start.next = nil
			c.start = next
			delete(c.m, oldKey)
		}
	} else {
		c.replaceItem(key)
	}
	elem := c.addItem(key, value)
	c.m[key] = &Data{
		data:  value,
		index: elem,
	}
}

func (c *Cache) replaceItem(key string) {
	e := c.m[key].index
	if e.prev != nil {
		e.prev.next = e.next
		e.prev = nil
	} else {
		c.start = c.start.next
	}
	if e.next != nil {
		e.next.prev = e.prev
		e.next = nil
	} else {
		c.end = c.end.prev
	}
}

func (c *Cache) addItem(key string, value interface{}) *Item {
	var elem *Item
	if c.start == nil && c.end == nil {
		c.start = &Item{key: key}
		c.end = c.start
		elem = c.end
	} else {
		elem = &Item{key: key}
		c.end.next = elem
		elem.prev = c.end
		c.end = c.end.next
	}
	return elem
}

func (c *Cache) Get(key string) (interface{}, bool) {
	data, ok := c.m[key]
	if ok {
		c.replaceItem(key)
		elem := c.addItem(key, data.data)
		c.m[key].index = elem
		return data.data, ok
	}
	return nil, ok
}

func main() {
	c := NewCache(3)
	c.Add("1", 1)
	c.Add("2", 2)
	c.Add("3", 3)
	c.Add("2", 41)
	c.Add("4", 4)
	fmt.Println(c.Get("1"))
	fmt.Println(c.Get("2"))

}
