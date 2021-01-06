package jobs

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestJobCalc(t *testing.T) {
	convey.Convey("测试时间计算", t, func() {
		scheduler := NewScheduler(100000)
		now := time.Now()
		job := scheduler.StartToday(20, 0, 0).JobType(ScheduledTask).Day(1)
		job.calcNextRun()
		fmt.Println(job.nextRun)
		expectedTime := time.Date(now.Year(), now.Month(), now.Day()+1, 20, 0, 0, 0, time.Local)
		convey.So(job.nextRun, convey.ShouldEqual, expectedTime)
	})
}
func TestJobRun(t *testing.T) {
	convey.Convey("测试任务运行", t, func() {
		scheduler := NewScheduler(100000)
		now := time.Now()
		err := scheduler.Start(now).JobType(ScheduledTask).Day(2).Hour(13).Minute(7).
			Second(45).Do("testTask", func() {
			fmt.Println(time.Now(), "task running!")
		})
		if err != nil {
			logrus.Error(err)
		}
		scheduler.Run()
		for {

		}
	})
}
func TestMultiJobRun(t *testing.T) {
	convey.Convey("测试任务运行", t, func() {
		scheduler := NewScheduler(100000)
		now := time.Now()
		err := scheduler.Start(now).JobType(RecurrenceTask).Second(5).Do("testTask", func() {
			fmt.Println(time.Now(), "task1 running!")
		})
		err = scheduler.Start(now).JobType(RecurrenceTask).Second(10).Do("testTask2", func() {
			fmt.Println(time.Now(), "task2 running!")
		})
		if err != nil {
			logrus.Error(err)
		}
		scheduler.Run()
		for {

		}
	})
}
func TestJobStop(t *testing.T) {
	convey.Convey("测试任务运行", t, func() {
		scheduler := NewScheduler(100000)
		now := time.Now()
		err := scheduler.Start(now).JobType(RecurrenceTask).Second(5).Do("testTask", func() {
			fmt.Println(time.Now(), "task1 running!")
		})
		err = scheduler.Start(now).JobType(RecurrenceTask).Second(10).Do("testTask2", func() {
			fmt.Println(time.Now(), "task2 running!")
		})
		if err != nil {
			logrus.Error(err)
		}
		scheduler.Run()
		time.Sleep(11 * time.Second)
		scheduler.Stop()
		for {

		}
	})
}
