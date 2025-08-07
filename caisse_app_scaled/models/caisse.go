package models

type Caisse struct {
	ID     int    `json:"id"`
	Nom    string `json:"nom"`
	Occupe bool   `json:"occupe"`
}
