package api

import (
  "testing"
  "strconv"
  "io/ioutil"
  "net/http"
  "time"
  . "github.com/smartystreets/goconvey/convey"
  "github.com/brysearl/omicrond/conf"
)

func TestStartServer(t *testing.T) {

  // Start the server
  var dummyChan chan ChanComm
  conf.Attr.APISSL = false
  go StartServer(dummyChan)

  // Give it a second to start
  time.Sleep(1 * time.Second)

  // Send a GET request for the /.status route
  client := &http.Client{}
  request, _ := http.NewRequest("GET","http://" + conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort) + "/.status", nil)
  request.SetBasicAuth(conf.Attr.APIUser, conf.Attr.APIPassword)
  response, err := client.Do(request)

  // Extract the body of the response
  body, _ := ioutil.ReadAll(response.Body)

  Convey("Should be able to query the API route /.status", t, func() {
    So(err, ShouldEqual, nil)
    So(string(body), ShouldEqual, "Omicrond is running")
  })
}

