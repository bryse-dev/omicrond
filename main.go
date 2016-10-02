package main

import (
  "time"
  "flag"
  "github.com/Sirupsen/logrus"
  "github.com/brysearl/omicrond/conf"
  "github.com/brysearl/omicrond/job"
  "github.com/davecgh/go-spew/spew"
)

func init() {

  // Configure command line arguments
  var logLevelPtr = flag.Int("v", conf.Attr.LogLevel, "Set the debug level: 1 = Panic, 2 = Fatal, 3 = Error, 4 = Warn, 5 = Info, 6 = Debug")
  var jobConfigPathPtr = flag.String("config", conf.Attr.JobConfigPath, "Path to the daemon configuration file")

  // Retrieve command line arguments
  flag.Parse()

  // Set the path to the daemon config file
  conf.Attr.JobConfigPath = *jobConfigPathPtr

  // Set the log level of the program
  conf.Attr.LogLevel = *logLevelPtr

  switch {
  case conf.Attr.LogLevel == 1:
    logrus.SetLevel(logrus.PanicLevel)
  case conf.Attr.LogLevel == 2:
    logrus.SetLevel(logrus.FatalLevel)
  case conf.Attr.LogLevel == 3:
    logrus.SetLevel(logrus.ErrorLevel)
  case conf.Attr.LogLevel == 4:
    logrus.SetLevel(logrus.WarnLevel)
  case conf.Attr.LogLevel == 5:
    logrus.SetLevel(logrus.InfoLevel)
  case conf.Attr.LogLevel == 6:
    logrus.SetLevel(logrus.DebugLevel)
  default:
    logrus.SetLevel(logrus.InfoLevel)
  }
}

func main() {

  logrus.Info("Starting")

  logrus.Info("Reading job configuration file: " + conf.Attr.JobConfigPath)
  var handler = job.JobHandler{}
  err := handler.ParseJobConfig(conf.Attr.JobConfigPath)
  if err != nil {
    logrus.Fatal(err)
  }

  logrus.Debug(spew.Sdump(handler))

  testTime, _ := time.Parse(time.RFC850, "Monday, 02-Jan-06 15:04:05 EST")
  for _, j := range handler.Job {
    for _, f := range j.Filters {
      test := f(testTime)
      if test == true {
        logrus.Debug("test passed")
      } else {
        logrus.Debug("test failed")
      }
    }
  }
}
