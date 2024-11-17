package device

type Computer struct {
	Name string `json:"name"`
	MAC  string `json:"mac"`
}

var Devices []Computer
