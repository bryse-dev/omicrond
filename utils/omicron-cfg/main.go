package main

import (
  "flag"
  "os"
  "regexp"
  "strings"
  "bufio"
  "strconv"
  "github.com/Sirupsen/logrus"
  "github.com/BurntSushi/toml"
  "github.com/brysearl/omicrond/job"
)

// Configure command line arguments
var logLevelPtr = flag.Int("v", 2, "Set the debug level: 1 = Panic, 2 = Fatal, 3 = Error, 4 = Warn, 5 = Info, 6 = Debug")
var configFilePtr = flag.String("config", "", "Path to the legacy configuration file")
var convertedConfigFilePtr = flag.String("outfile", "", "Path to the new configuration file")
var configFile string
var fileOut *os.File

func init() {

  // Retrieve command line arguments
  flag.Parse()

  // Set the path to the daemon config file
  configFile = *configFilePtr

  if *convertedConfigFilePtr == "" {
    // Set the default file to be stdout
    fileOut = os.Stdout
  } else {
    var err error
    fileOut, err = os.Create(*convertedConfigFilePtr)
    if err != nil {
      logrus.Fatal("Cannot create File: " + *convertedConfigFilePtr)
    }
  }

  // Set the log level of the program
  switch {
  case *logLevelPtr == 1:
    logrus.SetLevel(logrus.PanicLevel)
  case *logLevelPtr == 2:
    logrus.SetLevel(logrus.FatalLevel)
  case *logLevelPtr == 3:
    logrus.SetLevel(logrus.ErrorLevel)
  case *logLevelPtr == 4:
    logrus.SetLevel(logrus.WarnLevel)
  case *logLevelPtr == 5:
    logrus.SetLevel(logrus.InfoLevel)
  case *logLevelPtr == 6:
    logrus.SetLevel(logrus.DebugLevel)
  default:
    logrus.SetLevel(logrus.InfoLevel)
  }
}

func main() {

  // Make sure that the passed file exists
  if _, err := os.Stat(configFile); os.IsNotExist(err) {
    logrus.Fatal("File does not exists: " + configFile)
  }

  // Read in the file and parse into a job.JobHandler object
  jobHandler := readConfigFile(configFile)

  // Write the object to a TOML file
  writeConfigFile(jobHandler, fileOut)
}

// Write the TOML config to the passed in file
func writeConfigFile(jobHandler job.JobHandler, writer *os.File) {

  if err := toml.NewEncoder(writer).Encode(jobHandler); err != nil {
    logrus.Fatal("Error encoding TOML: %s", err)
  }

  return
}

// Read legacy config and parse out each Job
func readConfigFile(configFile string) (job.JobHandler) {

  var handler job.JobHandler

  // Open the file for reading and defer closing
  file, err := os.Open(configFile)
  if err != nil {
    logrus.Fatal(err)
  }
  defer file.Close()

  // Setup scanner and iteration variables
  scanner := bufio.NewScanner(file)
  title := "title"
  titleCounter := 1

  // Read in file line by line and build JobConfig objects
  for scanner.Scan() {
    line := scanner.Text()
    if isJob, _ := regexp.MatchString("^[" + regexp.QuoteMeta("*") + "0-9]", line); isJob == true {
      // Found a job, build JobConfig
      var jobObj job.JobConfig
      jobObj.Title = title + strconv.Itoa(titleCounter)
      lineElements := strings.Split(line, " ")
      jobObj.Schedule = strings.Join(lineElements[:5], " ")
      jobObj.Command = strings.Join(lineElements[5:], " ")

      handler.Job = append(handler.Job, jobObj)
      titleCounter++
    }
  }

  if err := scanner.Err(); err != nil {
    logrus.Fatal(err)
  }

  return handler
}

