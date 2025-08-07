package caissier

import (
	"bytes"
	"caisse-app-scaled/caisse_app_scaled/logger"
	"caisse-app-scaled/caisse_app_scaled/magasin/db"
	"caisse-app-scaled/caisse_app_scaled/models"
	. "caisse-app-scaled/caisse_app_scaled/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type caissier struct {
	Nom     string
	Magasin string
	cart    []models.Produit
}

var instance *caissier
var once sync.Once
var Host string

func InitialiserPOS(nom string, nomCaisse string, magasin string) bool {
	logger.Init(nom)
	once.Do(func() {
		db.Init()
	})

	db.SetupLog()
	available := db.GetCaissier(nomCaisse)
	if !available {
		log.Print("Erreur: la Caisse n'existe pas ou est occupé :" + nomCaisse)
		return false
	}
	nom = nomCaisse
	if nom == "New Caisse" {
		logger.Error("Erreur aucune caisse disponible")
		return false
	}

	// Occuper la caisse
	if err := db.OccuperCaisse(nom); err != nil {
		logger.Error("Erreur lors de l'occupation de la caisse: " + err.Error())
		return false
	}

	instance = &caissier{
		Nom:     nom,
		Magasin: magasin,
		cart:    make([]models.Produit, 0),
	}
	return true
}
func Nom() (string, error) {
	if instance != nil {

		return instance.Nom, nil
	}
	return "", fmt.Errorf("no instances")
}

func FermerPOS() {
	if instance != nil {
		err := db.LibererCaisse(instance.Nom)
		Errnotnil(err)
	}
}

func AfficherProduits() ([]models.Produit, error) {
	return db.ListProduit()
}

func AfficherTransactions() []models.Transaction {
	//recuperer les transactions

	resp, err := http.Get(API_MERE() + "/api/v1/transactions")
	if err != nil {
		logger.Error("Erreur lors de la requête: " + err.Error())
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Erreur lors de la requête: " + resp.Status)
		return nil
	}

	var transactions []models.Transaction
	if err := json.NewDecoder(resp.Body).Decode(&transactions); err != nil {
		logger.Error("Erreur lors du décodage: " + err.Error())
		return nil
	}
	return transactions
}

func TrouverProduit(nomPartiel string) ([]models.Produit, error) {
	return db.GetProduitsParNomWildcard(nomPartiel)
}

func AjouterALaCart(produitID int) error {
	produit, err := db.GetProduitParID(produitID)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	if produit.Quantite <= QuantiteDansLaCart(produitID) {
		return errors.New("produit insuffisant")
	}
	instance.cart = append(instance.cart, *produit)

	return nil
}
func GetCartItems() ([]models.Produit, error) {
	if instance == nil {
		return nil, fmt.Errorf("POS not initialized")
	}
	return instance.cart, nil
}

func TotalDeLACart() float64 {
	total := 0.0
	for _, p := range instance.cart {
		total += p.Prix
	}
	return total
}

func RetirerDeLaCart(produitID int) {
	for i, p := range instance.cart {
		if p.ID == produitID {
			instance.cart = append(instance.cart[:i], instance.cart[i+1:]...)
			break
		}
	}
}

func ViderLaCart() {
	instance.cart = make([]models.Produit, 0)
}
func QuantiteDansLaCart(produitID int) int {
	count := 0
	for _, p := range instance.cart {
		if p.ID == produitID {
			count++
		}
	}
	return count
}

func Reapprovisionner(produitID int, quantite int) error {
	if err := db.MettreAJourQuantite(produitID, quantite); err != nil {
		return err
	}
	return nil
}

func MiseAJourProduit(produitID int, nom string, prix float64, description string) error {
	produit, err := db.GetProduitParID(produitID)
	if err != nil {
		return errors.New("produit not found")
	}
	produit.Nom = nom
	produit.Prix = prix
	produit.Description = description
	if err := db.MettreAJour(*produit); err != nil {
		logger.Error(err.Error())
		return err
	}
	return nil
}

func DemmandeReapprovisionner(produitID int) {
	// Send POST request
	type hostInfo struct {
		Host string `json:"host"`
	}
	host := hostInfo{
		Host: Host,
	}
	jsonData, err := json.Marshal(host)
	if err != nil {
		logger.Error("Erreur lors de la sérialisation: " + err.Error())
		return
	}
	_, err = http.Post(fmt.Sprintf(API_LOGISTIC()+"/api/v1/commande/%s/%d", instance.Magasin, produitID), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Erreur lors de l'envoi de la transaction: " + err.Error())
	}
}

func FaireUneVente() error {
	produitIDs := ""
	total := 0.0
	if len(instance.cart) < 1 {
		return errors.New("empty cart")
	}
	for i, p := range instance.cart {
		if i > 0 {
			produitIDs += ","
		}
		total += p.Prix
		produitIDs += fmt.Sprintf("%d", p.ID)
	}

	// Create transaction data
	transactionData := map[string]any{
		"date":          time.Now(),
		"caisse":        instance.Nom,
		"type":          "VENTE",
		"produit_ids":   produitIDs,
		"montant":       total,
		"magasin":       instance.Magasin,
		"deja_retourne": false,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(transactionData)
	if err != nil {
		logger.Error("Erreur lors de la sérialisation: " + err.Error())
		return err
	}

	// Send POST request
	resp, err := http.Post(API_MERE()+"/api/v1/transactions", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Erreur lors de l'envoi de la transaction: " + err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Erreur lors de la vente: " + resp.Status)
		return fmt.Errorf("erreur lors de la vente: %s", resp.Status)
	}
	t := models.Transaction{
		Caisse:     instance.Nom,
		Type:       "VENTE",
		ProduitIDs: produitIDs,
		Montant:    total,
		Date:       time.Now(),
	}
	logger.Transaction(&t, "Vente effectuée")
	err = db.MettreAJourQuantiteParTrnasaction(&t, instance.Magasin)
	Errnotnil(err)
	instance.cart = make([]models.Produit, 0)
	return nil

}

func GetTransactionByID(transactionID int) (models.Transaction, error) {
	resp, err := http.Get(fmt.Sprintf(API_MERE()+"/api/v1/transactions/%d", transactionID))
	if err != nil {
		return models.Transaction{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.Transaction{}, fmt.Errorf("transaction not found: %s", resp.Status)
	}

	var transaction models.Transaction
	if err := json.NewDecoder(resp.Body).Decode(&transaction); err != nil {
		return models.Transaction{}, err
	}
	return transaction, nil
}

func FaireUnRetour(transactionID int) error {
	transaction, err := GetTransactionByID(transactionID)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	if transaction.Type == "RETOUR" || transaction.Deja_retourne {
		return fmt.Errorf("cette transaction ne peut etre retourne")
	}

	// Make HTTP request to delete transaction
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf(API_MERE()+"/api/v1/transactions/%d", transactionID), nil)
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Erreur lors du retour: " + err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Erreur lors du retour: " + resp.Status)
		return fmt.Errorf("erreur lors du retour: %s", resp.Status)
	}

	// Create return transaction
	returnTransaction := &models.Transaction{
		Caisse:     instance.Nom,
		Type:       "RETOUR",
		ProduitIDs: transaction.ProduitIDs,
		Montant:    -transaction.Montant,
		Date:       time.Now(),
	}
	err = db.MettreAJourQuantiteParTrnasaction(returnTransaction, instance.Magasin)
	Errnotnil(err)
	// Log the return transaction
	logger.Transaction(returnTransaction, "Retour effectué")

	return nil
}
