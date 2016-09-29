package job

import (
  "fmt"
  "time"
  "github.com/BurntSushi/toml"
)

type JobConfig struct {
  Title     string    // The name of the job.  Used in logging
  Command   string    // String to be run on the system
  Interval  durationObj  // Time to wait before running the job again
  StartTime timeObj // Job will not run from the start of the day until this time is reached
  EndTime   timeObj // Job will not run after this time and until the end of the day
  GroupName string    // Used to relate jobs and in logging *unused*
}

type JobHandler struct {
  Jobs []JobConfig
}

func (h *JobHandler) ParseJobConfig(confFile string) {

	if _, err := toml.DecodeFile(confFile, &h); err != nil {
		fmt.Println(err)
		return
	}
}

type durationObj struct {
  time.Duration
}
type timeObj struct {
  time.Time
}

func (d *durationObj) UnmarshalText(text []byte) error {
  var err error
  d.Duration, err = time.ParseDuration(string(text))
  return err
}
func (t *timeObj) UnmarshalText(text []byte) error {
  var err error
  t.Time, err = time.Parse(time.Kitchen,string(text))
  return err
}
