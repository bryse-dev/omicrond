package job

import (
  "errors"
  "strings"
  "time"
  "regexp"
  "strconv"
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

    var WeekDaysStrNum []string
    var WeekDaysInt []int

    if proceed, _ := regexp.MatchString("^[0-6]$", scheduleChunks[DAYOFWEEK]); proceed == true {

      intValue, err := strconv.Atoi(scheduleChunks[DAYOFWEEK])
      if err != nil {
        return err
      }
      WeekDaysInt = append(WeekDaysInt, intValue)

      // Create a slice of each day that is comma seperated
    } else if proceed, _ := regexp.MatchString("^[0-6]+,.*$", scheduleChunks[DAYOFWEEK]); proceed == true {
      WeekDaysStrNum = strings.Split(scheduleChunks[DAYOFWEEK], ",")
      for _, strValue := range WeekDaysStrNum {
        intValue, err := strconv.Atoi(strValue)
        if err != nil {
          return err
        }
        WeekDaysInt = append(WeekDaysInt, intValue)
      }

      // Create a slice of each day that is range implied
    } else if proceed, _ := regexp.MatchString("[0-6]+-[0-6]+", scheduleChunks[DAYOFWEEK]); proceed == true {
      WeekDaysStrNum = strings.Split(scheduleChunks[DAYOFWEEK], "-")
      startDay, err := strconv.Atoi(WeekDaysStrNum[0])
      if err != nil {
        return err
      }
      endDay, err := strconv.Atoi(WeekDaysStrNum[1])
      if err != nil {
        return err
      }
      if (startDay < endDay) {
        for i := startDay; i <= endDay; i++ {
          WeekDaysInt = append(WeekDaysInt, i)
        }
      } else {
        return errors.New("DAYOFWEEK range implication is not smaller to larger")
      }
      // Something is wrong with the string to parse, return error
    } else {
      return errors.New("Could not parse DAYOFWEEK string for '" + j.Title + "': " + scheduleChunks[DAYOFWEEK])
    }

    // Convert the slice of numbered weekdays to their proper names
    var scheduledWeekDays []string
    for _, intValue := range WeekDaysInt {
      dayName, err := intToDayOfWeek(intValue)
      if err != nil {
        return err
      }
      scheduledWeekDays = append(scheduledWeekDays, dayName)
    }

    // Add the filter using the slice of allowed weekdays
    j.Filters = append(j.Filters, func(currentTime time.Time) (bool) {
      if stringInSlice(string(currentTime.Weekday()), scheduledWeekDays) {
        return true
      }
      return false
    })
  }

  for i := len(scheduleChunks) - 1; i >= 0; i-- {

  }

  return err
}

func stringInSlice(a string, list []string) (bool) {
  for _, b := range list {
    if b == a {
      return true
    }
  }
  return false
}

func intToDayOfWeek(intDay int) (string, error) {
  var err error
  switch (intDay){
  case 0:
    return SUNDAY, err
  case 1:
    return MONDAY, err
  case 2:
    return TUESDAY, err
  case 3:
    return WEDNESDAY, err
  case 4:
    return THURSDAY, err
  case 5:
    return FRIDAY, err
  case 6:
    return SATURDAY, err
  default:
    return "", errors.New("Cannot convert (" + string(intDay) + ") into a weekday")
  }
}