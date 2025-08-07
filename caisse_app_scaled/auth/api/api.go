package api

import (
	"caisse-app-scaled/caisse_app_scaled/auth/authData"
	"caisse-app-scaled/caisse_app_scaled/logger"
	"caisse-app-scaled/caisse_app_scaled/models"
	. "caisse-app-scaled/caisse_app_scaled/utils"
	"log"
	"net/http"

	//	_ "caisse-app-scaled/docs/swagger/auth"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	swagger "github.com/gofiber/swagger"
	"github.com/prometheus/client_golang/prometheus"
)

var employeList []models.Employe = []models.Employe{}

// @title auth API
// @version 1.0
// @description This is the API for the auth.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email fiber@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost/auth
func NewDataApi() {
	// api mount
	api := fiber.New(fiber.Config{})
	api.Get("/swagger/*", swagger.HandlerDefault) // default
	prometheus := fiberprometheus.NewWithRegistry(prometheus.DefaultRegisterer, "httpservice", "auth", "http", nil)
	prometheus.RegisterAt(api, "/metrics")
	api.Use(prometheus.Middleware)
	api.Post("/login", loginHandler)
	api.Post("/validate", ValidateJWTHandler)
	api.Post("/register", registerHandler)

	employeList = append(employeList, authData.MakeEmployees("Alice", "commis"))
	employeList = append(employeList, authData.MakeEmployees("Bob", "manager"))
	employeList = append(employeList, authData.MakeEmployees("Claire", "commis"))
	employeList = append(employeList, authData.MakeEmployees("David", "commis"))
	employeList = append(employeList, authData.MakeEmployees("Eva", "manager"))

	logger.Init("Auth")
	app := fiber.New(fiber.Config{})
	app.Mount("/auth/api/v1", api)
	log.Fatal(app.Listen(":8092"))
}

// @Summary Login
// @Description Authenticates an employee
// @Tags auth
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param username formData string true "Username"
// @Param password formData string true "Password"
// @Param role formData string true "Role"
// @Success 200 {object} object{jwt=string}
// @Failure 401 {object} models.ApiError
// @Router /api/v1/login [post]
func loginHandler(c *fiber.Ctx) error {
	employe := c.FormValue("username")
	role := c.FormValue("role")
	if !authData.LoginAuth(employe, role, employeList) {
		return GetApiError(c, "failed to auth "+employe+" role:"+role, http.StatusUnauthorized)
	}
	var body struct {
		Jwt string `json:"jwt"`
	}
	body.Jwt, _ = authData.CreateJWT(employe)
	return c.JSON(body)
}

// @Summary Validate JWT
// @Description Validate a JWT token
// @Tags jwt
// @Accept json
// @Produce json
// @Param body body object{jwt=string} true "JWT Token"
// @Success 200 {object} models.ApiSuccess
// @Failure 400 {object} models.ApiError
// @Failure 401 {object} models.ApiError
// @Router /api/v1/validate [post]
func ValidateJWTHandler(c *fiber.Ctx) error {
	var body struct {
		JWT string `json:"jwt"`
	}
	if err := c.BodyParser(&body); err != nil {
		return GetApiError(c, SYNTAX_ERR("body", "json"), http.StatusBadRequest)
	}
	if authData.ValidateJWT(body.JWT) {
		return GetApiSuccess(make(map[string]any), c, 200)
	}

	return GetApiError(c, "failed to auth ", http.StatusUnauthorized)
}

// @Summary Register
// @Description add a new user
// @Tags register
// @Accept json
// @Produce json
// @Param body body object{username=string} true "username to add"
// @Success 201 {object} models.ApiSuccess
// @Failure 400 {object} models.ApiError
// @Router /api/v1/register [post]
func registerHandler(c *fiber.Ctx) error {
	var body struct {
		User string `json:"username"`
	}
	if err := c.BodyParser(&body); err != nil {
		return GetApiError(c, SYNTAX_ERR("body", "json"), http.StatusBadRequest)
	}

	employeList = append(employeList, authData.MakeEmployees(body.User, "manager"))
	return GetApiSuccess(map[string]any{}, c, http.StatusCreated)
}
