package main

import (
  "testing"
  "os"
  . "github.com/smartystreets/goconvey/convey"
  "github.com/brysearl/omicrond/job"
)

// TestConvertLegacyConfig - Convert a sample crontab export into an Omicrond TOML config file.  Read the TOML config
//   and make sure that it is viable for running Omicrond
func TestConvertLegacyConfig(t *testing.T) {

  var err error
  configFile = "unit_test/sample_crontab.txt"
  fileOutString := "unit_test/example_crontab_converted.toml"
  fileOut, err = os.Create(fileOutString)

  main()

  Convey("Legacy config should have been read and the converted config written without error", t, func() {
    So(err, ShouldEqual, nil)
  })

  var schedule = job.JobHandler{}
  err = schedule.ParseJobConfig(fileOutString)
  Convey("Should be able to unmarshal converted config into JobHandler", t, func() {
    So(err, ShouldEqual, nil)
  })

  // Cleanup test
  err = os.Remove(fileOutString)
}
