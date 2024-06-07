package serve

import (
	"sync"
)

type RequestItem struct {
	ID     string
	Notify chan []byte
}

var RequestQueue = make([]RequestItem, 0)
var QueueLock = sync.Mutex{}

func DeleteRequestItem(id string) {
	QueueLock.Lock()
	for i, item := range RequestQueue {
		if item.ID == id {
			RequestQueue = append(RequestQueue[:i], RequestQueue[i+1:]...)
			close(item.Notify)
			break
		}
	}
	QueueLock.Unlock()
}

func WriteAndDeleteRequestItem(id string, data []byte) {
	QueueLock.Lock()
	for i, item := range RequestQueue {
		if item.ID == id {
			item.Notify <- data
			RequestQueue = append(RequestQueue[:i], RequestQueue[i+1:]...)
			close(item.Notify)
			break
		}
	}
	QueueLock.Unlock()
}
