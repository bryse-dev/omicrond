package job

import (
  "fmt"
  "os/exec"
  "bufio"
  "io"
  "strings"
  "crypto/rand"
  "sync"
  "github.com/Sirupsen/logrus"
)

type RunningJobTracker struct {
  sync.RWMutex
  Jobs map[string]RunningJob
}

type RunningJob struct {
  Token   string
  Config  JobConfig
  Channel chan string
  Exec    *exec.Cmd
  StdOut  io.ReadCloser
  StdErr  io.ReadCloser
}

// Run - Executes command
func (r *RunningJob) Run() {

  var err error

  // Make the command executable
  r.Exec = r.buildCommand()
  if err != nil {
    logrus.Error(err)
    return
  }

  // Create handles for both stdin and stdout
  r.StdOut, err = r.Exec.StdoutPipe()
  if err != nil {
    logrus.Error(err)
    return
  }
  r.StdErr, err = r.Exec.StderrPipe()
  if err != nil {
    logrus.Error(err)
    return
  }

  // Attach scanners to the IO handles
  stdOutScanner := bufio.NewScanner(r.StdOut)
  stdErrScanner := bufio.NewScanner(r.StdErr)

  // Spawn goroutines to effectively tail the IO scanners
  go func() {
    for stdOutScanner.Scan() {
      logrus.Debug("STDOUT | " + stdOutScanner.Text())
    }
  }()

  go func() {
    for stdErrScanner.Scan() {
      logrus.Debug("STDERR | " + stdErrScanner.Text())
    }
  }()

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
  logrus.Debug("Command completed")

  return
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

func CreateRunToken() string {
  b := make([]byte, 8)
  rand.Read(b)
  return fmt.Sprintf("%x", b)
}
