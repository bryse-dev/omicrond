package conf

// DaemonConfig - Organise attributes that can be modified by command line parameter of by API call
type DaemonConfig struct {
  LogLevel      int
  Port          int
  SocketPath    string
  JobConfigPath string
  APIAddress    string
  APIPort       int
  APITimeout    int
}

var Attr = DaemonConfig{}

func init() {

  // Set default configurations
  Attr.LogLevel = 0
  Attr.Port = 51515
  Attr.SocketPath = "omicrond.sock"
  Attr.JobConfigPath = "sample/samplejobConf.toml"
  Attr.APIAddress = "127.0.0.1"
  Attr.APIPort = 12221
  Attr.APITimeout = 5

}