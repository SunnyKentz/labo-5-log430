package models

type Employe struct {
	ID   int    `json:"id"`
	Nom  string `json:"nom"`
	Role string `json:"role"` // "caissier", "gerant", etc.
}
