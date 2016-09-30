package conf

// DaemonConfig - Organise attributes that can be modified by command line parameter of by API call
type DaemonConfig struct {
  LogLevel      int
  Port          int
  SocketPath    string
  JobConfigPath string
}

var Attr = DaemonConfig{}

func init() {

  // Set default configurations
  Attr.LogLevel = 0
  Attr.Port = 51515
  Attr.SocketPath = "omicrond.sock"
  Attr.JobConfigPath = "sample/samplejobConf.toml"

}