package job

import (
  "errors"
  "strings"
  "time"
  "os/exec"
  "bufio"
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

var RunningSchedule *JobHandler

// JobHandler - Keep all the jobs together in an iterable slice
type JobHandler struct {
  Job []JobConfig
}

// JobConfig - Object representing a single scheduled job
type JobConfig struct {
  Title     string // The name of the job.  Used in logging
  Command   string // String to be run on the system
  GroupName string // Used to relate jobs and in logging *unused*
  Schedule  string // Traditional encoded string to represent the schedule
  Filters   []func(currentTime time.Time) (bool)
}

// JobHandlerAPI - Keep all the jobs together in an iterable slice and is JSON friendly for API use
type JobHandlerAPI struct {
  Job []JobConfigAPI
}

// JobConfigAPI - Object representing a single scheduled job and is JSON friendly for API use
type JobConfigAPI struct {
  ID        int    // Index of the job in the JobHandler
  Title     string // The name of the job.  Used in logging
  Command   string // String to be run on the system
  GroupName string // Used to relate jobs and in logging *unused*
  Schedule  string // Traditional encoded string to represent the schedule
}


///////////////// SETUP FUNCTIONS //////////////////////

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

  RunningSchedule = h

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

///////////////// SCHEDULING FUNCTIONS //////////////////////

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
func (j *JobConfig) Run() {

  var err error


  // Build the system level command from the configured command string
  command := j.buildCommand()
  if err != nil {
    logrus.Error(err)
  }

  // Create handles for both stdin and stdout
  stdOut, err := command.StdoutPipe()
  if err != nil {
    logrus.Error(err)
    return
  }
  stdErr, err := command.StderrPipe()
  if err != nil {
    logrus.Error(err)
    return
  }

  // Attach scanners to the IO handles
  stdOutScanner := bufio.NewScanner(stdOut)
  stdErrScanner := bufio.NewScanner(stdErr)

  // Spawn goroutines to effectively tail the IO scanners
  go func() {
    for stdOutScanner.Scan() {
      logrus.Debug("STDOUT | " + stdOutScanner.Text())
    }
  }()

  go func() {
    for stdErrScanner.Scan() {
      logrus.Debug("STDERR | " + stdErrScanner.Text())
    }
  }()

  // Start the command
  logrus.Info("Running [" + j.Title + "]: " + strings.Join(command.Args, " "))
  err = command.Start()
  if err != nil {
    logrus.Error(err)
    return
  }

  // Wait for the command to complete
  logrus.Debug("Waiting for command to complete")
  command.Wait()
  logrus.Debug("Command completed")

  return
}

// buildCommand - Convert string to executablte exec.Cmd type
func (j *JobConfig) buildCommand() *exec.Cmd {

  // Split on spaces
  components := strings.Split(string(j.Command), " ")
  if len(components) == 0 {
    logrus.Error("Missing exec command in job configuration")
  }

  // Shift off the executable from the arguments
  executable, components := components[0], components[1:]

  // Create the exec.Cmd object and attach to JobConfig
  cmdPtr := exec.Command(executable, components...)
  return cmdPtr
}

func (h *JobHandler) MakeAPIFormat() JobHandlerAPI {

  var apiHandler JobHandlerAPI
  for jobIndex, _ := range h.Job {

    apiHandler.Job = append(apiHandler.Job, JobConfigAPI{
      ID: jobIndex,
      Title: h.Job[jobIndex].Title,
      Command: h.Job[jobIndex].Command,
      GroupName: h.Job[jobIndex].GroupName,
      Schedule: h.Job[jobIndex].Schedule })
  }

  return apiHandler
}
