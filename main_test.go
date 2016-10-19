package main

import (
  "testing"
  "net/http"
  "io/ioutil"
  "strconv"
  "time"
  "strings"
  "encoding/json"
  "net/url"
  . "github.com/smartystreets/goconvey/convey"
  "github.com/brysearl/omicrond/job"
  "github.com/brysearl/omicrond/conf"
  "github.com/brysearl/omicrond/api"
  "github.com/Sirupsen/logrus"
)

//TestConvertLegacyConfig - Convert a sample crontab export into an Omicrond TOML config file.  Read the TOML config
//  and make sure that it is viable for running Omicrond
func TestApiRoutes(t *testing.T) {

  // Used by main to disable certain features
  isUnitTest = true

  // Set the unit_test config to be the config
  conf.Attr.JobConfigPath = "unit_test/TestApiRoutes.toml"
  conf.Attr.APIPort = 47685

  // Start the daemon
  go main()

  // Give it a few seconds to start
  time.Sleep(5 * time.Second)

  // getJobList Route test
  logrus.Info("Testing getJobList")
  resp, err := http.Get("http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/get/job/list")
  Convey("Should be able to build a Get request using daemon conf", t, func() {
    So(err, ShouldEqual, nil)
  })
  runApiTest(t,resp)
  resp.Body.Close()

  // getJobByID Route test
  logrus.Info("Testing getJobByID")
  resp, err = http.Get("http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/get/job/0")
  Convey("Should be able to build a Get request using daemon conf", t, func() {
    So(err, ShouldEqual, nil)
  })
  runApiTest(t,resp)
  resp.Body.Close()


  // modifyJobByID Route test
  logrus.Info("Testing modifyJobByID")
  resp, err = http.PostForm("http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/edit/job/0",
    url.Values{
      "label": {"Every Other Minute"},
      "schedule": {"*/2 * * * *"}})
  Convey("Should be able to build a Get request using daemon conf", t, func() {
    So(err, ShouldEqual, nil)
  })
  runApiTest(t,resp)
  resp.Body.Close()

  // createJob Route test
  logrus.Info("Testing createJob")
  resp, err = http.PostForm("http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/create/job",
    url.Values{
      "label": {"current dir every minute"},
      "schedule": {"* * * * *"},
      "groupName": {"test"},
      "command": {"/bin/pwd"}})
  Convey("Should be able to build a Get request using daemon conf", t, func() {
    So(err, ShouldEqual, nil)
  })
  runApiTest(t,resp)
  resp.Body.Close()

  // deleteJobByID Route test
  logrus.Info("Testing deleteJobByID")
  resp, err = http.PostForm("http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/delete/job/1",
    url.Values{
      "label": {"current dir every minute"},
      "schedule": {"* * * * *"},
      "groupName": {"test"},
      "command": {"/bin/pwd"}})
  Convey("Should be able to build a Get request using daemon conf", t, func() {
    So(err, ShouldEqual, nil)
  })
  runApiTest(t,resp)
  resp.Body.Close()

  // Turn off the daemon
  runningChanComm <- api.ChanComm{Signal: "shutdown", Handler: job.JobHandler{} }
}

func runApiTest (t *testing.T, resp *http.Response) {

  // Struct to test responses
  type jsonResponse struct {
    Error string
  }

  body, err := ioutil.ReadAll(resp.Body)
  Convey("Should be able to read the response body", t, func() {
    So(err, ShouldEqual, nil)
  })

  dec := json.NewDecoder(strings.NewReader(string(body)))
  var json jsonResponse
  dec.Decode(&json)
  Convey("API call should not return an error", t, func() {
    So(json.Error, ShouldEqual, "")
  })
}
