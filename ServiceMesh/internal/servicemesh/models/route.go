package models

type Route struct {
	Service     string   `json:"service"`
	TargetHosts []string `json:"targetHosts"`
}
