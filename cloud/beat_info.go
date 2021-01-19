package cloud

type BeatInfo struct {
	Port        int     `json:"port"`
	Ip          string  `json:"ip"`
	ServiceName string  `json:"serviceName"`
	Weight      float64 `json:"weight"`
	Ephemeral   bool    `json:"ephemeral"`
}
