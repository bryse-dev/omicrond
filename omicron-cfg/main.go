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
)

// JobHandlerToPrint - Keep all the jobs together in an iterable slice
type JobHandlerToPrint struct {
  Job []JobConfigToPrint
}

// JobConfigToPrint - Object representing a single scheduled job
type JobConfigToPrint struct {
  Title     string // The name of the job.  Used in logging
  Exec      string // String to be run on the system
  GroupName string // Used to relate jobs and in logging *unused*
  Schedule  string // Traditional encoded string to represent the schedule
}

// Configure command line arguments
var logLevelPtr = flag.Int("v", 2, "Set the debug level: 1 = Panic, 2 = Fatal, 3 = Error, 4 = Warn, 5 = Info, 6 = Debug")
var configFilePtr = flag.String("config", "", "Path to the daemon configuration file")
var configFile string

func init() {

  // Retrieve command line arguments
  flag.Parse()

  // Set the path to the daemon config file
  configFile = *configFilePtr

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

  jobHandler := readConfigFile(configFile)

  if err := toml.NewEncoder(os.Stdout).Encode(jobHandler); err != nil {
    logrus.Fatal("Error encoding TOML: %s", err)
  }
}

func readConfigFile(configFile string) (JobHandlerToPrint) {

  var handler JobHandlerToPrint

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
  //removeStartingWhitespace := regexp.MustCompile("^ *#? *")

  // Read in file line by line and build JobConfig objects
  for scanner.Scan() {
    line := scanner.Text()
    if isJob, _ := regexp.MatchString("^[" + regexp.QuoteMeta("*") + "0-9]", line); isJob == true {
      // Found a job, build JobConfig
      var jobObj JobConfigToPrint
      jobObj.Title = title + strconv.Itoa(titleCounter)
      lineElements := strings.Split(line, " ")
      jobObj.Schedule = strings.Join(lineElements[:5], " ")
      jobObj.Exec = strings.Join(lineElements[5:], " ")

      handler.Job = append(handler.Job, jobObj)
    }
  }

  if err := scanner.Err(); err != nil {
    logrus.Fatal(err)
  }

  return handler
}

