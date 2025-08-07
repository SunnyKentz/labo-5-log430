package utils

import (
	"bytes"
	"caisse-app-scaled/caisse_app_scaled/logger"
	"caisse-app-scaled/caisse_app_scaled/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

const CACHE_TIME = 30 * time.Second

func FAILURE_ERR(err error) string { return "Failed Request : " + err.Error() }
func SYNTAX_ERR(source string, value any) string {
	return fmt.Sprintf("%s : %v could not be parsed, check syntax", source, value)
}
func NOTFOUND_ERR(source string, value any) string {
	return fmt.Sprintf("%s : %v was not found", source, value)
}

func GetApiError(c *fiber.Ctx, message string, status int) error {
	errorMsg := ""
	switch status {
	case 400:
		errorMsg = "Bad Request"
	case 401:
		errorMsg = "Unauthorized"
	case 403:
		errorMsg = "Forbidden"
	case 404:
		errorMsg = "Not Found "
	case 500:
		errorMsg = "Internal Server Error"
	}
	logger.Error(message)
	return c.Status(status).JSON(models.ApiError{
		Timestamp: time.Now(),
		Status:    status,
		Error:     errorMsg,
		Message:   message,
		Success:   false,
		Path:      c.Route().Path,
	})
}

func GetApiSuccess(cache map[string]any, c *fiber.Ctx, status int) error {
	message := ""
	switch status {
	case 200:
		message = "Ok"
	case 201:
		message = "Created"
	}
	path := c.Method() + " " + c.Path()
	now := time.Now()
	cache["t - "+path] = now
	cache[path] = models.ApiSuccess{
		Timestamp: time.Now(),
		Success:   true,
		Status:    status,
		Message:   message,
		Path:      c.Route().Path,
	}
	return c.Status(status).JSON(cache[path].(models.ApiSuccess))
}

func API_MAGASIN() string {
	if os.Getenv("ENVTEST") == "TRUE" {
		return "http://" + os.Getenv("GATEWAY") + ":8080/magasin"
	}
	return "http://" + os.Getenv("GATEWAY") + "/magasin"
}
func API_MERE() string {
	if os.Getenv("ENVTEST") == "TRUE" {
		return "http://" + os.Getenv("GATEWAY") + ":8090/mere"
	}
	return "http://" + os.Getenv("GATEWAY") + "/mere"
}
func API_LOGISTIC() string {
	if os.Getenv("ENVTEST") == "TRUE" {
		return "http://" + os.Getenv("GATEWAY") + ":8091/logistique"
	}
	return "http://" + os.Getenv("GATEWAY") + "/logistique"
}

func API_AUTH() string {
	if os.Getenv("ENVTEST") == "TRUE" {
		return "http://" + os.Getenv("GATEWAY") + ":8092/auth"
	}
	return "http://" + os.Getenv("GATEWAY") + "/auth"
}

func Login(employe string, pw string, role string) (string, error) {
	resp, err := http.PostForm(API_AUTH()+"/api/v1/login", map[string][]string{
		"username": {employe},
		"role":     {role},
	})
	if err != nil {
		logger.Error("Erreur lors de la connexion: " + err.Error())
		return "", err
	}
	defer resp.Body.Close()
	var body struct {
		Jwt string `json:"jwt"`
	}
	// print resp.Body
	bodyBytes, _ := io.ReadAll(resp.Body)
	println(string(bodyBytes))
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	if pw != "password" {
		return "", errors.New("wrong password")
	}
	return body.Jwt, nil
}

func CheckLogedIn(jwt string) error {

	body := map[string]any{
		"jwt": jwt,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(body)
	if err != nil {
		logger.Error("Erreur lors de la sérialisation: " + err.Error())
		return err
	}

	// Send POST request
	resp, err := http.Post(API_AUTH()+"/api/v1/validate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("jwt non valide")
	}
	return nil
}

func Errnotnil(err error) {
	if err != nil {
		logger.Error(err.Error())
	}
}
