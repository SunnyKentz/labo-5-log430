package api

import (
	"bytes"
	"caisse-app-scaled/caisse_app_scaled/logger"
	"caisse-app-scaled/caisse_app_scaled/maison_mere/mere"
	"caisse-app-scaled/caisse_app_scaled/models"
	. "caisse-app-scaled/caisse_app_scaled/utils"
	"encoding/json"

	_ "caisse-app-scaled/docs/swagger/mere"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	swagger "github.com/gofiber/swagger"
	"github.com/prometheus/client_golang/prometheus"
)

type Body struct {
	Message string `json:"message"`
	Host    string `json:"host"`
}

// @title Maison mere API
// @version 1.0
// @description This is the API for the maison mere.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email fiber@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost/mere
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @BasePath /api/v1
func newDataApi() *fiber.App {
	// api mount
	api := fiber.New(fiber.Config{})
	api.Get("/swagger/*", swagger.HandlerDefault) // default
	prometheus := fiberprometheus.NewWithRegistry(prometheus.DefaultRegisterer, "httpservice", "mere", "http", nil)
	prometheus.RegisterAt(api, "/metrics")
	api.Use(prometheus.Middleware)
	api.Post("/merelogin", mereloginHandler)
	api.Post("/register", registerHandler)
	api.Post("/notify", notifyHandler)
	api.Post("/subscribe", subscribeHandler)
	api.Get("/alerts", authMiddleWare, alertsHandler)
	api.Get("/transactions", cacheMiddleware, getTransactionsHandler)
	api.Get("/transactions/:id", cacheMiddleware, getTransactionByIDHandler)
	api.Post("/transactions", createTransactionHandler)
	api.Delete("/transactions/:id", deleteTransactionHandler)
	api.Get("/magasins", cacheMiddleware, authMiddleWare, getMagasinsHandler)
	api.Get("/analytics/:mag", authMiddleWare, getAnalyticsHandler)
	api.Get("/raport", cacheMiddleware, authMiddleWare, getRaportHandler)
	api.Get("/produits/:nom", cacheMiddleware, authMiddleWare, findProductHandler)
	api.Put("/produit", authMiddleWare, updateProductHandler)

	return api
}

// @Summary Mere Login
// @Description Authenticates an employee with the Mere system
// @Tags auth
// @Accept json
// @Produce json
// @Param body body object{username=string,password=string} true "Login Credentials"
// @Success 200 {object} object{token=string}
// @Failure 400 {object} models.ApiError
// @Failure 403 {object} models.ApiError
// @Router /api/v1/merelogin [post]
func mereloginHandler(c *fiber.Ctx) error {
	var requestBody struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&requestBody); err != nil {
		return GetApiError(c, SYNTAX_ERR("body", "json"), http.StatusBadRequest)
	}
	employe := requestBody.Username
	pw := requestBody.Password
	jwt, err := Login(employe, pw, "manager")
	if err != nil {
		return GetApiError(c, "Failed to login: "+err.Error(), http.StatusForbidden)
	}
	return c.Status(200).JSON(fiber.Map{
		"token": jwt,
	})
}

// @Summary Mere register
// @Description enregistre un nouveau client
// @Tags auth
// @Accept json
// @Produce json
// @Param body body object{username=string,password=string} true "Credentials"
// @Success 200 {object} object{token=string}
// @Failure 400 {object} models.ApiError
// @Failure 403 {object} models.ApiError
// @Router /api/v1/merelogin [post]
func registerHandler(c *fiber.Ctx) error {
	var requestBody struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&requestBody); err != nil {
		return GetApiError(c, SYNTAX_ERR("body", "json"), http.StatusBadRequest)
	}
	user := requestBody.Username

	// Convert to JSON
	jsonData, err := json.Marshal(map[string]any{"username": user})
	if err != nil {
		logger.Error("Erreur lors de la sérialisation: " + err.Error())
		return GetApiError(c, FAILURE_ERR(err), http.StatusBadRequest)
	}
	// Send POST request
	resp, err := http.Post(API_AUTH()+"/api/v1/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return GetApiError(c, FAILURE_ERR(errors.New("User not created")), http.StatusBadRequest)
	}
	pw := requestBody.Password
	jwt, err := Login(user, pw, "manager")
	if err != nil {
		return GetApiError(c, "Failed to login: "+err.Error(), http.StatusForbidden)
	}
	return c.Status(200).JSON(fiber.Map{
		"token": jwt,
	})
}

