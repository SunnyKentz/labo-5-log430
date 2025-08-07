package mere

import (
	"bytes"
	"caisse-app-scaled/caisse_app_scaled/logger"
	"caisse-app-scaled/caisse_app_scaled/maison_mere/db"
	"caisse-app-scaled/caisse_app_scaled/models"
	. "caisse-app-scaled/caisse_app_scaled/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

var nom *string
var Notifications []string = []string{}

var Magasins []string = []string{API_LOGISTIC()}

func Nom() (string, error) {
	if nom == nil {
		return "", errors.New("no name")
	}
	return *nom, nil
}
func AfficherTousLesMagasins() []string {
	magasins := db.ListMagasin()
	return magasins
}
func AfficherTransactions() []models.Transaction {
	//recuperer les transactions
	transactions, err := db.ListTransactions()
	if err != nil {
		logger.Error("Erreur lors de la recupration des transactions: " + err.Error())
		return nil
	}
	return transactions
}

func AfficherUneTransactions(trasnID int) (models.Transaction, error) {
	//recuperer les transactions
	transaction, err := db.GetTransactionByID(trasnID)
	if err != nil {
		logger.Error("Erreur lors de la recupration de la transaction: " + err.Error())
		return models.Transaction{}, fmt.Errorf("not nil")
	}
	return transaction, nil
}

func FaireUneVente(transaction models.Transaction) error {

	if err := db.EnregistrerTransaction(&transaction); err != nil {
		logger.Error("Erreur lors de la vente: " + err.Error())
		return err
	}
	logger.Transaction(&transaction, "Vente effectuée")
	return nil
}

func FaireUnRetour(transactionID int) error {
	transaction, err := db.GetTransactionByID(transactionID)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	if transaction.Type == "RETOUR" || transaction.Deja_retourne {
		return fmt.Errorf("cette transaction ne peut etre retourne")
	}
	transaction.Montant *= -1
	transaction.Type = "RETOUR"
	transaction.Date = time.Now()
	if err := db.EnregistrerTransaction(&transaction); err != nil {
		logger.Error("Erreur lors du retour: " + err.Error())
		return err
	}
	err = db.SetTransactionToDejaRetourne(transactionID)
	Errnotnil(err)
	logger.Transaction(&transaction, "Retour effectué")
	return nil
}

func AnalyticsVentetout() (float64, []time.Time, []float64) {
	var total float64
	var dates []time.Time
	var sales []float64
	transactions, err := db.ListTransactions()
	if err != nil {
		logger.Error("Erreur lors de la récupération des transactions: " + err.Error())
		return 0, nil, nil
	}

	// Group transactions by date and calculate totals
	dateMap := make(map[time.Time]float64)
	for _, t := range transactions {
		if t.Type == "VENTE" {
			// Normalize date to start of day for grouping
			dateKey := time.Date(t.Date.Year(), t.Date.Month(), t.Date.Day(), 0, 0, 0, 0, t.Date.Location())
			dateMap[dateKey] += t.Montant
			total += t.Montant
		}
	}

	// Convert map to sorted slices
	for date, amount := range dateMap {
		dates = append(dates, date)
		sales = append(sales, amount)
	}
	return total, dates, sales
}

func AnalyticsVenteMagasin(magasin string) (float64, []time.Time, []float64) {
	var total float64
	var dates []time.Time
	var sales []float64
	transactions, err := db.ListTransactions()
	if err != nil {
		logger.Error("Erreur lors de la récupération des transactions: " + err.Error())
		return 0, nil, nil
	}

	// Group transactions by date and calculate totals for specific magasin
	dateMap := make(map[time.Time]float64)
	for _, t := range transactions {
		if t.Type == "VENTE" && t.Magasin == magasin {
			// Normalize date to start of day for grouping
			dateKey := time.Date(t.Date.Year(), t.Date.Month(), t.Date.Day(), 0, 0, 0, 0, t.Date.Location())
			dateMap[dateKey] += t.Montant
			total += t.Montant
		}
	}

	// Convert map to sorted slices
	for date, amount := range dateMap {
		dates = append(dates, date)
		sales = append(sales, amount)
	}
	return total, dates, sales
}

func GetRaportMagasin(mag string) (float64, []string, map[string]int) {
	total := 0.0
	best5 := []string{}
	stock5 := make(map[string]int, 0)

	// Get all transactions
	transactions, err := db.ListTransactions()
	if err != nil {
		logger.Error("Erreur lors de la récupération des transactions: " + err.Error())
		return 0, nil, nil
	}

	// Count product occurrences and calculate total sales for the magasin
	productCount := make(map[int]int)
	for _, t := range transactions {
		if t.Magasin == mag && t.Type == "VENTE" {
			total += t.Montant

			// Parse product IDs from transaction
			ids := strings.Split(t.ProduitIDs, ",")
			for _, idStr := range ids {
				if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
					productCount[id]++
				}
			}
		}
	}

	// Find top 5 most frequent products
	type productFreq struct {
		id   int
		freq int
	}

	var freqList []productFreq
	for id, freq := range productCount {
		freqList = append(freqList, productFreq{id: id, freq: freq})
	}

	// Sort by frequency (descending)
	sort.Slice(freqList, func(i, j int) bool {
		return freqList[i].freq > freqList[j].freq
	})

	// Get top 5 product IDs
	top5IDs := []int{}
	for i := 0; i < 5 && i < len(freqList); i++ {
		top5IDs = append(top5IDs, freqList[i].id)
	}

	// Get product names and stock for top 5 products
	for _, id := range top5IDs {
		if produit, err := getProduitParID(id); err == nil {
			best5 = append(best5, produit.Nom)
			stock5[produit.Nom] = produit.Quantite
		} else {
			logger.Error(err.Error())
		}
	}

	return total, best5, stock5
}

