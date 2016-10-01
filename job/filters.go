package job

import (
  "errors"
  "strings"
  "time"
  "regexp"
  "strconv"
)

func ParseDayOfWeekIntoFilter (rawStr string) (func(currentTime time.Time) (bool), error){

  var WeekDaysStrNum []string
  var WeekDaysInt []int
  var err error

  if proceed, _ := regexp.MatchString("^[0-6]$", rawStr); proceed == true {

    intValue, err := strconv.Atoi(rawStr)
    if err != nil {
      return nil,err
    }
    WeekDaysInt = append(WeekDaysInt, intValue)

    // Create a slice of each day that is comma seperated
  } else if proceed, _ := regexp.MatchString("^[0-6]+,.*$", rawStr); proceed == true {
    WeekDaysStrNum = strings.Split(rawStr, ",")
    for _, strValue := range WeekDaysStrNum {
      intValue, err := strconv.Atoi(strValue)
      if err != nil {
        return nil, err
      }
      WeekDaysInt = append(WeekDaysInt, intValue)
    }

    // Create a slice of each day that is range implied
  } else if proceed, _ := regexp.MatchString("[0-6]+-[0-6]+", rawStr); proceed == true {
    WeekDaysStrNum = strings.Split(rawStr, "-")
    startDay, err := strconv.Atoi(WeekDaysStrNum[0])
    if err != nil {
      return nil, err
    }
    endDay, err := strconv.Atoi(WeekDaysStrNum[1])
    if err != nil {
      return nil, err
    }
    if (startDay < endDay) {
      for i := startDay; i <= endDay; i++ {
        WeekDaysInt = append(WeekDaysInt, i)
      }
    } else {
      return nil, errors.New("DAYOFWEEK range implication is not smaller to larger")
    }
    // Something is wrong with the string to parse, return error
  } else {
    return nil, errors.New("Could not parse DAYOFWEEK string '" + rawStr + "'")
  }

  // Convert the slice of numbered weekdays to their proper names
  var scheduledWeekDays []string
  for _, intValue := range WeekDaysInt {
    dayName, err := intToDayOfWeek(intValue)
    if err != nil {
      return nil, err
    }
    scheduledWeekDays = append(scheduledWeekDays, dayName)
  }

  // Add the filter using the slice of allowed weekdays
  filterFunc := func(currentTime time.Time) (bool) {
    if stringInSlice(string(currentTime.Weekday()), scheduledWeekDays) {
      return true
    }
    return false
  }

  return filterFunc, err
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
