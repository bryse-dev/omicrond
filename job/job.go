package job

import (
  "errors"
  "strings"
  "time"
  "os/exec"
  "bytes"
  "github.com/BurntSushi/toml"
  "github.com/Sirupsen/logrus"
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

// JobHandler - Keep all the jobs together in an iterable slice
type JobHandler struct {
  Job []JobConfig
}

// JobConfig - Object representing a single scheduled job
type JobConfig struct {
  Title     string // The name of the job.  Used in logging
  Exec      CmdObj // String to be run on the system
  GroupName string // Used to relate jobs and in logging *unused*
  Schedule  string // Traditional encoded string to represent the schedule
  Filters   []func(currentTime time.Time) (bool)
}


// CheckIfScheduled - Initiates each filter for a job and returns whether or not to run the job
func (j *JobConfig) CheckIfScheduled(timeToCheck time.Time) (bool) {

  for _, filter := range j.Filters {
    result := filter(timeToCheck)
    if result == false {
      return false
    }
  }

  return true
}

// Run - Executes command
func (j *JobConfig) Run() error {

  var err error
  //stdOut, err := j.Exec.Cmd.StdoutPipe()
  // stdIn, err := j.Exec.Cmd.StdinPipe()
  // stdErr, err := j.Exec.Cmd.StderrPipe()
  var stdOut bytes.Buffer
	j.Exec.Cmd.Stdout = &stdOut

  logrus.Info("Running [" + j.Title + "]: " + j.Exec.Cmd.Path)
  err = j.Exec.Cmd.Start()

  //go io.Copy(os.Stdout, stdOut)
  j.Exec.Cmd.Wait()

  logrus.Info(stdOut.String())
  return err
}

// ParseJobConfig - Decode TOML config file and initiate ParseScheduleIntoFilters for each job
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

// CmdObj - Custom struct to enable a custom unmarshaler
type CmdObj struct {
  exec.Cmd
}

// UnmarshalText - Custom unmarshaler to build exec.Cmd type on JobConfig
func (c *CmdObj) UnmarshalText(text []byte) error {
  var err error
  cmdPtr := exec.Command(string(text))
  c.Cmd = *cmdPtr
  return err
}

// ParseScheduleIntoFilters - Translate schedule string into iterable functions
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
  // Add filter to limit to only certain months
  if scheduleChunks[MONTH] != "*" {
    filterFunc, err := ParseMonthIntoFilter(scheduleChunks[MONTH])
    if err != nil {
      return err
    }
    j.Filters = append(j.Filters, filterFunc)
  }
  // Add filter to limit to only certain days
  if scheduleChunks[DAY] != "*" {
    filterFunc, err := ParseDayOfMonthIntoFilter(scheduleChunks[DAY])
    if err != nil {
      return err
    }
    j.Filters = append(j.Filters, filterFunc)
  }
  // Add filter to limit to only certain hours
  if scheduleChunks[HOUR] != "*" {
    filterFunc, err := ParseHourIntoFilter(scheduleChunks[HOUR])
    if err != nil {
      return err
    }
    j.Filters = append(j.Filters, filterFunc)
  }
  // Add filter to limit to only certain minutes
  if scheduleChunks[MINUTE] != "*" {
    filterFunc, err := ParseMinuteIntoFilter(scheduleChunks[MINUTE])
    if err != nil {
      return err
    }
    j.Filters = append(j.Filters, filterFunc)
  }

  return err
}

