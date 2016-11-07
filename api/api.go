package api

import (
  "net/http"
  "encoding/json"
  "strconv"
  "time"
  "errors"
  "github.com/Sirupsen/logrus"
  "github.com/gorilla/mux"
  "github.com/brysearl/omicrond/job"
  "github.com/brysearl/omicrond/conf"
  "github.com/goji/httpauth"
)
//"github.com/davecgh/go-spew/spew"

type ChanComm struct {
  Signal  string
  Handler job.JobSchedule
}

var runningChanComm chan ChanComm

// StartServer - Create a TCP server running on the address and port configured in conf.go or cli arg.
//  Should be run in a goroutine
func StartServer(commChannel chan ChanComm) {

  runningChanComm = commChannel
  router := buildRoutes(mux.NewRouter())

  logrus.Info("Starting HTTP interface")
  srv := &http.Server{
    Handler:      httpauth.SimpleBasicAuth(conf.Attr.APIUser, conf.Attr.APIPassword)(router),
    Addr:         conf.Attr.APIAddress + ":" + strconv.Itoa(conf.Attr.APIPort),
    // Good practice: enforce timeouts for servers you create!
    WriteTimeout: time.Duration(conf.Attr.APITimeout) * time.Second,
    ReadTimeout:  time.Duration(conf.Attr.APITimeout) * time.Second,
  }

  if conf.Attr.APISSL == true {
    logrus.Fatal(srv.ListenAndServeTLS(conf.Attr.APIPubKeyPath, conf.Attr.APIPrivKeyPath))
  }else{
    logrus.Fatal(srv.ListenAndServe())
  }

}

// buildRoutes - Configure API routes and their functions
func buildRoutes(router *mux.Router) *mux.Router {

  router.HandleFunc("/.status", getStatus).Methods("GET")
  router.HandleFunc("/get/job/list", getJobList).Methods("GET")
  router.HandleFunc("/get/job/{jobLabel:[a-zA-Z0-9_]+}", getJobByLabel).Methods("GET")
  router.HandleFunc("/edit/job/{jobLabel:[a-zA-Z0-9_]+}", modifyJobByLabel).Methods("POST")
  router.HandleFunc("/create/job", createJob).Methods("POST")
  router.HandleFunc("/delete/job/{jobLabel:[a-zA-Z0-9_]+}", deleteJobByLabel).Methods("POST")


  return router
}

// getStatus - Send the status of the server.  Used as unit test, if you get a 404 your test failed.
func getStatus(w http.ResponseWriter, r *http.Request) {
  logrus.Debug("API request for Omicrond status")

  w.Write([]byte("Omicrond is running"))
  return
}

