package jobs

type timeUnit int

const (
	seconds timeUnit = iota + 1
	minutes
	hours
	days
)

type jobType int

const (
	// 定时任务 再某个循环的某个时间点执行
	ScheduledTask jobType = iota + 1
	// 按照一定的时间间隔循环运行
	RecurrenceTask
)

func NewScheduler(size int) *Scheduler {
	return &Scheduler{size: size, jobs: make([]*Job, size)}
}
