package models

type Location struct {
	ID          int    `json:"id"`
	Name        string `json:"location_name"`
	Description string `json:"location_description"`
}
