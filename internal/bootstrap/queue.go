package bootstrap

import (
	"github.com/weavatar/weavatar/pkg/queue"
)

func NewQueue() *queue.Queue {
	return queue.New(1000)
}
