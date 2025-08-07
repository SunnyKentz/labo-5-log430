package db

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"caisse-app-scaled/caisse_app_scaled/logger"
	"caisse-app-scaled/caisse_app_scaled/models"
	. "caisse-app-scaled/caisse_app_scaled/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	lg "gorm.io/gorm/logger"
)

type dbProxy struct {
	db       *gorm.DB
	username string
	password string
	host     string
	port     string
}

var (
	instance *dbProxy
	once     sync.Once
)

func Init() {
	once.Do(func() {
		instance = &dbProxy{
			username: os.Getenv("DB_USER"),
			password: os.Getenv("DB_PASSWORD"),
			port:     os.Getenv("DB_PORT"),
			host:     os.Getenv("GATEWAY"), //172.17.0.1
		}
		instance.connect()
	})
}

func (d *dbProxy) connect() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		instance.host,
		instance.port, //5434
		instance.username,
		instance.password,
		"postgres",
	)

	maxRetries := 4
	retryDelay := 4 * time.Second

	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		instance.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break // Success, exit retry loop
		}

		log.Printf("Database connection attempt %d failed: %v", attempt, err)

		if attempt < maxRetries {
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		log.Fatal("Failed to connect to database after", maxRetries, "attempts:", err)
	}
}

func GetCaissier(nom string) bool {
	var c models.Caisse
	result := instance.db.Where("nom = ?", nom).First(&c)
	if result.Error != nil {
		return false
	}
	return result.Error == nil //&& !c.Occupe
}

func SetupLog() {
	instance.db.Logger = lg.New(log.New(logger.GetFile(), "\r\n", log.LstdFlags), lg.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  lg.Warn,
		IgnoreRecordNotFoundError: false,
		Colorful:                  false,
	})
}
func OccuperCaisse(nom string) error {
	return instance.db.Model(&models.Caisse{}).Where("nom = ?", nom).Update("occupe", true).Error
}

func LibererCaisse(nom string) error {
	return instance.db.Model(&models.Caisse{}).Where("nom = ?", nom).Update("occupe", false).Error
}

func ListProduit() ([]models.Produit, error) {
	var produits []models.Produit
	err := instance.db.Find(&produits).Error
	return produits, err
}

func ListTransactions() ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := instance.db.Order("date desc").Find(&transactions).Error
	return transactions, err
}

func GetProduitParID(id int) (*models.Produit, error) {
	var p models.Produit
	err := instance.db.First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func GetProduitsParNomWildcard(nomWildcard string) ([]models.Produit, error) {
	var produits []models.Produit
	err := instance.db.Where("nom ILIKE ?", "%"+nomWildcard+"%").Find(&produits).Error
	return produits, err
}

func SetTransactionToDejaRetourne(transactionID int) error {
	err := instance.db.Model(&models.Transaction{}).Where("id = ?", transactionID).Update("deja_retourne", true).Error
	return err
}

func MettreAJourQuantite(id int, quantite int) error {
	if err := instance.db.Model(&models.Produit{}).
		Where("id = ?", id).
		UpdateColumn("quantite", gorm.Expr("quantite + ?", quantite)).Error; err != nil {
		return err
	}
	return nil
}
func MettreAJour(produit models.Produit) error {
	err := instance.db.Model(&models.Produit{}).Where("id = ?", produit.ID).Updates(produit).Error
	if err != nil {
		return err
	}
	return nil
}

func MettreAJourQuantiteParTrnasaction(t *models.Transaction, magasinNom string) error {
	// Update product quantities
	ids := strings.Split(t.ProduitIDs, ",")
	produitQMap := make(map[int]int)
	for _, v := range ids {
		id, _ := strconv.Atoi(v)
		produitQMap[id]++
	}
	for id, quantity := range produitQMap {
		if t.Type == "RETOUR" {
			quantity = -quantity
		}
		if err := MettreAJourQuantite(id, -quantity); err != nil {
			return err
		}
	}
	for id := range produitQMap {
		if produit, err := GetProduitParID(id); err == nil {
			if produit.Quantite > 10 {
				notifyMere(produit.Nom + " est en surstock dans " + magasinNom)
			} else if produit.Quantite < 5 && produit.Quantite > 0 {
				notifyMere(produit.Nom + " est en sous-stock dans " + magasinNom)
			} else {

				notifyMere(produit.Nom + " est en rupture de stock dans " + magasinNom)
			}
		}

	}
	return nil
}

func notifyMere(s string) {
	s = "{\"message\":\"" + s + "\"}"

	_, err := http.Post(API_MERE()+"/api/v1/notify", "application/json", bytes.NewBuffer([]byte(s)))
	Errnotnil(err)
}
