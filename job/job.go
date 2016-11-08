package job

import (
  "errors"
  "strings"
  "strconv"
  "time"
  "os"
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

// JobSchedule - Keep all the jobs together in an iterable slice
type JobSchedule struct {
  Job          []JobConfig
  LabelToIndex map[string]int
}

// JobConfig - Object representing a single scheduled job
type JobConfig struct {
  Label      string            // The name of the job.  Used in logging
  Command    string            // String to be run on the system
  GroupName  string            // Used to relate jobs and in logging *unused*
  Schedule   string            // Traditional encoded string to represent the schedule
  Locking    bool              // Self-locking daemon that won't step on its own toes
  Filters    []func(currentTime time.Time) (bool)
}

// JobScheduleAPI - Keep all the jobs together in an iterable slice and is JSON friendly for API use
type JobScheduleAPI struct {
  Job []JobConfigAPI
}

// JobConfigAPI - Object representing a single scheduled job and is JSON friendly for API use
type JobConfigAPI struct {
  Label      string            // The name of the job.  Used in logging
  Command    string            // String to be run on the system
  GroupName  string            // Used to relate jobs and in logging *unused*
  Locking    bool              // Self-locking daemon that won't step on its own toes
  Schedule   string            // Traditional encoded string to represent the schedule
}


///////////////// SETUP FUNCTIONS //////////////////////

// ParseJobConfig - Decode TOML config file and initiate ParseScheduleIntoFilters for each job
func (h *JobSchedule) ParseJobConfig(confFile string) (error) {

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
func (h *JobSchedule) WriteJobConfig(confFile string) (error) {

  var err error

  // If the file already exists, back it up
  backupFile := confFile + ".backup" + strconv.Itoa(int(time.Now().Unix()))
  if _, err := os.Stat(confFile); err == nil {
    err = os.Rename(confFile, backupFile)
  }

  writer, err := os.Create(confFile)
  schedule, err := h.MakeAPIFormat()
  if err := toml.NewEncoder(writer).Encode(schedule); err != nil {

    logrus.Error("Error encoding TOML: %s", err)
    err = os.Rename(backupFile, confFile)
  }

  return err
}

// CheckConfig - Sanity checks on the JobSchedule and builds LabelToIndex.
// MUST BE RUN EVERYTIME THE RUNNING CONFIG IS CHANGED!
func (h *JobSchedule) CheckConfig() error {

  var err error

  titleCheck := make(map[string]int)
  for jobIndex, _ := range h.Job {

    // Config sanity checks.  Make sure that labels exist and are unique.
    if h.Job[jobIndex].Label == "" {
      return errors.New("Config error: Job with an missing/empty label")
    }
    if strings.Contains(h.Job[jobIndex].Label, "_") {
      return errors.New("Config error: Job with reserved character '_' in label")
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

///////////////// API FUNCTIONS //////////////////////

// MakeAPIFormat - Convert internal object into external data
func (h *JobSchedule) MakeAPIFormat() (JobScheduleAPI, error) {

  var apiHandler JobScheduleAPI
  var err error
  for jobIndex, _ := range h.Job {

    // Convert the config to API format
    apiJobConfig, err := h.Job[jobIndex].MakeAPIFormat()
    if err != nil {
      return JobScheduleAPI{}, err
    }

    // Append the API job to the API schedule
    apiHandler.Job = append(apiHandler.Job, apiJobConfig)
    if err != nil {
      return JobScheduleAPI{}, err
    }
  }

  return apiHandler, err
}

// MakeAPIFormat - Convert internal object into external data
func (j *JobConfig) MakeAPIFormat() (JobConfigAPI, error) {

  var err error
  apiJobConfig := JobConfigAPI{
    Label: j.Label,
    Command: j.Command,
    GroupName: j.GroupName,
    Schedule: j.Schedule,
    Locking: j.Locking}

  return apiJobConfig, err
}

// GetJobByTitle - Find a job using its title
func (h *JobSchedule) GetJobByLabel(title string) (JobConfig, int, error) {
  var err error
  index, exists := h.LabelToIndex[title]
  if exists == true {
    return h.Job[index], index, err
  } else {
    spacesTitle := strings.Replace(title, "_", " ", -1)
    index, exists := h.LabelToIndex[spacesTitle]
    if exists == true {
      return h.Job[index], index, err
    } else {
      return JobConfig{}, -1, errors.New("Cannot find job with title: " + title)
    }
  }
}

// GetJobByID - Find a job using its ID
func (h *JobSchedule) GetJobByID(jobID int) (JobConfig, error) {
  var err error
  if jobID >= 0 && jobID <= len(h.Job) {
    return h.Job[jobID], err
  } else {
    return JobConfig{}, errors.New("Cannot find job with ID: " + strconv.Itoa(jobID))
  }
}
