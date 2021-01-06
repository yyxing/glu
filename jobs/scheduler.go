package jobs

import (
	"errors"
	"log"
	"sort"
	"time"
)

type Scheduler struct {
	jobs    []*Job
	size    int
	head    int
	tail    int
	stopped chan bool
}

func (scheduler *Scheduler) Len() int { return scheduler.tail }
func (scheduler *Scheduler) Swap(i, j int) {
	jobs := scheduler.jobs
	jobs[i], jobs[j] = jobs[j], jobs[i]
}
func (scheduler *Scheduler) Less(i, j int) bool {
	jobs := scheduler.jobs
	return jobs[i].nextRun.Before(jobs[j].nextRun)
}

// 开始运行
func (scheduler *Scheduler) Run() {
	scheduler.stopped = make(chan bool, 1)
	log.Println("scheduler startTime: ", time.Now().Format("2006-01-02 15:04:05"))
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				runnableJobs := scheduler.getRunnableJobs()
				sort.Sort(scheduler)
				for _, job := range runnableJobs {
					job.lastRun = time.Now()
					go job.run()
					job.calcNextRun()
				}
			case <-scheduler.stopped:
				ticker.Stop()
				log.Println("scheduler stop success")
				return
			}
		}
	}()

}
func (scheduler *Scheduler) getJobs() []*Job {
	return scheduler.jobs[:scheduler.tail]
}
func (scheduler *Scheduler) getRunnableJobs() []*Job {
	sort.Sort(scheduler)
	jobs := scheduler.getJobs()
	for i, job := range jobs {
		if !job.shouldRun() {
			return scheduler.jobs[:i]
		}
	}
	return jobs
}
func (scheduler *Scheduler) Start(startTime time.Time) *Job {
	now := time.Now()
	var err error
	if now.After(startTime) {
		err = errors.New("startTime must be later than now")
	}
	job := &Job{
		lastRun:              startTime,
		nextRun:              time.Unix(0, 0),
		unit:                 0,
		taskFunc:             nil,
		loc:                  time.Local,
		lock:                 false,
		jobType:              0,
		err:                  err,
		schedulerTriggerTime: map[timeUnit]int64{},
	}
	scheduler.jobs[scheduler.tail] = job
	scheduler.tail++
	return job
}
func (scheduler *Scheduler) StartToday(hour int, minute int, second int) *Job {
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second,
		0, time.FixedZone("CST", 8*3600))
	var err error
	if now.After(startTime) {
		err = errors.New("startTime must be later than now")
	}
	job := &Job{
		lastRun:              startTime,
		nextRun:              time.Unix(0, 0),
		unit:                 0,
		taskFunc:             nil,
		loc:                  time.Local,
		lock:                 false,
		jobType:              0,
		err:                  err,
		schedulerTriggerTime: map[timeUnit]int64{},
	}
	scheduler.jobs[scheduler.tail] = job
	scheduler.tail++
	return job
}
func (scheduler *Scheduler) Stop() {
	if scheduler.stopped != nil {
		scheduler.stopped <- true
	} else {
		log.Println("stop error. must run scheduler before stop")
	}
}
