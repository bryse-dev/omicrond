package job

import (
  "fmt"
  "os/exec"
  "bufio"
  "io"
  "os"
  "strings"
  "crypto/rand"
  "sync"
  "time"
  "net/http"
  "errors"
  "github.com/Sirupsen/logrus"
  "github.com/brysearl/omicrond/conf"
)

type RunningJobTracker struct {
  sync.RWMutex
  Jobs map[string]RunningJob
}

type RunningJob struct {
  Token     string
  Config    JobConfig
  Channel   chan ChanComm
  Exec      *exec.Cmd
  StdOut    io.ReadCloser
  StdErr    io.ReadCloser
  StartTime time.Time
}

type RunningJobTrackerAPI struct {
  Jobs []RunningJobAPI
}

type RunningJobAPI struct {
  Token       string
  StartTime   time.Time
  ElapsedTime time.Duration
  PID         int
  MemUse      int

  Config      JobConfigAPI
}

type ChanComm struct {
  Signal string
  Error  error
  Writer http.ResponseWriter
}

// MakeAPIFormat - Convert internal object into external data
func (t *RunningJobTracker)MakeAPIFormat() (RunningJobTrackerAPI, error) {

  var apiTracker RunningJobTrackerAPI
  var err error
  for jobToken, jobObj := range t.Jobs {

    // Convert the running job struct to API format
    apiRunningJob, err := jobObj.MakeAPIFormat(jobToken)
    if err != nil {
      return RunningJobTrackerAPI{}, err
    }

    // Append the job to the API tracker
    apiTracker.Jobs = append(apiTracker.Jobs, apiRunningJob)
    if err != nil {
      return RunningJobTrackerAPI{}, err
    }
  }

  return apiTracker, err

}

// MakeAPIFormat - Convert internal object into external data
func (j *RunningJob) MakeAPIFormat(jobToken string) (RunningJobAPI, error) {

  apiConf, err := j.Config.MakeAPIFormat()
  if err != nil {
    return RunningJobAPI{}, err
  }

  apiRunningJob := RunningJobAPI{
    Token:  jobToken,
    StartTime: j.StartTime,
    ElapsedTime: time.Now().Sub(j.StartTime),
    Config:  apiConf }

  return apiRunningJob, err
}

// Run - Executes command
func (r *RunningJob) Run(running *RunningJobTracker) {

  var err error

  // Make the command executable
  running.Lock()
  r.Exec = r.buildCommand()
  running.Unlock()
  if err != nil {
    logrus.Error(err)
    return
  }

  // Create handles for both stdin and stdout
  running.Lock()
  r.StdOut, err = r.Exec.StdoutPipe()
  running.Unlock()
  if err != nil {
    logrus.Error(err)
    return
  }
  running.Lock()
  r.StdErr, err = r.Exec.StderrPipe()
  running.Unlock()
  if err != nil {
    logrus.Error(err)
    return
  }

  // Attach scanners to the IO handles
  stdOutScanner := bufio.NewScanner(r.StdOut)
  stdErrScanner := bufio.NewScanner(r.StdErr)

  // Spawn goroutines to effectively tail the IO scanners
  go func(r *RunningJob) {

    // Setup logfile for STDOUT
    logPath := r.DetermineLoggingDir()
    if err := os.MkdirAll(logPath, 0755); err != nil {
      logrus.Error(err)
    }
    logFile, err := os.Create(logPath + "/stdout.txt")
    if err != nil {
      logrus.Error(err)
    }

    // Scan each line as they become available
    for stdOutScanner.Scan() {
      logrus.Debug("STDOUT | " + stdOutScanner.Text())
      logFile.WriteString(stdOutScanner.Text() + "\n")
    }
    logFile.Close()
  }(r)
  go func(r *RunningJob) {

    // Setup logfile for STDERR
    logPath := r.DetermineLoggingDir()
    if err := os.MkdirAll(logPath, 0755); err != nil {
      logrus.Error(err)
    }
    logFile, err := os.Create(logPath + "/stderr.txt")
    if err != nil {
      logrus.Error(err)
    }

    // Scan each line as they become available
    for stdErrScanner.Scan() {
      logrus.Debug("STDERR | " + stdErrScanner.Text())
      logFile.WriteString(stdErrScanner.Text() + "\n")
    }
    logFile.Close()
  }(r)

  // Open up channel to extend to API
  go r.listenOnChannel(stdOutScanner)

  // Start the command
  logrus.Info("Running [" + r.Config.Label + "]: " + strings.Join(r.Exec.Args, " "))
  err = r.Exec.Start()
  if err != nil {
    logrus.Error(err)
    return
  }

  // Wait for the command to complete
  logrus.Debug("Waiting for command to complete")
  r.Exec.Wait()
  r.Channel <- ChanComm{Signal:"end"}
  logrus.Debug("Command completed")

  return
}

// listenOnChannel - open up channel communication for API commands
func (r *RunningJob) listenOnChannel(stdOutScanner *bufio.Scanner) {
  stop := false
  for stop == false {
    comm := <-r.Channel
    switch comm.Signal {
    case "end":
      stop = true
    case "stop process":
      err := r.Exec.Process.Kill()
      if err != nil {
        r.Channel <- ChanComm{Error: errors.New("failed")}
      }
      r.Channel <- ChanComm{Signal: "success"}
    case "tail process":

      func(comm ChanComm, stdOutScanner *bufio.Scanner) {
        for stdOutScanner.Scan() {
          logrus.Debug("should see " + stdOutScanner.Text())
          //fmt.Fprintf(comm.Writer, stdOutScanner.Text())
          //bufrw.WriteString(stdOutScanner.Text())
          comm.Writer.Write(stdOutScanner.Bytes())
          f, ok := comm.Writer.(http.Flusher)
          if ok {
            logrus.Info("Flushing")
            f.Flush()
          }
          //bufrw.Flush()
        }


      }(comm, stdOutScanner)
      logrus.Info("HERE")
      r.Channel <- ChanComm{Signal: "success"}
    //}(conn, stdOutScanner, bufrw)
    default:
      r.Channel <- ChanComm{Error: errors.New("unknown command")}
    }
  }
  logrus.Debug("Stopped command channel")
}

// buildCommand - Convert string to executablte exec.Cmd type
func (r *RunningJob) buildCommand() *exec.Cmd {

  // Split on spaces
  components := strings.Split(string(r.Config.Command), " ")
  if len(components) == 0 {
    logrus.Error("Missing exec command in job configuration")
  }

  // Shift off the executable from the arguments
  executable, components := components[0], components[1:]

  // Create the exec.Cmd object and attach to JobConfig
  cmdPtr := exec.Command(executable, components...)
  return cmdPtr
}

// DetermineLoggingPath - Get the filepath to write new logs to
func (r *RunningJob) DetermineLoggingDir() string {

  return conf.Attr.LoggingPath + "/" + time.Now().Format("2006-01-02") + "/" + strings.Replace(r.Config.Label, " ", "_", -1) + "/" + r.Token
}

func CreateRunToken() string {
  b := make([]byte, 8)
  rand.Read(b)
  return fmt.Sprintf("%x", b)
}
