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
//"github.com/davecgh/go-spew/spew"

type ChanComm struct {
  Signal  string
  Handler job.JobHandler
}

var runningChanComm chan ChanComm

// StartServer - Create a TCP server running on the address and port configured in conf.go or cli arg.
//  Should be run in a goroutine
func StartServer(commChannel chan ChanComm) {

  runningChanComm = commChannel
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
  router.HandleFunc("/edit/job/{jobID:[0-9]+}", modifyJobByID).Methods("POST")
  router.HandleFunc("/create/job", newJob).Methods("POST")
  router.HandleFunc("/delete/job/{jobID:[0-9]+}", deleteJobByID).Methods("POST")
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

  encoder := json.NewEncoder(w)

  // Request the current running schedule from the main scheduling loop
  runningChanComm <- ChanComm{Signal: "getRunningSchedule", Handler: job.JobHandler{} }
  returnComm := <-runningChanComm
  currentSchedule := returnComm.Handler
  apiRunningSchedule, err := currentSchedule.MakeAPIFormat()
  if err != nil {
    w.Write([]byte("Error: " + err.Error()))
  }

  // Return the running schedule in JSON format
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
    http.Error(w, "{ \"Error\":\"" + err.Error() + "\"}", http.StatusBadRequest)
    return
  }

  // Request the current running schedule from the main scheduling loop
  runningChanComm <- ChanComm{Signal: "getRunningSchedule", Handler: job.JobHandler{} }
  returnComm := <-runningChanComm
  currentSchedule := returnComm.Handler

  // Retrieve the Job
  requestedJob, err := currentSchedule.GetJobByID(jobID)
  if err != nil {
    http.Error(w, "{ \"Error\":\"" + err.Error() + "\"}", http.StatusBadRequest)
    return
  }

  // Convert the job to API
  apiRequestedJob, err := requestedJob.MakeAPIFormat(currentSchedule)
  if err != nil {
    http.Error(w, "{ \"Error\":\"" + err.Error() + "\"}", http.StatusBadRequest)
    return
  }

  // Return in JSON format
  err = encoder.Encode(apiRequestedJob)
  if err != nil {
    http.Error(w, "{ \"Error\":\"" + err.Error() + "\"}", http.StatusBadRequest)
    return
  }

  return
}

func modifyJobByID(w http.ResponseWriter, r *http.Request) {
  logrus.Debug("API request to modify Omicrond job configuration")

  // Convert the route variables
  vars := mux.Vars(r)
  jobIDStr := vars["jobID"]
  jobID, err := strconv.Atoi(jobIDStr)

  // Request the current running schedule from the main scheduling loop
  runningChanComm <- ChanComm{Signal: "getRunningSchedule", Handler: job.JobHandler{} }
  returnComm := <-runningChanComm
  currentSchedule := returnComm.Handler

  // Retrieve the Job
  requestedJob, err := currentSchedule.GetJobByID(jobID)
  if err != nil {
    http.Error(w, "{ \"Error\":\"" + err.Error() + "\"}", http.StatusBadRequest)
    return
  }

  // Change the fields to the new settings
  newLabel := r.FormValue("label")
  if newLabel != "" {
    requestedJob.Label = newLabel
  }
  newScheduleStr := r.FormValue("schedule")
  if newScheduleStr != "" {
    requestedJob.Schedule = newScheduleStr
    err = requestedJob.ParseScheduleIntoFilters(false)
    if err != nil {
      http.Error(w, "{ \"Error\":\"" + err.Error() + "\"}", http.StatusBadRequest)
      return
    }
  }
  newCommand := r.FormValue("command")
  if newCommand != "" {
    requestedJob.Command = newCommand
  }
  newGroupName := r.FormValue("groupName")
  if newGroupName != "" {
    requestedJob.GroupName = newGroupName
  }

  // Put the job back into the schedule
  newSchedule := currentSchedule
  newSchedule.Job[jobID] = requestedJob

  // Make sure the changes are okay
  err = newSchedule.CheckConfig()
  if err != nil {
    http.Error(w, "{ \"Error\":\"" + err.Error() + "\"}", http.StatusBadRequest)
    return
  }

  // Put the new schedule into rotation
  runningChanComm <- ChanComm{Signal: "replaceRunningSchedule", Handler: newSchedule}

  return
}