func getProduitParID(id int) (models.Produit, error) {
	resp, err := http.Get(API_LOGISTIC() + "/api/v1/produits/id/" + strconv.Itoa(id))
	if err != nil {
		return models.Produit{}, err
	}
	defer resp.Body.Close()

	var produits models.Produit
	if err := json.NewDecoder(resp.Body).Decode(&produits); err != nil {
		return models.Produit{}, err
	}
	return produits, nil
}
func TrouverProduit(nom string) ([]models.Produit, error) {
	resp, err := http.Get(API_LOGISTIC() + "/api/v1/produits/" + nom)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var produits []models.Produit
	if err := json.NewDecoder(resp.Body).Decode(&produits); err != nil {
		return nil, err
	}
	return produits, nil
}

func MiseAJourProduit(id int, nom string, prix float64, description string) error {
	// for each mere.Magasins, post to maagsin PUT /produit/:id body : {"nom":nom,"prix":prix,"description":description}

	// Create request body
	body := map[string]interface{}{
		"nom":         nom,
		"prix":        prix,
		"description": description,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		logger.Error("Erreur lors de la sérialisation: " + err.Error())
		return errors.New("Erreur lors de la sérialisation: " + err.Error())
	}

	// Send PUT request to each magasin
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/produit/%d", API_MAGASIN(), id), bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Erreur lors de la création de la requête: " + err.Error())
		return errors.New("Erreur lors de la création de la requête: " + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Erreur lors de l'envoi de la requête: " + err.Error())
		return errors.New("Erreur lors de l'envoi de la requête: " + err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Error(fmt.Sprintf("Erreur lors de la mise à jour du produit dans %s, status: %d", API_LOGISTIC(), resp.StatusCode))
		return errors.New("erreur lors de la mise à jour du produit")
	}

	// Send PUT request to logistique
	req, err = http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/produit/%d", API_LOGISTIC(), id), bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Erreur lors de la création de la requête: " + err.Error())
		return errors.New("Erreur lors de la création de la requête: " + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	client = &http.Client{}
	resp2, err := client.Do(req)
	if err != nil {
		logger.Error("Erreur lors de l'envoi de la requête: " + err.Error())
		return errors.New("Erreur lors de l'envoi de la requête: " + err.Error())
	}
	defer resp2.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error(fmt.Sprintf("Erreur lors de la mise à jour du produit dans %s, status: %d", API_MAGASIN(), resp.StatusCode))
		return errors.New("erreur lors de la mise à jour du produit")
	}

	return nil
}
