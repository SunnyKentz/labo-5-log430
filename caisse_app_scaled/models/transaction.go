package models

import (
	"time"
)

type Transaction struct {
	ID            int       `json:"id"`
	Date          time.Time `json:"date"`
	Caisse        string    `json:"caisse"`
	Type          string    `json:"type"`
	Magasin       string    `json:"magasin"`
	ProduitIDs    string    `json:"produit_ids"`
	Montant       float64   `json:"montant"`
	Deja_retourne bool      `json:"deja_retourne"`
}
