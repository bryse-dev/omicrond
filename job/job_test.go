package job

import (
  "testing"
  "time"
  "fmt"
  "github.com/Sirupsen/logrus"
  "github.com/BurntSushi/toml"
  . "github.com/smartystreets/goconvey/convey"
)

const MYSQLDATETIMELAYOUT = "2006-01-02 15:04:05"

type TestingJobHandler struct {
  TestCase []TestingJob
}

type TestingJob struct {
  Job  JobConfig
  Test []TestConfig
}

type TestConfig struct {
  Description    string
  TestTime       time.Time
  ExpectedResult bool
}

func (h *TestingJobHandler) ParseTestJobConfig(confFile string) (error) {

  fmt.Println("Parsing unit-test file: " + confFile)
  _, err := toml.DecodeFile(confFile, &h)
  if err != nil {
    return err
  }

  for testCaseIndex, _ := range h.TestCase {
    err := h.TestCase[testCaseIndex].Job.ParseScheduleIntoFilters()
    if err != nil {
      return err
    }
  }
  return err
}

func TestJobParseAndFilterCreation(t *testing.T) {

  var handler = TestingJobHandler{}
  err := handler.ParseTestJobConfig("../unit_test/jobTestConfig.toml")
  if err != nil {
    fmt.Println(err)
  }

  for testCaseIndex, testCase := range handler.TestCase {
    for _, test := range testCase.Test {

      //Run test
      result := testCase.Job.RunThroughFilters(test.TestTime)

      var verb string
      if test.ExpectedResult == true {
        verb = "run"
      } else if test.ExpectedResult == false {
        verb = "not run"
      } else {
        logrus.Error("Cannot determine verb for testcase " + string(testCaseIndex))
      }
      Convey("With schedule [" + testCase.Job.Schedule + "] and date [" + test.TestTime.String() + "], the job should " + verb, t, func() {
        So(result, ShouldEqual, test.ExpectedResult)
      })
    }
  }
}
