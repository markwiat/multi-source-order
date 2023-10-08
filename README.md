# multi-source-order
Sort elements from many sources

## Example

Below code shows how to sort ascending next fire times of quartz jobs.

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/markwiat/multi-source-order/order"
	"github.com/reugn/go-quartz/quartz"
)

const noTrigerErrorMsg = "next trigger time is in the past"

type elementTime time.Time

func (et elementTime) Before(other order.Element) bool {
	return time.Time(et).Before(time.Time(other.(elementTime)))
}

type scheduleJob struct {
	scheduleID string
	trigger    quartz.Trigger
}

func (sj *scheduleJob) Description() string {
	return fmt.Sprintf("Schedule: %s", sj.scheduleID)
}

func (sj *scheduleJob) Key() int {
	return quartz.HashCode(sj.scheduleID)
}

func (sj *scheduleJob) ContainerId() any {
	return sj.scheduleID
}

func (sj *scheduleJob) NextAfter(e order.Element) (order.Element, error) {
	n, err := sj.trigger.NextFireTime(time.Time(e.(elementTime)).UnixNano())
	if err != nil {
		if err.Error() == noTrigerErrorMsg {
			return nil, nil
		}
		return nil, err
	}

	return elementTime(time.Unix(0, n)), nil
}

func (sj *scheduleJob) Execute(ctx context.Context) {
	fmt.Printf("Execute at %v\n", time.Now().UTC())
}

func main() {

	now := time.Now().UTC()
	after5sec := now.Add(5 * time.Second)
	after10sec := now.Add(10 * time.Second)

	cronAfter5Sec := fmt.Sprintf("%d %d %d %d %d ? %d", after5sec.Second(), after5sec.Minute(), after5sec.Hour(), after5sec.Day(), after5sec.Month(), after5sec.Year())
	cronEverySec := "*/1 * * * * *"
	cronEvery4Sec := "*/4 * * * * *"

	ctx := context.Background()
	sched := quartz.NewStdScheduler()

	sched.Start(ctx)

	crons := []string{cronAfter5Sec, cronEverySec, cronEvery4Sec}
	scheduleIDs := []string{"after5sec", "everySec", "every4sec"}

	for i, cron := range crons {
		trigger, err := quartz.NewCronTrigger(cron)
		if err != nil {
			fmt.Printf("Bad expression: %s", err)
			return
		}
		job := &scheduleJob{scheduleID: scheduleIDs[i], trigger: trigger}
		err = sched.ScheduleJob(ctx, job, trigger)
		if err != nil {
			fmt.Printf("Failed to schedule: %s", err)
			return
		}
	}

	schedules := []order.Container{}
	keys := sched.GetJobKeys()
	for _, k := range keys {
		sch, err := sched.GetScheduledJob(k)
		if err != nil {
			fmt.Printf("Failed to get schedule: %s", err)
			return
		}
		sj := sch.Job.(*scheduleJob)
		schedules = append(schedules, sj)
	}

	constraint := order.CreateConstraint(order.WithSizeLimit(1000), order.WithHighestElemnt(elementTime(after10sec)))
	sorted, hasNext, err := order.GetSortedElements(elementTime(now), constraint, schedules)
	if err != nil {
		fmt.Printf("failed to sort schedules: %s", err)
		return
	}

	fmt.Printf("Time now: %v\n", now)
	for _, s := range sorted {
		at := time.Time(s.Element.(elementTime))
		scheduleID := s.ContainerId.(string)
		fmt.Printf("scheduleID: %s, time: %v\n", scheduleID, at)
	}
	fmt.Printf("has next: %t\n", hasNext)
}
```

Example result is
```text
Time now: 2023-10-08 13:24:02.253611608 +0000 UTC
scheduleID: everySec, time: 2023-10-08 15:24:03 +0200 CEST
scheduleID: everySec, time: 2023-10-08 15:24:04 +0200 CEST
scheduleID: every4sec, time: 2023-10-08 15:24:04 +0200 CEST
scheduleID: everySec, time: 2023-10-08 15:24:05 +0200 CEST
scheduleID: everySec, time: 2023-10-08 15:24:06 +0200 CEST
scheduleID: everySec, time: 2023-10-08 15:24:07 +0200 CEST
scheduleID: after5sec, time: 2023-10-08 15:24:07 +0200 CEST
scheduleID: everySec, time: 2023-10-08 15:24:08 +0200 CEST
scheduleID: every4sec, time: 2023-10-08 15:24:08 +0200 CEST
scheduleID: everySec, time: 2023-10-08 15:24:09 +0200 CEST
scheduleID: everySec, time: 2023-10-08 15:24:10 +0200 CEST
scheduleID: everySec, time: 2023-10-08 15:24:11 +0200 CEST
scheduleID: everySec, time: 2023-10-08 15:24:12 +0200 CEST
scheduleID: every4sec, time: 2023-10-08 15:24:12 +0200 CEST
has next: false
```