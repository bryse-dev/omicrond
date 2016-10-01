package job

import (
  "fmt"
  "time"
  "github.com/BurntSushi/toml"
)

type WeekDay int

const (
  Sunday = 0
  Monday = 1
  Tuesday = 2
  Wednesday = 3
  Thursday = 4
  Friday = 5
  Saturday = 6
)

type JobHandler struct {
  Job []JobConfig
}

type JobConfig struct {
  Title           string // The name of the job.  Used in logging
  Command         string // String to be run on the system
  GroupName       string // Used to relate jobs and in logging *unused*
  ClassicSchedule string // Traditional encoded string to represent the schedule
  Schedule        ScheduleObj
}

type ScheduleObj struct {
  Minute    MinuteScope
  Hour      HourScope
  Day       DayScope
  Month     MonthScope
  DayOfWeek DayOfWeekScope
}

type Scope struct{}

type MinuteScope struct {
  Scope
  IsInterval bool
  Interval   durationObj // Time to wait before running the job again
  StartTime  []timeObj   // Jobs will not run from the start of the day until the earliest time is reached
  EndTime    timeObj     // Jobs will not run after this time and until the end of the day
}

type HourScope struct {
  Scope
  IsInterval bool
  Interval   durationObj // Time to wait before running the job again
  StartTime  []timeObj   // Jobs will not run from the start of the day until the earliest time is reached
  EndTime    timeObj     // Jobs will not run after this time and until the end of the day
}

type DayScope struct {
  Scope
  IsInterval bool
  Interval   durationObj // Time to wait before running the job again
  StartTime  []timeObj   // Jobs will not run from the start of the day until the earliest time is reached
  EndTime    timeObj     // Jobs will not run after this time and until the end of the day
}

type MonthScope struct {
  Scope
  IsInterval bool
  Interval   durationObj // Time to wait before running the job again
  StartTime  []timeObj   // Jobs will not run from the start of the day until the earliest time is reached
  EndTime    timeObj     // Jobs will not run after this time and until the end of the day
}

type DayOfWeekScope struct {
  Scope
  StartTime []WeekDay
}

func (h *JobHandler) ParseJobConfig(confFile string) {

  if _, err := toml.DecodeFile(confFile, &h); err != nil {
    fmt.Println(err)
    return
  }
}

type durationObj struct {
  time.Duration
}
type timeObj struct {
  time.Time
}

func (d *durationObj) UnmarshalText(text []byte) error {
  var err error
  d.Duration, err = time.ParseDuration(string(text))
  return err
}
func (t *timeObj) UnmarshalText(text []byte) error {
  var err error
  t.Time, err = time.Parse(time.Kitchen, string(text))
  return err
}
