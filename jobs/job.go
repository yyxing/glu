package jobs

import (
	"errors"
	"log"
	"reflect"
	"time"
)

type Job struct {
	// job名字
	taskName string
	// 触发间隔
	interval int64
	// 上次运行时间
	lastRun time.Time
	// 下次运行时间
	nextRun time.Time
	// 时间单位
	unit timeUnit
	// 执行函数
	taskFunc func()
	// 时区
	loc *time.Location
	// 是否需要同步
	lock bool
	// 任务类型
	jobType jobType
	// 是否有异常
	err error
	// 定时任务的触发时间 每一天的具体时刻
	schedulerTriggerTime map[timeUnit]int64
}

var (
	ErrTimeFormat         = errors.New("time format error")
	ErrPeriodNotSpecified = errors.New("unspecified job period")
)

func (j *Job) shouldRun() bool {
	return time.Now().Unix() >= j.nextRun.Unix()
}

func (j *Job) run() {
	j.taskFunc()
}

// 设置任务类型
func (j *Job) JobType(jobType jobType) *Job {
	j.jobType = jobType
	return j
}

func (j *Job) checkInterval(interval int64) {
	if interval < 1 {
		log.Println("interval must more than 0")
		j.err = errors.New("interval must more than 0")
	}
}

// 设置秒
func (j *Job) Second(interval int64) *Job {
	if j.jobType == ScheduledTask {
		j.schedulerTriggerTime[seconds] = interval
	} else {
		j.checkInterval(interval)
		j.unit = seconds
		j.interval = interval
	}
	return j
}

// 设置分
func (j *Job) Minute(interval int64) *Job {
	if j.jobType == ScheduledTask {
		j.schedulerTriggerTime[minutes] = interval
	} else {
		j.checkInterval(interval)
		j.unit = minutes
		j.interval = interval
	}
	return j
}

// 设置时
func (j *Job) Hour(interval int64) *Job {
	if j.jobType == ScheduledTask {
		j.schedulerTriggerTime[hours] = interval
	} else {
		j.checkInterval(interval)
		j.unit = hours
		j.interval = interval
	}
	return j
}

// 设置天
func (j *Job) Day(interval int64) *Job {
	j.checkInterval(interval)
	if j.jobType == ScheduledTask {
		j.schedulerTriggerTime[days] = interval
	} else {
		j.unit = days
		j.interval = interval
	}
	return j
}

// 计算下次运行时间
func (j *Job) calcNextRun() {
	if j.jobType == ScheduledTask {
		j.nextRun = j.lastRun.Add(time.Duration(j.schedulerTriggerTime[days]) * time.Hour * 24)
	} else {
		switch j.unit {
		case seconds:
			j.nextRun = j.lastRun.Add(time.Duration(j.interval) * time.Second)
		case minutes:
			j.nextRun = j.lastRun.Add(time.Duration(j.interval) * time.Minute)
		case hours:
			j.nextRun = j.lastRun.Add(time.Duration(j.interval) * time.Hour)
		case days:
			j.nextRun = j.lastRun.Add(time.Duration(j.interval) * time.Hour * 24)
		default:
			j.err = ErrPeriodNotSpecified
		}
	}
}

// 设置执行函数
func (j *Job) Do(taskName string, task func()) error {
	if len(taskName) == 0 {
		typ := reflect.TypeOf(task)
		j.taskName = typ.Name()
	} else {
		j.taskName = taskName
	}
	if j.err != nil {
		log.Printf("%s task has error, error message: %s\n", taskName, j.err.Error())
		return j.err
	}
	now := time.Now().In(j.loc)
	j.taskFunc = task
	if j.jobType == ScheduledTask {
		j.lastRun = time.Date(j.lastRun.Year(), j.lastRun.Month(), j.lastRun.Day(), int(j.schedulerTriggerTime[hours]),
			int(j.schedulerTriggerTime[minutes]), int(j.schedulerTriggerTime[seconds]), 0, j.loc)
		j.nextRun = time.Date(j.lastRun.Year(), j.lastRun.Month(), j.lastRun.Day(), int(j.schedulerTriggerTime[hours]),
			int(j.schedulerTriggerTime[minutes]), int(j.schedulerTriggerTime[seconds]), 0, j.loc)
	}
	if !j.nextRun.After(now) {
		// 修正时间为设置的时分秒
		j.calcNextRun()
	}
	return nil
}
