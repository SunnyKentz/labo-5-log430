package models

type Produit struct {
	ID          int     `json:"id"`
	Nom         string  `json:"nom"`
	Prix        float64 `json:"prix"`
	Categorie   string  `json:"category"`
	Description string  `json:"description"`
	Quantite    int     `json:"quantite"`
}
