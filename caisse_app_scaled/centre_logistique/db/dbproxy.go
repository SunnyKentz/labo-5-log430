package db

import (
	"caisse-app-scaled/caisse_app_scaled/logger"
	"caisse-app-scaled/caisse_app_scaled/models"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

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
			host:     os.Getenv("GATEWAY"),
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

func SetupLog() {
	instance.db.Logger = lg.New(log.New(logger.GetFile(), "\r\n", log.LstdFlags), lg.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  lg.Warn,
		IgnoreRecordNotFoundError: false,
		Colorful:                  false,
	})
}

func GetProduitsParNomWildcard(nomWildcard string) ([]models.Produit, error) {
	var produits []models.Produit
	err := instance.db.Where("nom ILIKE ?", "%"+nomWildcard+"%").Find(&produits).Error
	return produits, err
}

func GetProduitParID(id int) (*models.Produit, error) {
	var p models.Produit
	err := instance.db.First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func MettreAJour(produit models.Produit) error {
	err := instance.db.Model(&models.Produit{}).Where("id = ?", produit.ID).Updates(produit).Error
	if err != nil {
		return err
	}
	return nil
}

func MettreAJourQuantite(id int, quantite int) error {
	if err := instance.db.Model(&models.Produit{}).
		Where("id = ?", id).
		UpdateColumn("quantite", gorm.Expr("quantite + ?", quantite)).Error; err != nil {
		return err
	}
	return nil
}
