package cloud

type Instance struct {
	Ip          string  `json:"ip"`
	Port        int     `json:"port"`
	ServiceName string  `json:"serviceName"`
	Weight      float64 `json:"weight"`
	App         string  `json:"app"`
	Ephemeral   bool    `json:"ephemeral"`
	Healthy     bool    `json:"healthy"`
}