// @Summary Notify
// @Description Receives a notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param body body Body true "Notification Body"
// @Success 200 {object} models.ApiSuccess
// @Failure 400 {object} models.ApiError
// @Router /api/v1/notify [post]
func notifyHandler(c *fiber.Ctx) error {
	var b Body
	if err := c.BodyParser(&b); err != nil {

		return GetApiError(c, SYNTAX_ERR("body", "json"), http.StatusBadRequest)
	}
	println(b.Message)
	mere.Notifications = append([]string{b.Message}, mere.Notifications...) //enqueue

	return GetApiSuccess(cache, c, 200)
}

// @Summary Subscribe
// @Description Subscribes a new store
// @Tags notifications
// @Accept json
// @Produce json
// @Param body body Body true "Subscription Body"
// @Success 200 {object} models.ApiSuccess
// @Failure 400 {object} models.ApiError
// @Router /api/v1/subscribe [post]
func subscribeHandler(c *fiber.Ctx) error {
	var b Body
	if err := c.BodyParser(&b); err != nil {

		return GetApiError(c, SYNTAX_ERR("body", "json"), http.StatusBadRequest)
	}
	if !slices.Contains(mere.Magasins, b.Host) {
		mere.Magasins = append(mere.Magasins, b.Host)
	}

	return GetApiSuccess(cache, c, 200)
}

// @Summary Get Alerts
// @Description Retrieves all notifications
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Body
// @Router /api/v1/alerts [get]
func alertsHandler(c *fiber.Ctx) error {
	var notifs []Body = []Body{}
	for _, v := range mere.Notifications {
		notifs = append(notifs, Body{Message: v})
	}

	return c.Status(http.StatusOK).JSON(notifs)
}

// @Summary Get Transactions
// @Description Retrieves all transactions
// @Tags transactions
// @Produce json
// @Success 200 {array} models.Transaction
// @Failure 400 {object} models.ApiError
// @Router /api/v1/transactions [get]
func getTransactionsHandler(c *fiber.Ctx) error {
	if transactions := mere.AfficherTransactions(); transactions != nil {
		path := c.Method() + " " + c.Path()
		now := time.Now()
		cache["t - "+path] = now
		cache[path] = transactions
		return c.JSON(transactions)
	}
	logger.Error("transactions is nil")
	return GetApiError(c, FAILURE_ERR(errors.New("transactions is null")), http.StatusBadRequest)
}

// @Summary Get Transaction by ID
// @Description Retrieves a single transaction by its ID
// @Tags transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} models.Transaction
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Router /api/v1/transactions/{id} [get]
func getTransactionByIDHandler(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return GetApiError(c, SYNTAX_ERR("id", id), http.StatusBadRequest)
	}

	transaction, err := mere.AfficherUneTransactions(id)
	if err != nil {
		return GetApiError(c, NOTFOUND_ERR("id", id), http.StatusNotFound)
	}
	path := c.Method() + " " + c.Path()
	now := time.Now()
	cache["t - "+path] = now
	cache[path] = transaction
	return c.JSON(transaction)
}

// @Summary Create Transaction
// @Description Creates a new transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Param transaction body models.Transaction true "Transaction Object"
// @Success 200 {object} models.Transaction
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /api/v1/transactions [post]
func createTransactionHandler(c *fiber.Ctx) error {
	var transaction models.Transaction
	if err := c.BodyParser(&transaction); err == nil {
		if err = mere.FaireUneVente(transaction); err == nil {
			return c.Status(http.StatusOK).JSON(transaction)
		}
		return GetApiError(c, FAILURE_ERR(err), http.StatusInternalServerError)
	}
	return GetApiError(c, SYNTAX_ERR("body", "json"), http.StatusBadRequest)
}

// @Summary Delete Transaction
// @Description Deletes a transaction by its ID (refund)
// @Tags transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} models.ApiSuccess
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /api/v1/transactions/{id} [delete]
func deleteTransactionHandler(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return GetApiError(c, SYNTAX_ERR("id", id), http.StatusBadRequest)
	}
	if err := mere.FaireUnRetour(id); err != nil {
		return GetApiError(c, FAILURE_ERR(err), http.StatusInternalServerError)
	}

	return GetApiSuccess(cache, c, 200)
}

