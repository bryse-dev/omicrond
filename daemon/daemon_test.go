package daemon

import (
  "testing"
  "bytes"
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
  conf.Attr.APISSL = false

  // Start the daemon
  go StartDaemon()

  // Give it a few seconds to start
  time.Sleep(5 * time.Second)


  // Init an HTTP client
  client := &http.Client{}

  // scheduleGetList Route test success
  logrus.Info("Testing scheduleGetList")

  request, _ := http.NewRequest("GET","http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/schedule/get/list", nil)
  request.SetBasicAuth(conf.Attr.APIUser, conf.Attr.APIPassword)
  request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
  resp, err := client.Do(request)

  Convey("Should be able to build a Get request using daemon conf", t, func() {
    So(err, ShouldEqual, nil)
  })
  runApiTest(t,resp,true)
  resp.Body.Close()

  // scheduleGetJob Route test success
  logrus.Info("Testing scheduleGetJob")

  request, _ = http.NewRequest("GET","http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/schedule/get/job/Quarter_Hourly", nil)
  request.SetBasicAuth(conf.Attr.APIUser, conf.Attr.APIPassword)
  request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
  resp, err = client.Do(request)

  Convey("Should be able to build a Get request using daemon conf", t, func() {
    So(err, ShouldEqual, nil)
  })
  runApiTest(t,resp,true)
  resp.Body.Close()


  // scheduleEditJob Route test success
  logrus.Info("Testing scheduleEditJob")

  request, _ = http.NewRequest("POST","http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/schedule/edit/job/Minutely",
    bytes.NewBufferString(url.Values{
      "label": {"Every Other Minute"},
      "schedule": {"*/2 * * * *"}}.Encode()))
  request.SetBasicAuth(conf.Attr.APIUser, conf.Attr.APIPassword)
  request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
  resp, err = client.Do(request)

  Convey("Should be able to build a Get request using daemon conf", t, func() {
    So(err, ShouldEqual, nil)
  })
  runApiTest(t,resp,true)
  resp.Body.Close()

  // scheduleCreateJob Route test failure
  logrus.Info("Testing scheduleCreateJob with underscores")
  request, _ = http.NewRequest("POST","http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/schedule/create/job",
    bytes.NewBufferString(url.Values{
      "label": {"current_dir_every_minute"},
      "schedule": {"* * * * *"},
      "groupName": {"test"},
      "command": {"/bin/pwd"}}.Encode()))
  request.SetBasicAuth(conf.Attr.APIUser, conf.Attr.APIPassword)
  request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
  resp, err = client.Do(request)
  Convey("Should be able to build a Get request using daemon conf", t, func() {
    So(err, ShouldEqual, nil)
  })
  runApiTest(t,resp,false)
  resp.Body.Close()

  // scheduleCreateJob Route test success
  logrus.Info("Testing scheduleCreateJob")
  request, _ = http.NewRequest("POST","http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/schedule/create/job",
    bytes.NewBufferString(url.Values{
      "label": {"current dir every minute"},
      "schedule": {"* * * * *"},
      "groupName": {"test"},
      "command": {"/bin/pwd"}}.Encode()))
  request.SetBasicAuth(conf.Attr.APIUser, conf.Attr.APIPassword)
  request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
  resp, err = client.Do(request)
  Convey("Should be able to build a Get request using daemon conf", t, func() {
    So(err, ShouldEqual, nil)
  })
  runApiTest(t,resp,true)
  resp.Body.Close()

  // scheduleDeleteJob Route test success
  logrus.Info("Testing scheduleDeleteJob")
  request, _ = http.NewRequest("POST","http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/schedule/delete/job/Quarter_Hourly",
    bytes.NewBufferString(url.Values{
      "label": {"current dir every minute"},
      "schedule": {"* * * * *"},
      "groupName": {"test"},
      "command": {"/bin/pwd"}}.Encode()))
  request.SetBasicAuth(conf.Attr.APIUser, conf.Attr.APIPassword)
  request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
  resp, err = client.Do(request)
  Convey("Should be able to build a Get request using daemon conf", t, func() {
    So(err, ShouldEqual, nil)
  })
  runApiTest(t,resp,true)
  resp.Body.Close()

  // Turn off the daemon
  runningChanComm <- api.ChanComm{Signal: "shutdown", RunningSchedule: job.JobSchedule{} }
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
    So(string(body), ShouldNotContainSubstring, "Unauthorized")
  })
}
