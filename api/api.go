package api

import (
  "net/http"
  "encoding/json"
  "github.com/Sirupsen/logrus"
  "github.com/gorilla/mux"
  "github.com/brysearl/omicrond/job"
  "github.com/davecgh/go-spew/spew"
)

func init() {

}

func StartServer() {

  router := buildRoutes(mux.NewRouter())

  logrus.Info("Starting HTTP interface")
  logrus.Fatal(http.ListenAndServe(":12221", router))

  return
}

func buildRoutes(router *mux.Router) *mux.Router {

  router.HandleFunc("/.status", getStatus)
  router.HandleFunc("/job/list", getJobList)
  return router
}

func getStatus(w http.ResponseWriter,r *http.Request) {
  logrus.Info("API request for Omicrond status")
  w.Write([]byte("Omicrond is running\n"))
  return
}

func getJobList(w http.ResponseWriter,r *http.Request) {

  logrus.Info("API request for Omicrond job list")
  logrus.Info(spew.Sdump(job.RunningSchedule))
  encoder := json.NewEncoder(w)
  err := encoder.Encode(job.RunningSchedule.MakeAPIFormat())
  if err !=nil {
    logrus.Error(err)
  }
}
