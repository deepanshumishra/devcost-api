package models

type ProjectCost struct {
	Project  string `json:"project"`
	Cost     string `json:"cost"`
	Currency string `json:"currency"`
}