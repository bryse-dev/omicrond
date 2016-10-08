package api

import (
  "net/http"
  "encoding/json"
  "strconv"
  "time"
  "github.com/Sirupsen/logrus"
  "github.com/gorilla/mux"
  "github.com/brysearl/omicrond/job"
  "github.com/brysearl/omicrond/conf"
)

// StartServer - Create a TCP server running on the address and port configured in conf.go or cli arg.
//  Should be run in a goroutine
func StartServer() {

  router := buildRoutes(mux.NewRouter())

  logrus.Info("Starting HTTP interface")
  srv := &http.Server{
    Handler:      router,
    Addr:         conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort),
    // Good practice: enforce timeouts for servers you create!
    WriteTimeout: time.Duration(conf.Attr.APITimeout) * time.Second,
    ReadTimeout:  time.Duration(conf.Attr.APITimeout) * time.Second,
  }

  logrus.Fatal(srv.ListenAndServe())

}

// buildRoutes - Configure API routes and their functions
func buildRoutes(router *mux.Router) *mux.Router {

  router.HandleFunc("/.status", getStatus)
  router.HandleFunc("/job/list", getJobList)
  router.HandleFunc("/job/{jobID:[0-9]+}", getJobByID)
  return router
}

// getStatus - Send the status of the server.  Used as unit test, if you get a 404 your test failed.
func getStatus(w http.ResponseWriter, r *http.Request) {
  logrus.Info("API request for Omicrond status")
  w.Write([]byte("Omicrond is running"))
  return
}

// getJobList - Send a JSON representation of the JobHandler object within job.go
func getJobList(w http.ResponseWriter, r *http.Request) {

  logrus.Info("API request for Omicrond job list")
  encoder := json.NewEncoder(w)
  err := encoder.Encode(job.RunningSchedule.MakeAPIFormat())
  if err != nil {
    logrus.Error(err)
  }

  return
}

func getJobByID(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  jobIDStr := vars["jobID"]
  jobID, err := strconv.Atoi(jobIDStr)
  if err != nil {
    logrus.Error(err)
  }

  logrus.Info("API request for single Omicrond job configuration")
  encoder := json.NewEncoder(w)
  err = encoder.Encode(job.RunningSchedule.MakeAPIFormat().Job[jobID])
  if err != nil {
    logrus.Error(err)
  }

  return
}
