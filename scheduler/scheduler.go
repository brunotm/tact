package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/brunotm/sema"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/log"
	"github.com/robfig/cron"
)

// Scheduler type
type Scheduler struct {
	mtx     sync.Mutex
	ctx     context.Context          // Main context that will be propagated to running collectors
	cancel  context.CancelFunc       // Cancel function of main context
	grace   time.Duration            // Grace period before failing a start when acquiring a run slot
	sema    sema.Sema                // Semaphore to control maxTasks run slots
	cron    *cron.Cron               // The cron scheduler
	running map[string]*tact.Context // The store for current running ctxs
	wchan   chan []byte
}

// New returns a initialized scheduler
func New(maxTasks int, grace time.Duration, wchan chan []byte) (sched *Scheduler) {
	ctx, cancel := context.WithCancel(context.Background())

	sema, err := sema.New(maxTasks)
	if err != nil {
		panic(err)
	}

	return &Scheduler{
		mtx:     sync.Mutex{},
		ctx:     ctx,
		cancel:  cancel,
		grace:   grace,
		sema:    sema,
		cron:    cron.New(),
		running: make(map[string]*tact.Context),
		wchan:   wchan,
	}
}

// AddJob function
func (s *Scheduler) AddJob(spec string, coll *tact.Collector, node *tact.Node, ttl time.Duration) (err error) {
	jobname := fmt.Sprintf("%s/%s", coll.Name, node.HostName)

	fn := func() {
		ctx, err := tact.NewContext(s.ctx, coll.Name, node, tact.Store, ttl)
		if err != nil {
			log.Error(
				"scheduler creating new ctx",
				"collector", coll.Name, "node", node.HostName, "error", err.Error())
			return
		}
		if !s.sema.AcquireWithin(s.grace) {
			ctx.LogError("scheduler: Timeout waiting for slot")
			return
		}
		defer s.sema.Release()
		ctx.LogDebug("aquired scheduler run slot")

		if !s.addRun(jobname, ctx) {
			ctx.LogError("scheduler: Already running")
			return
		}

		coll.Start(ctx, s.wchan)

		if !s.removeRun(jobname) {
			ctx.LogError("scheduler: Not found for removal after completion")
		}
	}
	log.Info("schedule: add job", "collector", coll.Name, "node", node.HostName, "schedule", spec)
	return s.cron.AddFunc(spec, fn)
}

// Start the scheduler
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop the scheduler and wait for runnning jobs to finish
func (s *Scheduler) Stop() {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.cron.Stop()
	s.waitJobs()
	s.cancel()

}

// Cancel the scheduler and runnning jobs
func (s *Scheduler) Cancel() {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.cron.Stop()
	s.cancel()
	s.waitJobs()
}

func (s *Scheduler) waitJobs() {
	cancel := time.After(15 * time.Second)
	for s.sema.Holders() > 0 {
		select {
		case <-cancel:
			return
		case <-time.After(time.Second * 1):
			continue
		}
	}
}

func (s *Scheduler) addRun(name string, ctx *tact.Context) (ok bool) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, running := s.running[name]; running {
		return false
	}
	s.running[name] = ctx
	return true
}

func (s *Scheduler) removeRun(name string) (ok bool) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, running := s.running[name]; !running {
		return false
	}
	delete(s.running, name)

	return true
}

// func logStats() {
// 	for {
// 		var m runtime.MemStats
// 		runtime.ReadMemStats(&m)
// 		fmt.Printf("############# Goroutines running: %d\n", runtime.NumGoroutine())
// 		fmt.Printf("#############  Alloc = %v, TotalAlloc = %v, Sys = %v, NumGC = %v\n", m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC)
// 		time.Sleep(5 * time.Second)
// 	}
// }
