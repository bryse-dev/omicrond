package job

import (
  "errors"
  "strings"
  "strconv"
  "time"
  "os"
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

// JobHandler - Keep all the jobs together in an iterable slice
type JobHandler struct {
  Job          []JobConfig
  LabelToIndex map[string]int
}

// JobConfig - Object representing a single scheduled job
type JobConfig struct {
  Label     string // The name of the job.  Used in logging
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
  Label     string // The name of the job.  Used in logging
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

  err = h.CheckConfig()
  if err != nil {
    logrus.Fatal(err)
  }

  for jobIndex, _ := range h.Job {

    err = h.Job[jobIndex].ParseScheduleIntoFilters(false)
    if err != nil {
      return err
    }
  }

  return err
}

// WriteJobConfig - update the written config file with any changes that have occured
func (h *JobHandler) WriteJobConfig(confFile string) (error) {

  var err error

  // If the file already exists, back it up
  backupFile := confFile + ".backup" + strconv.Itoa(int(time.Now().Unix()))
  if _, err := os.Stat(confFile); err == nil {
    err = os.Rename(confFile, backupFile)
  }

  writer, err := os.Create(confFile)
  handler, err := h.MakeAPIFormat()
  if err := toml.NewEncoder(writer).Encode(handler); err != nil {

    logrus.Error("Error encoding TOML: %s", err)
    err = os.Rename(backupFile, confFile)
  }

  return err
}

// CheckConfig - Sanity checks on the JobHandler and builds LabelToIndex.
// MUST BE RUN EVERYTIME THE RUNNING CONFIG IS CHANGED!
func (h *JobHandler) CheckConfig() error {

  var err error

  titleCheck := make(map[string]int)
  for jobIndex, _ := range h.Job {

    // Config sanity checks.  Make sure that labels exist and are unique.
    if h.Job[jobIndex].Label == "" {
      return errors.New("Config error: Job with an missing/empty label")
    }
    _, exists := titleCheck[h.Job[jobIndex].Label]
    if exists == true {
      return errors.New("Config error: Jobs with duplicate labels.")
    }
    titleCheck[h.Job[jobIndex].Label] = jobIndex

    err = h.Job[jobIndex].ParseScheduleIntoFilters(true)
    if err != nil {
      return err
    }
  }
  h.LabelToIndex = titleCheck

  return err
}

// ParseScheduleIntoFilters - Translate schedule string into iterable functions
func (j *JobConfig) ParseScheduleIntoFilters(testing bool) (error) {

  var err error
  scheduleChunks := strings.Split(j.Schedule, " ")
  if len(scheduleChunks) != 5 {
    return errors.New("Cannot parse schedule string " + j.Label + ": " + j.Schedule)
  }

  // Add filter to only run on certain days of the week
  if scheduleChunks[DAYOFWEEK] != "*" {
    filterFunc, err := ParseDayOfWeekIntoFilter(scheduleChunks[DAYOFWEEK])
    if err != nil {
      return err
    }
    if testing == false {
      j.Filters = append(j.Filters, filterFunc)
    }
  }
  // Add filter to limit to only certain months
  if scheduleChunks[MONTH] != "*" {
    filterFunc, err := ParseMonthIntoFilter(scheduleChunks[MONTH])
    if err != nil {
      return err
    }
    if testing == false {
      j.Filters = append(j.Filters, filterFunc)
    }
  }
  // Add filter to limit to only certain days
  if scheduleChunks[DAY] != "*" {
    filterFunc, err := ParseDayOfMonthIntoFilter(scheduleChunks[DAY])
    if err != nil {
      return err
    }
    if testing == false {
      j.Filters = append(j.Filters, filterFunc)
    }
  }
  // Add filter to limit to only certain hours
  if scheduleChunks[HOUR] != "*" {
    filterFunc, err := ParseHourIntoFilter(scheduleChunks[HOUR])
    if err != nil {
      return err
    }
    if testing == false {
      j.Filters = append(j.Filters, filterFunc)
    }
  }
  // Add filter to limit to only certain minutes
  if scheduleChunks[MINUTE] != "*" {
    filterFunc, err := ParseMinuteIntoFilter(scheduleChunks[MINUTE])
    if err != nil {
      return err
    }
    if testing == false {
      j.Filters = append(j.Filters, filterFunc)
    }
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
  logrus.Info("Running [" + j.Label + "]: " + strings.Join(command.Args, " "))
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

///////////////// API FUNCTIONS //////////////////////

// MakeAPIFormat - Remove internal data structures from JobHandler
func (h *JobHandler) MakeAPIFormat() (JobHandlerAPI, error) {

  var apiHandler JobHandlerAPI
  var err error
  for jobIndex, _ := range h.Job {

    // Convert the config to API format
    apiJobConfig, err := h.Job[jobIndex].MakeAPIFormat(*h)
    if err != nil {
      return JobHandlerAPI{}, err
    }

    // Append the API job to the API schedule
    apiHandler.Job = append(apiHandler.Job, apiJobConfig)
    if err != nil {
      return JobHandlerAPI{}, err
    }
  }

  return apiHandler, err
}

// MakeAPIFormat - Remove internal data structures from JobHandler
func (j *JobConfig) MakeAPIFormat(parentHandler JobHandler) (JobConfigAPI, error) {

  var err error
  _, myID, err := parentHandler.GetJobByLabel(j.Label)
  if err != nil {
    return JobConfigAPI{}, errors.New("Cannot find job in the passed schedule")
  }
  apiJobConfig := JobConfigAPI{
    ID: myID,
    Label: j.Label,
    Command: j.Command,
    GroupName: j.GroupName,
    Schedule: j.Schedule }

  return apiJobConfig, err
}

// GetJobByTitle - Find a job using its title
func (h *JobHandler) GetJobByLabel(title string) (JobConfig, int, error) {
  var err error
  index, exists := h.LabelToIndex[title]
  if exists == true {
    return h.Job[index], index, err
  } else {
    return JobConfig{}, -1, errors.New("Cannot find job with title: " + title)
  }
}

// GetJobByID - Find a job using its ID
func (h *JobHandler) GetJobByID(jobID int) (JobConfig, error) {
  var err error
  if jobID >= 0 && jobID <= len(h.Job) {
    return h.Job[jobID], err
  } else {
    return JobConfig{}, errors.New("Cannot find job with ID: " + strconv.Itoa(jobID))
  }
}
