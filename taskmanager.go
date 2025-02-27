package app

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

var servicePool map[string]*service

const (
	StatusStopped int = iota
	StatusRunning
	StatusExecuting
	StatusInterrupted
)

var statuses = []string{"stopped", "running", "executing", "interrupted"}

type (
	service struct {
		ticker    *time.Ticker
		mux       sync.Mutex
		status    int
		task      Task
		interval  uint64
		ctx       context.Context
		cancel    context.CancelFunc // for interrupt task
		interrupt chan bool          // for interrupt service
	}

	// Task - интерфейс для задания
	Task interface {
		Init() error
		Id() string
		Description() string
		Execute(ctx context.Context, done chan bool)
	}

	Info struct {
		Id          string
		Description string
		Interval    uint64
		Status      string
	}
)

func init() {
	servicePool = make(map[string]*service)
}

func AddService(task Task) error {
	if _, exists := servicePool[task.Id()]; exists {
		return errors.New("service already exists")
	}

	servicePool[task.Id()] = &service{task: task, status: StatusStopped}

	return nil
}

func StartService(id string, interval uint64) error {
	j, exists := servicePool[id]
	if !exists {
		return errors.New("service not exists and can't be started")
	}
	return j.start(interval)
}

func StopService(id string) error {
	j, exists := servicePool[id]
	if !exists {
		return errors.New("service not exists and can't be started")
	}

	j.stop()
	return nil
}

func StopAll() {
	for _, j := range servicePool {
		if j.ticker != nil {
			j.stop()
		}
	}
}

func ServicesInfo() []Info {
	var res []Info
	for _, j := range servicePool {
		res = append(res, j.info())
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Id < res[j].Id
	})

	return res
}

func ServiceInfo(id string) Info {
	j, exists := servicePool[id]
	if !exists {
		return Info{}
	}
	return j.info()
}

func (j *service) info() Info {
	j.mux.Lock()
	defer j.mux.Unlock()
	return Info{Id: j.task.Id(), Description: j.task.Description(), Interval: j.interval, Status: statuses[j.status]}
}

func (j *service) start(interval uint64) error {
	j.mux.Lock()
	if j.status != StatusStopped {
		j.mux.Unlock()
		return errors.New("service is already started")
	}

	j.status = StatusRunning
	j.interval = interval
	j.ticker = time.NewTicker(time.Second)
	j.interrupt = make(chan bool)
	j.mux.Unlock()

	go func() {
		done := make(chan bool) // for task
		// main event loop
		for {
			select {
			case <-j.ticker.C: // execute
				j.ticker.Stop()
				j.mux.Lock()
				if j.status != StatusRunning {
					j.mux.Unlock()
					continue
				}
				j.status = StatusExecuting
				j.ctx, j.cancel = context.WithCancel(context.Background())
				j.mux.Unlock()

				go j.task.Execute(j.ctx, done)

			case <-j.interrupt: // stop service, Executor not started
				j.mux.Lock()
				j.status = StatusStopped
				j.ticker.Stop()
				j.ctx, j.cancel = context.WithCancel(context.Background())
				j.mux.Unlock()
				return

			case <-done: // task done
				j.mux.Lock()

				if j.status == StatusInterrupted {
					j.status = StatusStopped
					j.mux.Unlock()
					return
				}
				if j.interval != 0 {
					j.status = StatusRunning
					j.ticker = time.NewTicker(time.Duration(j.interval) * time.Second)
					j.mux.Unlock()
				} else {
					j.status = StatusStopped
					j.mux.Unlock()
					return
				}
			}
		}
	}()

	return nil
}

func (j *service) stop() {
	j.mux.Lock()
	defer j.mux.Unlock()

	interrupter := func() {}
	switch j.status {
	case StatusExecuting:
		interrupter = j.cancel
	case StatusRunning:
		interrupter = func() {
			j.interrupt <- true
		}
	default:
		return
	}
	j.status = StatusInterrupted
	interrupter()
}
