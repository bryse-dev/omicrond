package daemon

import (
  "time"
  "errors"
  "github.com/Sirupsen/logrus"
  "github.com/brysearl/omicrond/conf"
  "github.com/brysearl/omicrond/job"
  "github.com/brysearl/omicrond/api"
)

var runningChanComm chan api.ChanComm
var isUnitTest bool

func StartDaemon() {

  logrus.Info("Reading job configuration file: " + conf.Attr.JobConfigPath)
  var schedule job.JobSchedule
  err := schedule.ParseJobConfig(conf.Attr.JobConfigPath)
  if err != nil {
    logrus.Fatal(err)
  }

  logrus.Info("Starting API")
  runningChanComm = make(chan api.ChanComm)
  go api.StartServer(runningChanComm)
  time.Sleep(time.Second)

  logrus.Info("Starting scheduling loop")
  startSchedulingLoop(schedule, conf.Attr.JobConfigPath)
}

// startSchedulingLoop - Endless loop that checks jobs every minute and executes them if scheduled
func startSchedulingLoop(schedule job.JobSchedule, jobConfig string) {

  // Keep track of the last minute that was run.  This way we can sit quietly until the next minute comes.
  lastCheckTime := time.Now().Truncate(time.Minute)

  // Keep track of running jobs
  Running := job.RunningJobTracker{}
  Running.Jobs = make(map[string]job.RunningJob)

  // To infinity, and beyond
  for {

    // Get the current minute with the seconds rounded down
    currentTime := time.Now().Truncate(time.Minute)

    // Wait patiently for a new minute
    if currentTime != lastCheckTime {

      //Check each configured job to see if it needs to be run in this minute
      logrus.Debug("Running filters: " + currentTime.String())
      for jobIndex, _ := range schedule.Job {

        logrus.Debug("Checking: " + schedule.Job[jobIndex].Label)
        runJob := schedule.Job[jobIndex].CheckIfScheduled(currentTime)

        if runJob == true {

          // Check to see if its running and skip if locking attribute enabled
          if schedule.Job[jobIndex].Locking == true {
            var skip bool
            Running.RLock()
            for runToken, _ := range Running.Jobs {
              if schedule.Job[jobIndex].Label == Running.Jobs[runToken].Config.Label {
                skip = true
                break
              }
            }
            Running.RUnlock()

            if skip {
              logrus.Info("[" + schedule.Job[jobIndex].Label + "] currently running and locked.  Skipping.")
              continue
            }
          }

          // Prep the Job for Running and create a tracking token
          runToken := job.CreateRunToken()
          newJob := job.RunningJob{
            Token: runToken,
            Config: schedule.Job[jobIndex],
            Channel: make(chan string)}

          // Add the tracking token to the tracker
          logrus.Debug("Adding job " + runToken + " to tracker")
          Running.Lock()
          Running.Jobs[runToken] = newJob
          Running.Unlock()

          // Split off the job into a goroutine
          go func(Running job.RunningJobTracker, newJob job.RunningJob, runToken string, isUnitTest bool) {

            // Start the job
            if isUnitTest != true {
              newJob.Run()
            }

            // On completion, remove the tracking token from the tracker
            Running.RLock()
            _, ok := Running.Jobs[runToken]
            Running.RUnlock()

            if ok {
              logrus.Debug("Removing job " + runToken + " from tracker")
              Running.Lock()
              delete(Running.Jobs, runToken)
              Running.Unlock()
            } else {
              logrus.Error("Could not find runToken on completion")
            }
          }(Running, newJob, runToken, isUnitTest)
        }
      }

      // Update the minute lock and take a break
      lastCheckTime = currentTime

    } else {

      // Between scheduling, be open to schedule changes via API
      stop := false
      for stop == false {

        // Determine the amount of free time available to listen to a channel
        timeout := time.Now().Truncate(time.Minute).Add(time.Minute).Sub(time.Now())
        logrus.Debug("Listening to channel for the next " + timeout.String() + " seconds")

        select {

        // Timeout a second before the next minute and break out of channel loop
        case <-time.After(timeout):
          logrus.Debug("No longer listing on channel")
          stop = true

        // Spawn thread on channel traffic and go back to listening
        case incomingChanComm := <-runningChanComm:

        // Spawn thread so we can get back to listening
          Running.RLock()
          go func(schedule job.JobSchedule, running job.RunningJobTracker) {

            // Send the running schedule to a requestor over the same channel
            switch incomingChanComm.Signal {
            case "scheduleGetList":
              runningChanComm <- api.ChanComm{RunningSchedule: schedule, Signal: "scheduleGetList"}
            case "runningjobGetList":
              runningChanComm <- api.ChanComm{RunningJobs: running, Signal: "runningjobGetList"}
            case "replaceRunningSchedule":
              // Replace the running schedule with that of the requestor
              err := incomingChanComm.RunningSchedule.CheckConfig()
              if err != nil {
                logrus.Error(err)
              }

              logrus.Debug("Schedule Refreshed")
              if isUnitTest != true {
                incomingChanComm.RunningSchedule.WriteJobConfig(jobConfig)
              }
              schedule = incomingChanComm.RunningSchedule
            case "shutdown":
              logrus.Info("Recieved shutdown command.  Goodbye...")
              return
            default:
              runningChanComm <- api.ChanComm{Error: errors.New("API ChanComm signal unknown or deprecated: " + incomingChanComm.Signal)}
            }
          }(schedule, Running)
          Running.RUnlock()
        }
      }
    }
  }
}
