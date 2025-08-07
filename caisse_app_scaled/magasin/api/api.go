package api

import (
	"caisse-app-scaled/caisse_app_scaled/logger"
	"caisse-app-scaled/caisse_app_scaled/magasin/caissier"
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

	app.Static("/magasin/js", "./commonjs")
	app.Mount("/magasin/api/v1", newDataApi())
	app.Get("/magasin/", func(c *fiber.Ctx) error {
		return c.Render("login", fiber.Map{
			"Title": "Login - Caisse App",
		})
	})

	app.Get("/magasin/home", func(c *fiber.Ctx) error {
		return c.Render("product", nil)
	})
	app.Get("/magasin/panier", func(c *fiber.Ctx) error {
		return c.Render("checkout", nil)
	})
	app.Get("/magasin/transactions", func(c *fiber.Ctx) error {
		return c.Render("transactions", nil)
	})

	port := ":8080"
	logger.Init("Magasin")
	caissier.InitialiserPOS("init", "Caisse 1", "NO MAGASIN")
	log.Fatal(app.Listen(port))
}

func authMiddleWare(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	err := CheckLogedIn(authHeader)
	if err != nil {
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
