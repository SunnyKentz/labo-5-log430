package api

import (
	"caisse-app-scaled/caisse_app_scaled/centre_logistique/db"
	"caisse-app-scaled/caisse_app_scaled/logger"
	. "caisse-app-scaled/caisse_app_scaled/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func NewApp() {
	engine := html.New("./view", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Static("/logistique/js", "./commonjs")
	app.Mount("/logistique/api/v1", newDataApi())
	app.Get("/logistique/", func(c *fiber.Ctx) error {
		return c.Render("login", nil)
	})
	app.Get("/logistique/home", func(c *fiber.Ctx) error {
		return c.Render("commande", nil)
	})
	db.Init()
	logger.Init("Logistique")
	db.SetupLog()
	log.Fatal(app.Listen(":8091"))
}

func authMiddleWare(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	err := CheckLogedIn(authHeader)
	if err != nil {
		if c.Path() == "api/v1/login" {
			return c.Next()
		}
		if strings.HasPrefix(c.Path(), "/api") {
			return GetApiError(c, "this action requires authentification", http.StatusUnauthorized)
		}
		return c.Redirect("/")
	}
	return c.Next()
}

var cache = make(map[string]any)

func cacheMiddleware(c *fiber.Ctx) error {
	noCache := c.Get("no-cache", "false")
	if noCache != "false" {
		return c.Next()
	}
	path := c.Method() + " " + c.Path()
	now := time.Now()

	// Check if we have a cached value and it's not older than 30 seconds
	if t, ok := cache["t - "+path]; ok {
		if now.Sub(t.(time.Time)) < CACHE_TIME {
			if val, ok := cache[path]; ok {
				return c.JSON(val)
			}
		}
	}
	err := c.Next()
	if err != nil {
		return err
	}
	return nil
}
