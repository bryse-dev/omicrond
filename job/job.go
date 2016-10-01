package job

import (
  "errors"
  "strings"
  "time"
  "github.com/BurntSushi/toml"
)

const (
  DAYOFWEEK = 4
  MONTH = 3
  DAY = 2
  HOUR = 1
  MINUTE = 0
  SUNDAY = "Sunday"
  MONDAY = "Monday"
  TUESDAY = "Tuesday"
  WEDNESDAY = "Wednesday"
  THURSDAY = "Thursday"
  FRIDAY = "Friday"
  SATURDAY = "Saturday"
)

type JobHandler struct {
  Job []JobConfig
}

type JobConfig struct {
  Title     string // The name of the job.  Used in logging
  Command   string // String to be run on the system
  GroupName string // Used to relate jobs and in logging *unused*
  Schedule  string // Traditional encoded string to represent the schedule
  Filters   []func(currentTime time.Time) (bool)
}

func (h *JobHandler) ParseJobConfig(confFile string) (error) {

  _, err := toml.DecodeFile(confFile, &h)
  if err != nil {
    return err
  }

  for jobIndex, _ := range h.Job {
    err = h.Job[jobIndex].ParseScheduleIntoFilters()
    if err != nil {
      return err
    }
  }
  return err
}

func (j *JobConfig) ParseScheduleIntoFilters() (error) {

  var err error
  scheduleChunks := strings.Split(j.Schedule, " ")
  if len(scheduleChunks) != 5 {
    return errors.New("Not enough elements in schedule for " + j.Title + ": " + j.Schedule)
  }

  // Add filter to only run on certain days of the week
  if scheduleChunks[DAYOFWEEK] != "*" {
    filterFunc, err := ParseDayOfWeekIntoFilter(scheduleChunks[DAYOFWEEK])
    if err != nil {
      return err
    }
    j.Filters = append(j.Filters, filterFunc)
  }

  return err
}

