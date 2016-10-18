package api

import (
  "testing"
  "strconv"
  "io/ioutil"
  "net/http"
  "time"
  . "github.com/smartystreets/goconvey/convey"
  "github.com/brysearl/omicrond/conf"
  "github.com/brysearl/omicrond/job"
)

func TestStartServer(t *testing.T) {

  // Start the server
  var dummyChan chan job.JobHandler
  go StartServer(dummyChan)

  // Give it a second to start
  time.Sleep(1 * time.Second)

  // Send a GET request for the /.status route
  response, _ := http.Get("http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/.status")

  // Extract the body of the response
  body, _ := ioutil.ReadAll(response.Body)

  Convey("Should be able to query the API route /.status", t, func() {
    So(string(body), ShouldEqual, "Omicrond is running")
  })
}

