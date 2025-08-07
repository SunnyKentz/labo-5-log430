package api

import (
	"caisse-app-scaled/caisse_app_scaled/magasin/caissier"
	. "caisse-app-scaled/caisse_app_scaled/utils"
	"time"

	_ "caisse-app-scaled/docs/swagger/magasin"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	swagger "github.com/gofiber/swagger"
	"github.com/prometheus/client_golang/prometheus"
)

// @title Magasin API
// @version 1.0
// @description This is the API for the Magasin service.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email fiber@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost/magasin
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @securityDefinitions.apikey Magasin
// @in header
// @name C-Mag
// @securityDefinitions.apikey Caisse
// @in header
// @name C-Caisse
func newDataApi() *fiber.App {
	api := fiber.New(fiber.Config{})
	api.Get("/swagger/*", swagger.HandlerDefault)
	prometheus := fiberprometheus.NewWithRegistry(prometheus.DefaultRegisterer, "httpservice", "magasin", "http", nil)
	prometheus.RegisterAt(api, "/metrics")
	api.Use(prometheus.Middleware)
	api.Post("/login", loginHandler)
	api.Get("/produits", cacheMiddleware, authMiddleWare, getProductsHandler)
	api.Get("/produits/:nom", cacheMiddleware, authMiddleWare, findProductHandler)
	api.Post("/cart/:id", authMiddleWare, addToCartHandler)
	api.Get("/cart", authMiddleWare, getCartItemsHandler)
	api.Delete("/cart/:id", authMiddleWare, removeFromCartHandler)
	api.Post("/vendre", authMiddleWare, makeSaleHandler)
	api.Get("/transactions", cacheMiddleware, authMiddleWare, getTransactionsHandler)
	api.Post("/rembourser/:id", authMiddleWare, refundTransactionHandler)
	api.Post("/produit/:id", authMiddleWare, requestRestockHandler)
	api.Put("/produit/:id", updateProductHandler)
	api.Put("/produit/:id/:qt", restockProductHandler)
	return api
}

// @Summary Login
// @Description Authenticate an employee with the Mere system
// @Tags Authentication
// @Accept application/json
// @Produce application/json
// @Param body body object{username=string,password=string,caisse=string,magasin=string} true "Login Credentials"
// @Success 200 {object} object{token=string} "JWT Token"
// @Failure 400 {object} models.ApiError "Bad Request"
// @Failure 403 {object} models.ApiError "Forbidden"
// @Router /api/v1/login [post]
func loginHandler(c *fiber.Ctx) error {
	var requestBody struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Caisse   string `json:"caisse"`
		Magasin  string `json:"magasin"`
	}

	if err := c.BodyParser(&requestBody); err != nil {
		return GetApiError(c, SYNTAX_ERR("body", "json"), http.StatusBadRequest)
	}
	employe := requestBody.Username
	pw := requestBody.Password
	caisse := requestBody.Caisse
	magasin := requestBody.Magasin
	jwt, err := Login(employe, pw, "any")
	if err != nil {
		return GetApiError(c, "Failed to login, verifier la caisse, le nom, le mdp", http.StatusForbidden)
	}
	if !caissier.InitialiserPOS(employe, caisse, magasin) {
		return GetApiError(c, FAILURE_ERR(errors.New("echec d'ouverture de la caisse")), http.StatusInternalServerError)
	}
	return c.Status(200).JSON(fiber.Map{
		"token": jwt,
	})
}

// @Summary Get Products
// @Description Get all products
// @Tags products
// @Produce json
// @Security BearerAuth
// @Security Magasin
// @Security Caisse
// @Success 200 {array} models.Produit
// @Failure 500 {object} models.ApiError
// @Router /api/v1/produits [get]
func getProductsHandler(c *fiber.Ctx) error {
	produits, err := caissier.AfficherProduits()
	if err != nil {
		return GetApiError(c, FAILURE_ERR(err), http.StatusInternalServerError)
	}
	path := c.Method() + " " + c.Path()
	now := time.Now()
	cache["t - "+path] = now
	cache[path] = produits
	return c.JSON(produits)
}

// @Summary Find Product
// @Description Find a product by name
// @Tags products
// @Produce json
// @Security BearerAuth
// @Security Magasin
// @Security Caisse
// @Param nom path string true "Product Name"
// @Success 200 {array} models.Produit
// @Failure 400 {object} models.ApiError
// @Router /api/v1/produits/{nom} [get]
func findProductHandler(c *fiber.Ctx) error {
	nomParam := c.Params("nom")
	nom, err := url.QueryUnescape(nomParam)
	if err != nil {
		return GetApiError(c, SYNTAX_ERR("nom", nomParam), http.StatusBadRequest)
	}
	produits, err := caissier.TrouverProduit(nom)
	if err != nil {
		return GetApiError(c, NOTFOUND_ERR("nom", nom), http.StatusNotFound)
	}
	path := c.Method() + " " + c.Path()
	now := time.Now()
	cache["t - "+path] = now
	cache[path] = produits
	return c.JSON(produits)
}

// @Summary Add to Cart
// @Description Add a product to the cart
// @Tags cart
// @Produce json
// @Security BearerAuth
// @Security Magasin
// @Security Caisse
// @Param id path int true "Product ID"
// @Success 200 {object} models.ApiSuccess
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Router /api/v1/cart/{id} [post]
func addToCartHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return GetApiError(c, SYNTAX_ERR("id", idParam), http.StatusBadRequest)
	}
	err = caissier.AjouterALaCart(id)
	if err != nil {
		return GetApiError(c, NOTFOUND_ERR("id", id), http.StatusNotFound)
	}

	return GetApiSuccess(cache, c, 200)
}

