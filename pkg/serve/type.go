package serve

import (
	"sync"
)

// type RequestItem struct {
// 	ID     string
// 	Notify chan []byte
// }

// var RequestQueue = make([]RequestItem, 0)
// var QueueLock = sync.Mutex{}

var requestQueue = RequestQueue{
	mutex:    sync.RWMutex{},
	elements: make(map[string]chan []byte),
}

type RequestQueue struct {
	mutex    sync.RWMutex
	elements map[string]chan []byte
}

func (sm *RequestQueue) Load(key string) (chan []byte, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	val, ok := sm.elements[key]
	return val, ok
}

func (sm *RequestQueue) Store(key string, value chan []byte) (chan []byte, bool) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	old, ok := sm.elements[key]
	sm.elements[key] = value
	return old, ok
}

func (sm *RequestQueue) Delete(key string) (chan []byte, bool) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	value, ok := sm.elements[key]
	delete(sm.elements, key)
	return value, ok
}

func AddRequestItem(id string, notify chan []byte) {
	requestQueue.Store(id, notify)
}

func DeleteRequestItem(id string) {
	requestQueue.Delete(id)
}

func ExistRequestItem(id string) bool {
	_, ok := requestQueue.Load(id)
	return ok
}

// func DeleteRequestItem(id string) {
// 	QueueLock.Lock()
// 	for i, item := range RequestQueue {
// 		if item.ID == id {
// 			RequestQueue = append(RequestQueue[:i], RequestQueue[i+1:]...)
// 			close(item.Notify)
// 			break
// 		}
// 	}
// 	QueueLock.Unlock()
// }

func WriteAndDeleteRequestItem(id string, data []byte) {
	if value, ok := requestQueue.Load(id); ok && value != nil {
		value <- data
		close(value)
	}
	requestQueue.Delete(id)
}
