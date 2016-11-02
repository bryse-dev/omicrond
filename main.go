package main

import (
  "flag"
  "os"
  "github.com/Sirupsen/logrus"
  "github.com/brysearl/omicrond/conf"
  "github.com/brysearl/omicrond/daemon"
)
//"github.com/davecgh/go-spew/spew"


func init() {

  // Configure command line arguments
  var logLevelPtr = flag.Int("v", conf.Attr.LogLevel, "Set the debug level: 1 = Panic, 2 = Fatal, 3 = Error, 4 = Warn, 5 = Info, 6 = Debug")
  var jobConfigPathPtr = flag.String("config", conf.Attr.JobConfigPath, "Path to the daemon configuration file")
  var apiAddressPtr = flag.String("api_address", conf.Attr.APIAddress, "IP to run the API service")
  var apiPortPtr = flag.Int("api_port", conf.Attr.APIPort, "Port to run the API service")
  var apiTimeoutPtr = flag.Int("api_timeout", conf.Attr.APITimeout, "API service request timeout in seconds")

  // Retrieve command line arguments
  flag.Parse()

  // Set the path to the daemon config file
  conf.Attr.JobConfigPath = *jobConfigPathPtr

  // Set the log level of the program
  conf.Attr.LogLevel = *logLevelPtr

  // Set the address of the api service
  conf.Attr.APIAddress = *apiAddressPtr

  // Set the port of the api service
  conf.Attr.APIPort = *apiPortPtr

  // Set the port of the api service
  conf.Attr.APITimeout = *apiTimeoutPtr

  // Create directories if they don't exist
  if err := os.MkdirAll(conf.Attr.BaseDir,0755); err != nil {
    logrus.Fatal(err)
  }
  if err := os.MkdirAll(conf.Attr.LoggingPath,0755); err != nil {
    logrus.Fatal(err)
  }

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

  // Output with absolute time
  customFormatter := new(logrus.TextFormatter)
  customFormatter.TimestampFormat = "2006-01-02 15:04:05"
  logrus.SetFormatter(customFormatter)
  customFormatter.FullTimestamp = true
}

func main() {

  logrus.Info("Starting")
  daemon.StartDaemon()
}

