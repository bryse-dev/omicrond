package daemon

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
  conf.Attr.JobConfigPath = "../unit_test/TestApiRoutes.toml"
  conf.Attr.APIPort = 47685

  // Start the daemon
  go StartDaemon()

  // Give it a few seconds to start
  time.Sleep(5 * time.Second)

  // getJobList Route test success
  logrus.Info("Testing getJobList")
  resp, err := http.Get("http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/get/job/list")
  Convey("Should be able to build a Get request using daemon conf", t, func() {
    So(err, ShouldEqual, nil)
  })
  runApiTest(t,resp,true)
  resp.Body.Close()

  // getJobByLabel Route test success
  logrus.Info("Testing getJobByLabel")
  resp, err = http.Get("http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/get/job/Quarter_Hourly")
  Convey("Should be able to build a Get request using daemon conf", t, func() {
    So(err, ShouldEqual, nil)
  })
  runApiTest(t,resp,true)
  resp.Body.Close()


  // modifyJobByLabel Route test success
  logrus.Info("Testing modifyJobByLabel")
  resp, err = http.PostForm("http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/edit/job/Minutely",
    url.Values{
      "label": {"Every Other Minute"},
      "schedule": {"*/2 * * * *"}})
  Convey("Should be able to build a Get request using daemon conf", t, func() {
    So(err, ShouldEqual, nil)
  })
  runApiTest(t,resp,true)
  resp.Body.Close()

  // createJob Route test failure
  logrus.Info("Testing createJob with underscores")
  resp, err = http.PostForm("http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/create/job",
    url.Values{
      "label": {"current_dir_every_minute"},
      "schedule": {"* * * * *"},
      "groupName": {"test"},
      "command": {"/bin/pwd"}})
  Convey("Should be able to build a Get request using daemon conf", t, func() {
    So(err, ShouldEqual, nil)
  })

  // createJob Route test success
  runApiTest(t,resp,false)
  resp.Body.Close()
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
  runApiTest(t,resp,true)
  resp.Body.Close()

  // deleteJobByLabel Route test success
  logrus.Info("Testing deleteJobByLabel")
  resp, err = http.PostForm("http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/delete/job/Quarter_Hourly",
    url.Values{
      "label": {"current dir every minute"},
      "schedule": {"* * * * *"},
      "groupName": {"test"},
      "command": {"/bin/pwd"}})
  Convey("Should be able to build a Get request using daemon conf", t, func() {
    So(err, ShouldEqual, nil)
  })
  runApiTest(t,resp,true)
  resp.Body.Close()

  // Turn off the daemon
  runningChanComm <- api.ChanComm{Signal: "shutdown", Handler: job.JobSchedule{} }
}

func runApiTest (t *testing.T, resp *http.Response, testingNoErr bool) {

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
    if testingNoErr == true {
      So(json.Error, ShouldEqual, "")
    }else{
      So(json.Error, ShouldNotEqual, "")
    }
  })
}
