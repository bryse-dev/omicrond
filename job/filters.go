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
  WeekDaysIntMap, _, err := parseScheduleStringToIntMap(rawStr)
  if err != nil {
    return nil, err
  }

  // Convert the slice of numbered weekdays to their proper names
  scheduledWeekDaysStringMap := make(map[string]bool)
  for intValue, _ := range WeekDaysIntMap {
    dayName, err := intToDayOfWeek(intValue)
    if err != nil {
      return nil, err
    }
    scheduledWeekDaysStringMap[dayName] = true
  }

  // Add the filter using the slice of allowed weekdays
  filterFunc := func(testTime time.Time) (bool) {
    _, isScheduled := scheduledWeekDaysStringMap[testTime.Weekday().String()]
    if isScheduled == true {
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
  scheduledMonthsIntMap, intervalModulo, err := parseScheduleStringToIntMap(rawStr)
  if err != nil {
    return nil, err
  }
  // Add the filter using the slice of ints
  filterFunc := func(testTime time.Time) (bool) {
    _, isScheduled := scheduledMonthsIntMap[int(testTime.Month())]
    if isScheduled == true || (int(testTime.Month()) % intervalModulo == 0 && intervalModulo > 1) {
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
  scheduledDaysOfMonthsIntMap, intervalModulo, err := parseScheduleStringToIntMap(rawStr)
  if err != nil {
    return nil, err
  }
  // Add the filter using the slice of ints
  filterFunc := func(testTime time.Time) (bool) {
    _, isScheduled := scheduledDaysOfMonthsIntMap[testTime.Day()]
    if isScheduled == true || (testTime.Day() % intervalModulo == 0 && intervalModulo > 1) {
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
  scheduledHoursIntMap, intervalModulo, err := parseScheduleStringToIntMap(rawStr)
  if err != nil {
    return nil, err
  }
  // Add the filter using the slice of ints
  filterFunc := func(testTime time.Time) (bool) {
    _, isScheduled := scheduledHoursIntMap[testTime.Hour()]
    if isScheduled == true || (testTime.Hour() % intervalModulo == 0 && intervalModulo > 1) {
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
  scheduledMinutesIntMap, intervalModulo, err := parseScheduleStringToIntMap(rawStr)
  if err != nil {
    return nil, err
  }
  // Add the filter using the slice of ints
  filterFunc := func(testTime time.Time) (bool) {
    _, isScheduled := scheduledMinutesIntMap[testTime.Minute()]
    if isScheduled == true || (testTime.Minute() % intervalModulo == 0 && intervalModulo > 1) {
      return true
    }
    return false
  }

  return filterFunc, err
}

// parseScheduleStringToIntMap - Parse a schedule notation string into an integer map representing points
//  of time within that scope (such as 30 being the 30th minute in an hour) and an integer modulo (ex 5) which represents
//  the shorthand schedule notation for an interval (ex '*/5')
func parseScheduleStringToIntMap(rawStr string) (map[int]bool, int, error) {

  var err error
  var elementStrNumSlice []string // Slice of stringified integers that need to be converted into ints
  elementIntMap := make(map[int]bool) // Map of integers after being Atoi-ed from elementStrNumSlice
  intervalModulo := 1 // Interval operand.  Default 1 so that it will always equal true if checked

  elementStrNumSlice = strings.Split(rawStr, ",")
  for _, elementStrNum := range elementStrNumSlice {

    if proceed, _ := regexp.MatchString("^" + regexp.QuoteMeta("*/") + "[0-9]+$", elementStrNum); proceed == true {
      re := regexp.MustCompile("^" + regexp.QuoteMeta("*/") + "([0-9]+)$")
      matches := re.FindStringSubmatch(elementStrNum)
      intervalModulo, err = strconv.Atoi(matches[1])

      if err != nil {
        return nil, intervalModulo, err
      }
      // Create a (single element) slice of a single implicit time scope
    } else if proceed, _ := regexp.MatchString("^[0-9]+$", elementStrNum); proceed == true {

      intValue, err := strconv.Atoi(elementStrNum)
      if err != nil {
        return nil, intervalModulo, err
      }
      elementIntMap[intValue] = true
    } else if proceed, _ := regexp.MatchString("[0-9]+-[0-9]+", elementStrNum); proceed == true {
      elementStrNumSlice = strings.Split(elementStrNum, "-")
      startDay, err := strconv.Atoi(elementStrNumSlice[0])
      if err != nil {
        return nil, intervalModulo, err
      }
      endDay, err := strconv.Atoi(elementStrNumSlice[1])
      if err != nil {
        return nil, intervalModulo, err
      }
      if (startDay < endDay) {
        for i := startDay; i <= endDay; i++ {
          elementIntMap[i] = true
        }
      } else {
        return nil, intervalModulo, errors.New("Element range implication is not smaller to larger.  ")
      }
    } else if proceed, _ := regexp.MatchString("^" + regexp.QuoteMeta("*/") + "[0-9]+$", elementStrNum); proceed == true {
      re := regexp.MustCompile("^" + regexp.QuoteMeta("*/") + "([0-9]+)$")
      matches := re.FindStringSubmatch(elementStrNum)
      intervalModulo, err = strconv.Atoi(matches[1])

      if err != nil {
        return nil, intervalModulo, err
      }
    } else {
      return nil, intervalModulo, errors.New("Could not parse [" + elementStrNum + "] of " + rawStr)
    }
  }

  return elementIntMap, intervalModulo, err
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
