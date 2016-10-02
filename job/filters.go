package job

import (
  "errors"
  "strings"
  "time"
  "regexp"
  "strconv"
)

func ParseDayOfWeekIntoFilter(rawStr string) (func(testTime time.Time) (bool), error) {

  WeekDaysInt, err := parseScheduleStringToIntSlice(rawStr)
  if err != nil {
    return nil, err
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
  filterFunc := func(testTime time.Time) (bool) {
    if stringInSlice(testTime.Weekday().String(), scheduledWeekDays) {
      return true
    }
    return false
  }

  return filterFunc, err
}

func ParseMonthIntoFilter(rawStr string) (func(testTime time.Time) (bool), error) {

  scheduledMonthsInt, err := parseScheduleStringToIntSlice(rawStr)
  if err != nil {
    return nil, err
  }
  // Add the filter using the slice of ints
  filterFunc := func(testTime time.Time) (bool) {
    if intInSlice(int(testTime.Month()), scheduledMonthsInt) {
      return true
    }
    return false
  }

  return filterFunc, err
}

func ParseDayOfMonthIntoFilter(rawStr string) (func(testTime time.Time) (bool), error) {

  scheduledDaysOfMonthsInt, err := parseScheduleStringToIntSlice(rawStr)
  if err != nil {
    return nil, err
  }
  // Add the filter using the slice of ints
  filterFunc := func(testTime time.Time) (bool) {
    if intInSlice(testTime.Day(), scheduledDaysOfMonthsInt) {
      return true
    }
    return false
  }

  return filterFunc, err
}

func ParseHourIntoFilter(rawStr string) (func(testTime time.Time) (bool), error) {

  scheduledHoursInt, err := parseScheduleStringToIntSlice(rawStr)
  if err != nil {
    return nil, err
  }
  // Add the filter using the slice of ints
  filterFunc := func(testTime time.Time) (bool) {
    if intInSlice(testTime.Hour(), scheduledHoursInt) {
      return true
    }
    return false
  }

  return filterFunc, err
}

func ParseMinuteIntoFilter(rawStr string) (func(testTime time.Time) (bool), error) {

  scheduledMinutesInt, err := parseScheduleStringToIntSlice(rawStr)
  if err != nil {
    return nil, err
  }
  // Add the filter using the slice of ints
  filterFunc := func(testTime time.Time) (bool) {
    if intInSlice(testTime.Minute(), scheduledMinutesInt) {
      return true
    }
    return false
  }

  return filterFunc, err
}

func parseScheduleStringToIntSlice(rawStr string) ([]int, error) {

  var elementStrNumSlice []string
  var elementIntSlice []int
  var err error

  if proceed, _ := regexp.MatchString("^[0-9]+$", rawStr); proceed == true {

    intValue, err := strconv.Atoi(rawStr)
    if err != nil {
      return nil, err
    }
    elementIntSlice = append(elementIntSlice, intValue)

    // Create a slice of each day that is comma seperated
  } else if proceed, _ := regexp.MatchString("^[0-9]+,.*$", rawStr); proceed == true {
    elementStrNumSlice = strings.Split(rawStr, ",")
    for _, strValue := range elementStrNumSlice {
      intValue, err := strconv.Atoi(strValue)
      if err != nil {
        return nil, err
      }
      elementIntSlice = append(elementIntSlice, intValue)
    }

    // Create a slice of each day that is range implied
  } else if proceed, _ := regexp.MatchString("[0-9]+-[0-9]+", rawStr); proceed == true {
    elementStrNumSlice = strings.Split(rawStr, "-")
    startDay, err := strconv.Atoi(elementStrNumSlice[0])
    if err != nil {
      return nil, err
    }
    endDay, err := strconv.Atoi(elementStrNumSlice[1])
    if err != nil {
      return nil, err
    }
    if (startDay < endDay) {
      for i := startDay; i <= endDay; i++ {
        elementIntSlice = append(elementIntSlice, i)
      }
    } else {
      return nil, errors.New("Element range implication is not smaller to larger.  ")
    }
    // Something is wrong with the string to parse, return error
  } else {
    return nil, errors.New("Could not parse element string '" + rawStr + "'")
  }

  return elementIntSlice, err
}

func stringInSlice(a string, list []string) (bool) {
  for _, b := range list {
    if b == a {
      return true
    }
  }
  return false
}

func intInSlice(a int, list []int) (bool) {
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