// getJobList - Send a JSON representation of the JobSchedule object within job.go
func getJobList(w http.ResponseWriter, r *http.Request) {
  logrus.Debug("API request for Omicrond job list")

  encoder := json.NewEncoder(w)

  // Request the current running schedule from the main scheduling loop
  runningChanComm <- ChanComm{Signal: "getRunningSchedule", Handler: job.JobSchedule{} }
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

func getJobByLabel(w http.ResponseWriter, r *http.Request) {
  logrus.Debug("API request for single Omicrond job configuration")

  // Assign the JSON encoder
  encoder := json.NewEncoder(w)

  // Convert the route variables
  vars := mux.Vars(r)
  jobLabelStr := vars["jobLabel"]

  // Request the current running schedule from the main scheduling loop
  runningChanComm <- ChanComm{Signal: "getRunningSchedule", Handler: job.JobSchedule{} }
  returnComm := <-runningChanComm
  currentSchedule := returnComm.Handler

  // Retrieve the Job
  requestedJob, _, err := currentSchedule.GetJobByLabel(jobLabelStr)
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

func modifyJobByLabel(w http.ResponseWriter, r *http.Request) {
  logrus.Debug("API request to modify Omicrond job configuration")

  // Convert the route variables
  vars := mux.Vars(r)
  jobLabelStr := vars["jobLabel"]

  // Request the current running schedule from the main scheduling loop
  runningChanComm <- ChanComm{Signal: "getRunningSchedule", Handler: job.JobSchedule{} }
  returnComm := <-runningChanComm
  currentSchedule := returnComm.Handler

  // Retrieve the Job
  requestedJob, jobIndex, err := currentSchedule.GetJobByLabel(jobLabelStr)
  if err != nil {
    http.Error(w, "{ \"Error\":\"" + err.Error() + "\"}", http.StatusBadRequest)
    return
  }

  err = parseFormFieldsIntoJobConfig(&requestedJob, r)
  if err != nil {
    http.Error(w, "{ \"Error\":\"" + err.Error() + "\"}", http.StatusBadRequest)
    return
  }

  // Put the job back into the schedule
  newSchedule := currentSchedule
  if err != nil {
    http.Error(w, "{ \"Error\":\"" + err.Error() + "\"}", http.StatusBadRequest)
    return
  }
  newSchedule.Job[jobIndex] = requestedJob

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

func createJob(w http.ResponseWriter, r *http.Request) {
  logrus.Debug("API request to create Omicrond job configuration")

  // Request the current running schedule from the main scheduling loop
  runningChanComm <- ChanComm{Signal: "getRunningSchedule", Handler: job.JobSchedule{} }
  returnComm := <-runningChanComm
  currentSchedule := returnComm.Handler

  // Create the empty Job
  newJob := job.JobConfig{}

  err := parseFormFieldsIntoJobConfig(&newJob, r)
  if err != nil {
    http.Error(w, "{ \"Error\":\"" + err.Error() + "\"}", http.StatusBadRequest)
    return
  }

  // Put the new job into the schedule
  newSchedule := currentSchedule
  newSchedule.Job = append(newSchedule.Job, newJob)

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

func deleteJobByLabel(w http.ResponseWriter, r *http.Request) {
  logrus.Debug("API request to delete Omicrond job configuration")

  // Convert the route variables
  vars := mux.Vars(r)
  jobLabelStr := vars["jobLabel"]

  // Request the current running schedule from the main scheduling loop
  runningChanComm <- ChanComm{Signal: "getRunningSchedule", Handler: job.JobSchedule{} }
  returnComm := <-runningChanComm
  currentSchedule := returnComm.Handler

  // Retrieve the Job
  _, jobIndex, err := currentSchedule.GetJobByLabel(jobLabelStr)
  if err != nil {
    http.Error(w, "{ \"Error\":\"" + err.Error() + "\"}", http.StatusBadRequest)
    return
  }

  // Put the job back into the schedule
  newSchedule := currentSchedule
  newSchedule.Job = append(newSchedule.Job[:jobIndex], newSchedule.Job[jobIndex + 1:]...)

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

func parseFormFieldsIntoJobConfig (newJob *job.JobConfig, r *http.Request) (error) {

  var err error

  // Change the fields to the new settings
  newLabel := r.PostFormValue("label")
  if newLabel != "" {
    newJob.Label = newLabel
  } else {
    return errors.New("{ \"Error\":\"" + "Requires parameter[label]" + "\"}")
  }
  newScheduleStr := r.PostFormValue("schedule")
  if newScheduleStr != "" {
    newJob.Schedule = newScheduleStr
    err := newJob.ParseScheduleIntoFilters(false)
    if err != nil {
      return errors.New("{ \"Error\":\"" + err.Error() + "\"}")
      return errors.New("Error parsing form fields")
    }
  } else {
    return errors.New("{ \"Error\":\"" + "Requires parameter[schedule]" + "\"}")
  }
  newCommand := r.PostFormValue("command")
  if newCommand != "" {
    newJob.Command = newCommand
  } else {
    return errors.New("{ \"Error\":\"" + "Requires parameter[command]" + "\"}")
  }
  newGroupName := r.PostFormValue("groupName")
  if newGroupName != "" {
    newJob.GroupName = newGroupName
  } else {
    return errors.New("{ \"Error\":\"" + "Requires parameter[groupName]" + "\"}")
  }
  newLocking := r.PostFormValue("locking")
  if newLocking != "" {
    if newLocking == "true" {
      newJob.Locking = true
    } else if newLocking == "false" {
      newJob.Locking = true
    } else {
      return errors.New("{ \"Error\":\"form field 'locking' must equal either 'true' or 'false'\"}")
    }
  }

  return err
}
