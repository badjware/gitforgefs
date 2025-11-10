package queue

import (
	"fmt"
	"log/slog"
	"sync/atomic"
)

type TaskQueue interface {
	AddTask(task Task) error
	Shutdown()

	worker(id int)
}

type Task func()

type memoryQueue struct {
	logger *slog.Logger

	workers int
	buffer  chan Task
	active  atomic.Bool
}

func NewMemoryQueue(logger *slog.Logger, name string, workerCount int, bufferSize int) TaskQueue {
	queue := memoryQueue{
		logger:  logger.With("queueName", name),
		workers: workerCount,
		buffer:  make(chan Task, bufferSize),
		active:  atomic.Bool{},
	}
	for i := 0; i < workerCount; i++ {
		go queue.worker(i)
	}
	queue.active.Store(true)
	logger.Debug("TaskQueue started")
	return &queue
}

func (q *memoryQueue) AddTask(task Task) error {
	if !q.active.Load() {
		return fmt.Errorf("task queue is shutdown")
	}

	select {
	case q.buffer <- task:
		q.logger.Debug("Task added to queue")
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}

func (q *memoryQueue) Shutdown() {
	if q.active.Load() {
		q.active.Store(false)
		close(q.buffer)
		q.logger.Debug("TaskQueue shutdown")
	}
}

func (q *memoryQueue) worker(id int) {
	logger := q.logger.With("workerId", id)
	logger.Debug("TaskQueue worker started")
	for task := range q.buffer {
		logger.Debug("TaskQueue worker processing task")
		task()
	}
	logger.Debug("TaskQueue worker shutdown")
}