// @Summary Get Magasins
// @Description Retrieves all stores
// @Tags magasins
// @Produce json
// @Security BearerAuth
// @Success 200 {array} string
// @Router /api/v1/magasins [get]
func getMagasinsHandler(c *fiber.Ctx) error {
	magasins := mere.AfficherTousLesMagasins()
	path := c.Method() + " " + c.Path()
	now := time.Now()
	cache["t - "+path] = now
	cache[path] = magasins
	return c.JSON(magasins)
}

// @Summary Get Analytics
// @Description Retrieves analytics for a store
// @Tags analytics
// @Produce json
// @Security BearerAuth
// @Security BearerAuth
// @Param mag path string true "Store Name (or 'tout' for all)"
// @Success 200 {object} object
// @Failure 400 {object} models.ApiError
// @Router /api/v1/analytics/{mag} [get]
func getAnalyticsHandler(c *fiber.Ctx) error {
	var data struct {
		Magasin string `json:"magasin"`
		Vente   struct {
			Total float64     `json:"total"`
			Dates []time.Time `json:"dates"`
			Sales []float64   `json:"sales"`
		} `json:"vente"`
	}
	mag, err := url.QueryUnescape(c.Params("mag"))
	if err != nil {
		return GetApiError(c, SYNTAX_ERR("mag", mag), http.StatusBadRequest)
	}
	if mag == "tout" {
		total, date, sales := mere.AnalyticsVentetout()
		data.Vente.Total = total
		data.Vente.Dates = date
		data.Vente.Sales = sales
		data.Magasin = "tout"
	} else {
		total, date, sales := mere.AnalyticsVenteMagasin(mag)
		data.Vente.Total = total
		data.Vente.Dates = date
		data.Vente.Sales = sales
		data.Magasin = mag
	}
	return c.JSON(data)
}

// @Summary Get Report
// @Description Retrieves a report for all stores
// @Tags analytics
// @Produce json
// @Security BearerAuth
// @Success 200 {array} object
// @Router /api/v1/raport [get]
func getRaportHandler(c *fiber.Ctx) error {
	type data struct {
		Magasin string         `json:"magasin"`
		Total   float64        `json:"total"`
		Best5   []string       `json:"best5"`
		Stock5  map[string]int `json:"stock5"`
	}
	var datas []data = []data{}
	mags := mere.AfficherTousLesMagasins()
	for _, mag := range mags {
		total, best5, stock5 := mere.GetRaportMagasin(mag)
		datas = append(datas, data{
			Magasin: mag,
			Total:   total,
			Best5:   best5,
			Stock5:  stock5,
		})
	}
	path := c.Method() + " " + c.Path()
	now := time.Now()
	cache["t - "+path] = now
	cache[path] = datas
	return c.JSON(datas)
}

// @Summary Find Product
// @Description Finds a product by name
// @Tags produits
// @Produce json
// @Security BearerAuth
// @Param nom path string true "Product Name"
// @Success 200 {array} models.Produit
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Router /api/v1/produits/{nom} [get]
func findProductHandler(c *fiber.Ctx) error {
	nom, err := url.QueryUnescape(c.Params("nom"))
	if err != nil {
		return GetApiError(c, SYNTAX_ERR("nom", nom), http.StatusBadRequest)
	}
	produits, err := mere.TrouverProduit(nom)
	if err != nil {
		return GetApiError(c, NOTFOUND_ERR("nom", nom), http.StatusNotFound)
	}
	path := c.Method() + " " + c.Path()
	now := time.Now()
	cache["t - "+path] = now
	cache[path] = produits
	return c.JSON(produits)
}

// @Summary Update Product
// @Description Updates a product
// @Tags produits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product body object{productId=int,nom=string,prix=float64,description=string} true "Product Update Data"
// @Success 200 {object} models.ApiSuccess
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /api/v1/produit [put]
func updateProductHandler(c *fiber.Ctx) error {
	var data struct {
		ID          int     `json:"productId"`
		Nom         string  `json:"nom"`
		Prix        float64 `json:"prix"`
		Description string  `json:"description"`
	}
	if err := c.BodyParser(&data); err != nil {
		return GetApiError(c, SYNTAX_ERR("body", "json"), http.StatusBadRequest)
	}

	err := mere.MiseAJourProduit(data.ID, data.Nom, data.Prix, data.Description)
	if err != nil {
		return GetApiError(c, FAILURE_ERR(err), http.StatusInternalServerError)
	}
	return GetApiSuccess(cache, c, 200)
}
