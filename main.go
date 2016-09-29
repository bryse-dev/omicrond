package main

import (
	"flag"
	"github.com/Sirupsen/logrus"
	"github.com/brysearl/omicrond/conf"
  "github.com/brysearl/omicrond/job"
  "github.com/davecgh/go-spew/spew"
)

var logLevel = flag.Int("debug", 0, "Set the debug level: 1 = Panic, 2 = Fatal, 3 = Error, 4 = Warn, 5 = Info, 6 = Debug")

func init() {

  // Set the log level of the program
  conf.Attr.LogLevel = *logLevel
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
  handler.ParseJobConfig(conf.Attr.JobConfigPath)

  spew.Dump(handler)
}
