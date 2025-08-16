package service_tasks_dispatcher

import "context"

type TaskType int

const (
	TaskCreateUser TaskType = iota
	TaskGetUser
	TaskCreateOrder
	TaskGetUserOrders
	TaskCreateWithdrawal
	TaskGetUserWithdrawals
)

type Task struct {
	Type     TaskType
	Context  context.Context
	Payload  interface{}
	ResultCh chan TaskResult
}

type TaskResult struct {
	Result interface{}
	Err    error
}

type TaskDispatcher struct {
	taskQueue chan Task
}

func NewTaskDispatcher() *TaskDispatcher {
	return &TaskDispatcher{
		taskQueue: make(chan Task, 300),
	}
}

// Enqueue добавляет задачу в очередь.
func (d *TaskDispatcher) Enqueue(task Task) (interface{}, error) {
	if task.ResultCh == nil {
		task.ResultCh = make(chan TaskResult, 1)
	}

	d.taskQueue <- task

	select {
	case <-task.Context.Done():
		return nil, task.Context.Err()
	case res := <-task.ResultCh:
		return res.Result, res.Err
	}
}

// StartWorker запускает обработчик задач.
func (d *TaskDispatcher) StartWorker(handler func(Task) (interface{}, error)) {
	go func() {
		for task := range d.taskQueue {
			result, err := handler(task)
			if task.ResultCh != nil {
				task.ResultCh <- TaskResult{Result: result, Err: err}
				close(task.ResultCh)
			}
		}
	}()
}