func newJob(w http.ResponseWriter, r *http.Request) {
  logrus.Debug("API request to create Omicrond job configuration")

  // Request the current running schedule from the main scheduling loop
  runningChanComm <- ChanComm{Signal: "getRunningSchedule", Handler: job.JobHandler{} }
  returnComm := <-runningChanComm
  currentSchedule := returnComm.Handler

  // Create the empty Job
  newJob := job.JobConfig{}

  // Change the fields to the new settings
  newLabel := r.FormValue("label")
  if newLabel != "" {
    newJob.Label = newLabel
  } else {
    http.Error(w, "{ \"Error\":\"" + "Requires parameter[label]" + "\"}", http.StatusBadRequest)
    return
  }
  newScheduleStr := r.FormValue("schedule")
  if newScheduleStr != "" {
    newJob.Schedule = newScheduleStr
    err := newJob.ParseScheduleIntoFilters(false)
    if err != nil {
      http.Error(w, "{ \"Error\":\"" + err.Error() + "\"}", http.StatusBadRequest)
      return
    }
  } else {
    http.Error(w, "{ \"Error\":\"" + "Requires parameter[schedule]" + "\"}", http.StatusBadRequest)
    return
  }
  newCommand := r.FormValue("command")
  if newCommand != "" {
    newJob.Command = newCommand
  } else {
    http.Error(w, "{ \"Error\":\"" + "Requires parameter[command]" + "\"}", http.StatusBadRequest)
    return
  }
  newGroupName := r.FormValue("groupName")
  if newGroupName != "" {
    newJob.GroupName = newGroupName
  } else {
    http.Error(w, "{ \"Error\":\"" + "Requires parameter[groupName]" + "\"}", http.StatusBadRequest)
    return
  }

  // Put the new job into the schedule
  newSchedule := currentSchedule
  newSchedule.Job = append(newSchedule.Job, newJob)

  // Make sure the changes are okay
  err := newSchedule.CheckConfig()
  if err != nil {
    http.Error(w, "{ \"Error\":\"" + err.Error() + "\"}", http.StatusBadRequest)
    return
  }

  // Put the new schedule into rotation
  runningChanComm <- ChanComm{Signal: "replaceRunningSchedule", Handler: newSchedule}

  return
}

func deleteJobByID(w http.ResponseWriter, r *http.Request) {
  logrus.Debug("API request to delete Omicrond job configuration")

  // Convert the route variables
  vars := mux.Vars(r)
  jobIDStr := vars["jobID"]
  jobID, err := strconv.Atoi(jobIDStr)

  // Request the current running schedule from the main scheduling loop
  runningChanComm <- ChanComm{Signal: "getRunningSchedule", Handler: job.JobHandler{} }
  returnComm := <-runningChanComm
  currentSchedule := returnComm.Handler

  // Retrieve the Job
  _, err = currentSchedule.GetJobByID(jobID)
  if err != nil {
    http.Error(w, "{ \"Error\":\"" + err.Error() + "\"}", http.StatusBadRequest)
    return
  }

  // Put the job back into the schedule
  newSchedule := currentSchedule
  newSchedule.Job = append(newSchedule.Job[:jobID], newSchedule.Job[jobID + 1:]...)

  // Make sure the changes are okay
  err = newSchedule.CheckConfig()
  if err != nil {
    http.Error(w, "{ \"Error\":\"" + err.Error() + "\"}", http.StatusBadRequest)
    return
  }

  // Put the new schedule into rotation
  runningChanComm <- ChanComm{Signal: "replaceRunningSchedule", Handler: newSchedule}

  return
}
