package job

import (
  "errors"
  "strings"
  "time"
  "regexp"
  "strconv"
)

// ParseDayOfWeekIntoFilter - Translate schedule notation of a day of the week (0: Sun - 6: Sat)
//  into a function that when called with a datetime will, tell you if the date is on that day
func ParseDayOfWeekIntoFilter(rawStr string) (func(testTime time.Time) (bool), error) {

  // Run regex parsers against schedule string notation
  WeekDaysInt, _, err := parseScheduleStringToIntSlice(rawStr)
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

// ParseMonthIntoFilter - Translate schedule notation of a month (1: Jan - 12: Dec)
//  into a function that when called with a datetime, will tell you if the date is in that month
func ParseMonthIntoFilter(rawStr string) (func(testTime time.Time) (bool), error) {

  // Run regex parsers against schedule string notation
  scheduledMonthsInt, modulo, err := parseScheduleStringToIntSlice(rawStr)
  if err != nil {
    return nil, err
  }
  // Add the filter using the slice of ints
  filterFunc := func(testTime time.Time) (bool) {
    if intInSlice(int(testTime.Month()), scheduledMonthsInt) || (int(testTime.Month()) % modulo == 0 && modulo > 1) {
      return true
    }
    return false
  }

  return filterFunc, err
}

// ParseDayOfMonthIntoFilter - Translate schedule notation of a day in a month (1 - 32)
//  into a function that when called with a datetime, will tell you if the date is on that day of the month
func ParseDayOfMonthIntoFilter(rawStr string) (func(testTime time.Time) (bool), error) {

  // Run regex parsers against schedule string notation
  scheduledDaysOfMonthsInt, modulo, err := parseScheduleStringToIntSlice(rawStr)
  if err != nil {
    return nil, err
  }
  // Add the filter using the slice of ints
  filterFunc := func(testTime time.Time) (bool) {
    if intInSlice(testTime.Day(), scheduledDaysOfMonthsInt) || (testTime.Day() % modulo == 0 && modulo > 1) {
      return true
    }
    return false
  }

  return filterFunc, err
}

// ParseHourIntoFilter - Translate schedule notation of an hour in a day (0 - 23)
//  into a function that when called with a datetime, will tell you if the time is currently in that hour
func ParseHourIntoFilter(rawStr string) (func(testTime time.Time) (bool), error) {

  // Run regex parsers against schedule string notation
  scheduledHoursInt, modulo, err := parseScheduleStringToIntSlice(rawStr)
  if err != nil {
    return nil, err
  }
  // Add the filter using the slice of ints
  filterFunc := func(testTime time.Time) (bool) {
    if intInSlice(testTime.Hour(), scheduledHoursInt) || (testTime.Hour() % modulo == 0 && modulo > 1) {
      return true
    }
    return false
  }

  return filterFunc, err
}

// ParseMinuteIntoFilter - Translate schedule notation of a minute in an hour (0 - 59)
//  into a function that when called with a datetime, will tell you if the time is currently in that minute
func ParseMinuteIntoFilter(rawStr string) (func(testTime time.Time) (bool), error) {

  // Run regex parsers against schedule string notation
  scheduledMinutesInt, modulo, err := parseScheduleStringToIntSlice(rawStr)
  if err != nil {
    return nil, err
  }
  // Add the filter using the slice of ints
  filterFunc := func(testTime time.Time) (bool) {
    if intInSlice(testTime.Minute(), scheduledMinutesInt) || (testTime.Minute() % modulo == 0 && modulo > 1) {
      return true
    }
    return false
  }

  return filterFunc, err
}

// parseScheduleStringToIntSlice - Parse a schedule notation string into an integer array representing points
//  of time within that scope (such as 30 being the 30th minute in an hour) and an integer modulo (ex 5) which represents
//  the shorthand schedule notation for an interval (ex '*/5')
func parseScheduleStringToIntSlice(rawStr string) ([]int, int, error) {

  var elementStrNumSlice []string // Slice of stringified integers that need to be converted into ints
  var elementIntSlice []int // Slice of integers after being Atoi-ed from elementStrNumSlice
  var err error
  modulo := 1 // Interval operand.  Default 1 so that it will always equal true if checked

  // Set the modulo if the shorthand interval notation is used
  if proceed, _ := regexp.MatchString("^" + regexp.QuoteMeta("*/") + "[0-9]+$", rawStr); proceed == true {
    re := regexp.MustCompile("^" + regexp.QuoteMeta("*/") + "([0-9]+)$")
    matches := re.FindStringSubmatch(rawStr)
    modulo, err = strconv.Atoi(matches[1])

    if err != nil {
      return nil, modulo, err
    }
    // Create a (single element) slice of a single implicit time scope
  } else if proceed, _ := regexp.MatchString("^[0-9]+$", rawStr); proceed == true {

    intValue, err := strconv.Atoi(rawStr)
    if err != nil {
      return nil, modulo, err
    }
    elementIntSlice = append(elementIntSlice, intValue)

    // Create a slice of each time scope that is comma seperated
  } else if proceed, _ := regexp.MatchString("^[0-9]+,.*$", rawStr); proceed == true {
    elementStrNumSlice = strings.Split(rawStr, ",")
    for _, strValue := range elementStrNumSlice {
      intValue, err := strconv.Atoi(strValue)
      if err != nil {
        return nil, modulo, err
      }
      elementIntSlice = append(elementIntSlice, intValue)
    }

    // Create a slice of each time scope that is range implied
  } else if proceed, _ := regexp.MatchString("[0-9]+-[0-9]+", rawStr); proceed == true {
    elementStrNumSlice = strings.Split(rawStr, "-")
    startDay, err := strconv.Atoi(elementStrNumSlice[0])
    if err != nil {
      return nil, modulo, err
    }
    endDay, err := strconv.Atoi(elementStrNumSlice[1])
    if err != nil {
      return nil, modulo, err
    }
    if (startDay < endDay) {
      for i := startDay; i <= endDay; i++ {
        elementIntSlice = append(elementIntSlice, i)
      }
    } else {
      return nil, modulo, errors.New("Element range implication is not smaller to larger.  ")
    }
    // Something is wrong with the string to parse, return error
  } else {
    return nil, modulo, errors.New("Could not parse element string '" + rawStr + "'")
  }

  return elementIntSlice, modulo, err
}

// stringInSlice - Used by filters to determine if an element of time is scheduled
func stringInSlice(a string, list []string) (bool) {
  for _, b := range list {
    if b == a {
      return true
    }
  }
  return false
}

// intInSlice - Used by filters to determine if an element of time is scheduled
func intInSlice(a int, list []int) (bool) {
  for _, b := range list {
    if b == a {
      return true
    }
  }
  return false
}

// intToDayOfWeek - Used to convert integer representations of a day of the week into the English standard name
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
