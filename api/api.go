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

  router.HandleFunc("/.status", getStatus).Methods("GET")
  router.HandleFunc("/get/job/list", getJobList).Methods("GET")
  router.HandleFunc("/get/job/{jobID:[0-9]+}", getJobByID).Methods("GET")
  router.HandleFunc("/post/edit/job/{jobID:[0-9]+}", modifyJobByID).Methods("POST")
  return router
}

// getStatus - Send the status of the server.  Used as unit test, if you get a 404 your test failed.
func getStatus(w http.ResponseWriter, r *http.Request) {
  logrus.Debug("API request for Omicrond status")

  w.Write([]byte("Omicrond is running"))
  return
}

// getJobList - Send a JSON representation of the JobHandler object within job.go
func getJobList(w http.ResponseWriter, r *http.Request) {
  logrus.Debug("API request for Omicrond job list")

  // Return the running schedule in JSON format
  encoder := json.NewEncoder(w)
  apiRunningSchedule, err := job.RunningSchedule.MakeAPIFormat()
  if err != nil {
    w.Write([]byte("Error: " + err.Error()))
  }
  err = encoder.Encode(apiRunningSchedule)
  if err != nil {
    w.Write([]byte("Error: " + err.Error()))
  }

  return
}

func getJobByID(w http.ResponseWriter, r *http.Request) {
  logrus.Debug("API request for single Omicrond job configuration")

  // Assign the JSON encoder
  encoder := json.NewEncoder(w)

  // Convert the route variables
  vars := mux.Vars(r)
  jobIDStr := vars["jobID"]
  jobID, err := strconv.Atoi(jobIDStr)
  if err != nil {
    http.Error(w,"{ \"Error\":\"" + err.Error() + "\"}",http.StatusBadRequest)
    return
  }

  // Retrieve the Job
  requestedJob, err := job.RunningSchedule.GetJobByID(jobID)
  if err != nil {
    http.Error(w,"{ \"Error\":\"" + err.Error() + "\"}",http.StatusBadRequest)
    return
  }

  // Convert the job to API
  apiRequestedJob, err := requestedJob.MakeAPIFormat(job.RunningSchedule)
  if err != nil {
    http.Error(w,"{ \"Error\":\"" + err.Error() + "\"}",http.StatusBadRequest)
    return
  }

  // Return in JSON format
  err = encoder.Encode(apiRequestedJob)
  if err != nil {
    http.Error(w,"{ \"Error\":\"" + err.Error() + "\"}",http.StatusBadRequest)
    return
  }

  return
}

func modifyJobByID(w http.ResponseWriter, r *http.Request) {
  logrus.Debug("API request to modify Omicrond job configuration")

  // Convert the route variables
  //vars := mux.Vars(r)
  //jobIDStr := vars["jobID"]
  //jobID, err := strconv.Atoi(jobIDStr)

  // Retrieve the Job
  //requestedJob, err := job.RunningSchedule.GetJobByID(jobID)
  //if err != nil {
  //  w.Write([]byte("Error: " + err))
  //}

}