// @Summary Get Cart Items
// @Description Get all items in the cart
// @Tags cart
// @Produce json
// @Security BearerAuth
// @Security Magasin
// @Security Caisse
// @Success 200 {array} models.Produit
// @Failure 500 {object} models.ApiError
// @Router /api/v1/cart [get]
func getCartItemsHandler(c *fiber.Ctx) error {
	items, err := caissier.GetCartItems()
	if err != nil {
		return GetApiError(c, FAILURE_ERR(err), http.StatusInternalServerError)
	}
	return c.JSON(items)
}

// @Summary Remove from Cart
// @Description Remove a product from the cart
// @Tags cart
// @Produce json
// @Security BearerAuth
// @Security Magasin
// @Security Caisse
// @Param id path int true "Product ID"
// @Success 200 {object} models.ApiSuccess
// @Failure 400 {object} models.ApiError
// @Router /api/v1/cart/{id} [delete]
func removeFromCartHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return GetApiError(c, SYNTAX_ERR("id", idParam), http.StatusBadRequest)
	}
	caissier.RetirerDeLaCart(id)

	return GetApiSuccess(cache, c, 200)
}

// @Summary Make a Sale
// @Description Finalize the sale of items in the cart
// @Tags sales
// @Produce json
// @Security BearerAuth
// @Security Magasin
// @Security Caisse
// @Success 200 {object} models.ApiSuccess
// @Failure 500 {object} models.ApiError
// @Router /api/v1/vendre [post]
func makeSaleHandler(c *fiber.Ctx) error {
	err := caissier.FaireUneVente()
	caissier.ViderLaCart()
	if err != nil {
		return GetApiError(c, FAILURE_ERR(err), http.StatusInternalServerError)
	}

	return GetApiSuccess(cache, c, 200)
}

// @Summary Get Transactions
// @Description Get all transactions
// @Tags sales
// @Produce json
// @Security BearerAuth
// @Security Magasin
// @Security Caisse
// @Success 200 {array} models.Transaction
// @Failure 500 {object} models.ApiError
// @Router /api/v1/transactions [get]
func getTransactionsHandler(c *fiber.Ctx) error {
	transactions := caissier.AfficherTransactions()
	if transactions == nil {
		return GetApiError(c, FAILURE_ERR(errors.New("transaction list was null")), http.StatusInternalServerError)
	}
	path := c.Method() + " " + c.Path()
	now := time.Now()
	cache["t - "+path] = now
	cache[path] = transactions
	return c.JSON(transactions)
}

// @Summary Refund Transaction
// @Description Refund a transaction by its ID
// @Tags sales
// @Produce json
// @Security BearerAuth
// @Security Magasin
// @Security Caisse
// @Param id path int true "Transaction ID"
// @Success 200 {object} models.ApiSuccess
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Router /api/v1/rembourser/{id} [post]
func refundTransactionHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return GetApiError(c, SYNTAX_ERR("id", idParam), http.StatusBadRequest)
	}

	err = caissier.FaireUnRetour(id)
	if err != nil {
		return GetApiError(c, NOTFOUND_ERR("id", id), http.StatusNotFound)
	}
	return GetApiSuccess(cache, c, 200)
}

// @Summary Request Restock
// @Description Request to restock a product
// @Tags products
// @Produce json
// @Security BearerAuth
// @Security Magasin
// @Security Caisse
// @Param id path int true "Product ID"
// @Success 200 {object} models.ApiSuccess
// @Failure 400 {object} models.ApiError
// @Router /api/v1/produit/{id} [post]
func requestRestockHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return GetApiError(c, SYNTAX_ERR("id", idParam), http.StatusBadRequest)
	}
	caissier.DemmandeReapprovisionner(id)
	return GetApiSuccess(cache, c, 200)
}

// @Summary Update Product
// @Description Update a product's details
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param product body object{nom=string,prix=float64,description=string} true "Product Data"
// @Success 200 {object} models.ApiSuccess
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /api/v1/produit/{id} [put]
func updateProductHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return GetApiError(c, SYNTAX_ERR("id", idParam), http.StatusBadRequest)
	}
	var body struct {
		Nom         string  `json:"nom"`
		Prix        float64 `json:"prix"`
		Description string  `json:"description"`
	}

	if err := c.BodyParser(&body); err != nil {
		return GetApiError(c, SYNTAX_ERR("body", "json"), http.StatusBadRequest)
	}

	err = caissier.MiseAJourProduit(id, body.Nom, body.Prix, body.Description)
	if err != nil {
		return GetApiError(c, FAILURE_ERR(err), http.StatusInternalServerError)
	}
	return GetApiSuccess(cache, c, 200)
}

// @Summary Restock Product
// @Description Restock a product with a given quantity
// @Tags products
// @Produce json
// @Param id path int true "Product ID"
// @Param qt path int true "Quantity"
// @Success 200 {object} models.ApiSuccess
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /api/v1/produit/{id}/{qt} [put]
func restockProductHandler(c *fiber.Ctx) error {
	qtParam := c.Params("qt")
	qt, err := strconv.Atoi(qtParam)
	if err != nil {
		return GetApiError(c, SYNTAX_ERR("qt", qtParam), http.StatusBadRequest)
	}
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return GetApiError(c, SYNTAX_ERR("id", idParam), http.StatusBadRequest)
	}
	err = caissier.Reapprovisionner(id, qt)
	if err != nil {
		return GetApiError(c, FAILURE_ERR(err), http.StatusInternalServerError)
	}
	return GetApiSuccess(cache, c, 200)
}
