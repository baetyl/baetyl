// Package qpsstats qps监控实现
package qpsstats

type UpstreamCfg struct {
	Server     string `yaml:"server" json:"server"`
	RequestCnt int    `yaml:"requestCounter" json:"requestCounter"`
	Responses  struct {
		Resp1xx int `yaml:"1xx" json:"1xx"`
		Resp2xx int `yaml:"2xx" json:"2xx"`
		Resp3xx int `yaml:"3xx" json:"3xx"`
		Resp4xx int `yaml:"4xx" json:"4xx"`
		Resp5xx int `yaml:"5xx" json:"5xx"`
	}
}

type NginxStatus struct {
	HostName      string                   `yaml:"hostName" json:"hostName"`
	UpstreamZones map[string][]UpstreamCfg `yaml:"upstreamZones" json:"upstreamZones"`
}

type ServerStats struct {
	ServerName        string `yaml:"serverName" json:"serverName"`
	RequestCnt        int    `yaml:"requestCounter" json:"requestCounter"`
	RequestTotal      int    `yaml:"requestTotal" json:"requestTotal"`
	RequestCntSuccess int    `yaml:"requestCounterSuccess" json:"requestCounterSuccess"`
	RequestCntFail    int    `yaml:"requestCounterFail" json:"requestCounterFail"`
}
