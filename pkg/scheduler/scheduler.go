package scheduler

import (
	"context"
	"sync"
	"time"
)

type Job struct {
	Interval time.Duration
	Task     func()
}

func (j *Job) Start(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()

	j.Task()

	ticker := time.NewTicker(j.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			j.Task()

		case <-ctx.Done():
			return
		}
	}
}

type Scheduler struct {
	Jobs   []*Job
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *Scheduler) AddJob(interval time.Duration, task func()) {
	job := &Job{
		Interval: interval,
		Task:     task,
	}

	s.Jobs = append(s.Jobs, job)
}

func (s *Scheduler) Start() {
	for _, job := range s.Jobs {
		s.wg.Add(1)

		go job.Start(&s.wg, s.ctx)
	}
}

func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
}

func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		Jobs:   make([]*Job, 0),
		ctx:    ctx,
		cancel: cancel,
	}
}
