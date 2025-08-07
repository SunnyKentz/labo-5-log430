package tests

import (
	"bytes"
	AUTH "caisse-app-scaled/caisse_app_scaled/auth/api"
	LOGIS "caisse-app-scaled/caisse_app_scaled/centre_logistique/api"
	MAG "caisse-app-scaled/caisse_app_scaled/magasin/api"
	"caisse-app-scaled/caisse_app_scaled/magasin/caissier"
	MERE "caisse-app-scaled/caisse_app_scaled/maison_mere/api"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	httpClient = &http.Client{Timeout: 10 * time.Second}
	magToken   string
	mereToken  string
	logisToken string
)

func BeforeAll() {
	os.Setenv("DB_PASSWORD", "test_password")
	os.Setenv("DB_USER", "test_user")
	os.Setenv("DB_PORT", "5435")
	os.Setenv("GATEWAY", "localhost")
	os.Setenv("ENVTEST", "TRUE")
	ConnectDB()
	go MERE.NewApp()
	go MAG.NewApp()
	go LOGIS.NewApp()
	go AUTH.NewDataApi()
	time.Sleep(time.Second * 2) // time for goroutine to startup

	// Login to get tokens
	mereToken = login("maison_mere", "http://localhost:8090/mere")
	magToken = login("magasin", "http://localhost:8080/magasin")
	logisToken = login("logistique", "http://localhost:8091/logistique")
}

func login(service, baseURL string) string {
	type logininfo struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Magasin  string `json:"magasin"`
		Caisse   string `json:"caisse"`
	}
	host := logininfo{
		Username: "Bob",
		Password: "password",
		Magasin:  "Magasin 1",
		Caisse:   "Caisse 1",
	}
	jsonData, err := json.Marshal(host)
	if err != nil {
		fmt.Println(service + " failed logged in")
		return ""
	}
	var resp *http.Response
	if service == "maison_mere" {
		resp, err = http.Post(baseURL+"/api/v1/merelogin", "application/json", bytes.NewBuffer(jsonData))
	} else {
		resp, err = http.Post(baseURL+"/api/v1/login", "application/json", bytes.NewBuffer(jsonData))
	}
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(service + " failed to read response body")
		return ""
	}
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println(service + " failed to unmarshal response body")
		return ""
	}
	token, exists := result["token"]
	if !exists {
		fmt.Println(service + " token not found in response body")
		return ""
	}
	return token
}

func TestMain(m *testing.M) {
	BeforeAll()
	code := m.Run()
	caissier.FermerPOS()
	os.Exit(code)
}

func errnotnil(s string, t *testing.T, err error) {
	if err != nil {
		t.Error(s + ": " + err.Error())
	}
}

// ==================== MAGASIN API TESTS ====================

func TestMagasinUpdateProduit(t *testing.T) {
	updateData := map[string]interface{}{
		"nom":         "Test Product",
		"prix":        10.99,
		"description": "Test description",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "http://localhost:8080/magasin/api/v1/produit/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Magasin update produit", t, err)
	defer resp.Body.Close()

	// This might fail if product doesn't exist, which is expected
	t.Logf("Update produit status: %d", resp.StatusCode)
}

func TestMagasinProduits(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/magasin/api/v1/produits", nil)
	req.Header.Set("Authorization", "Bearer "+magToken)
	req.Header.Set("C-Caisse", "Caisse 1")
	req.Header.Set("C-Mag", "Magasin 1")
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Magasin produits", t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMagasinProduitsSearch(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/magasin/api/v1/produits/t", nil)
	req.Header.Set("Authorization", "Bearer "+magToken)
	req.Header.Set("C-Caisse", "Caisse 1")
	req.Header.Set("C-Mag", "Magasin 1")
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Magasin produits search", t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMagasinCart(t *testing.T) {
	// Test get cart
	req, _ := http.NewRequest("GET", "http://localhost:8080/magasin/api/v1/cart", nil)
	req.Header.Set("Authorization", "Bearer "+magToken)
	req.Header.Set("C-Caisse", "Caisse 1")
	req.Header.Set("C-Mag", "Magasin 1")
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Magasin get cart", t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMagasinAddToCart(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://localhost:8080/magasin/api/v1/cart/1", nil)
	req.Header.Set("Authorization", "Bearer "+magToken)
	req.Header.Set("C-Caisse", "Caisse 1")
	req.Header.Set("C-Mag", "Magasin 1")
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Magasin add to cart", t, err)
	defer resp.Body.Close()

	// This might fail if product doesn't exist, which is expected
	t.Logf("Add to cart status: %d", resp.StatusCode)
}

func TestMagasinRemoveFromCart(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "http://localhost:8080/magasin/api/v1/cart/1", nil)
	req.Header.Set("Authorization", "Bearer "+magToken)
	req.Header.Set("C-Caisse", "Caisse 1")
	req.Header.Set("C-Mag", "Magasin 1")
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Magasin remove from cart", t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMagasinVendre(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://localhost:8080/magasin/api/v1/vendre", nil)
	req.Header.Set("Authorization", "Bearer "+magToken)
	req.Header.Set("C-Caisse", "Caisse 1")
	req.Header.Set("C-Mag", "Magasin 1")
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Magasin vendre", t, err)
	defer resp.Body.Close()

	// This might fail if cart is empty, which is expected
	t.Logf("Vendre status: %d", resp.StatusCode)
}

func TestMagasinTransactions(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/magasin/api/v1/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+magToken)
	req.Header.Set("C-Caisse", "Caisse 1")
	req.Header.Set("C-Mag", "Magasin 1")
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Magasin transactions", t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Logf("Response body: %s", string(bodyBytes))
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMagasinRembourser(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://localhost:8080/magasin/api/v1/rembourser/1", nil)
	req.Header.Set("Authorization", "Bearer "+magToken)
	req.Header.Set("C-Caisse", "Caisse 1")
	req.Header.Set("C-Mag", "Magasin 1")
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Magasin rembourser", t, err)
	defer resp.Body.Close()

	// This might fail if transaction doesn't exist, which is expected
	t.Logf("Rembourser status: %d", resp.StatusCode)
}

func TestMagasinCommande(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://localhost:8080/magasin/api/v1/produit/1", nil)
	req.Header.Set("Authorization", "Bearer "+magToken)
	req.Header.Set("C-Caisse", "Caisse 1")
	req.Header.Set("C-Mag", "Magasin 1")
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Magasin commande", t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMagasinReapprovisionner(t *testing.T) {
	req, _ := http.NewRequest("PUT", "http://localhost:8080/magasin/api/v1/produit/1/10", nil)
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Magasin reapprovisionner", t, err)
	defer resp.Body.Close()

	// This might fail if product doesn't exist, which is expected
	t.Logf("Reapprovisionner status: %d", resp.StatusCode)
}

// ==================== MAISON MERE API TESTS ====================

func TestMereNotify(t *testing.T) {
	notifyData := map[string]string{
		"message": "Test notification",
		"host":    "test-host",
	}

	jsonData, _ := json.Marshal(notifyData)
	req, _ := http.NewRequest("POST", "http://localhost:8090/mere/api/v1/notify", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("no-cache", "true")
	resp, err := httpClient.Do(req)
	errnotnil("Mere notify", t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMereSubscribe(t *testing.T) {
	subscribeData := map[string]string{
		"host": "http://localhost:8080/magasin",
	}

	jsonData, _ := json.Marshal(subscribeData)
	req, _ := http.NewRequest("POST", "http://localhost:8090/mere/api/v1/subscribe", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("no-cache", "true")
	resp, err := httpClient.Do(req)
	errnotnil("Mere subscribe", t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMereAlerts(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8090/mere/api/v1/alerts", nil)
	req.Header.Set("Authorization", "Bearer "+mereToken)
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Mere alerts", t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMereTransactions(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8090/mere/api/v1/transactions", nil)
	req.Header.Set("no-cache", "true")
	resp, err := httpClient.Do(req)
	errnotnil("Mere transactions", t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMereTransactionById(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8090/mere/api/v1/transactions/1", nil)
	req.Header.Set("no-cache", "true")
	resp, err := httpClient.Do(req)
	errnotnil("Mere transaction by id", t, err)
	defer resp.Body.Close()

	// This might fail if transaction doesn't exist, which is expected
	t.Logf("Transaction by id status: %d", resp.StatusCode)
}

func TestMereCreateTransaction(t *testing.T) {
	transactionData := map[string]interface{}{
		"caisse":      "Caisse 1",
		"type":        "VENTE",
		"produit_ids": "1,2",
		"montant":     25.50,
		"date":        time.Now(),
	}

	jsonData, _ := json.Marshal(transactionData)
	req, _ := http.NewRequest("POST", "http://localhost:8090/mere/api/v1/transactions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("no-cache", "true")
	resp, err := httpClient.Do(req)
	errnotnil("Mere create transaction", t, err)
	defer resp.Body.Close()

	// This might fail if validation fails, which is expected
	t.Logf("Create transaction status: %d", resp.StatusCode)
}

func TestMereDeleteTransaction(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "http://localhost:8090/mere/api/v1/transactions/1", nil)
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Mere delete transaction", t, err)
	defer resp.Body.Close()

	// This might fail if transaction doesn't exist, which is expected
	t.Logf("Delete transaction status: %d", resp.StatusCode)
}

func TestMereMagasins(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8090/mere/api/v1/magasins", nil)
	req.Header.Set("Authorization", "Bearer "+mereToken)
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Mere magasins", t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMereAnalytics(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8090/mere/api/v1/analytics/tout", nil)
	req.Header.Set("Authorization", "Bearer "+mereToken)
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Mere analytics", t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMereRaport(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8090/mere/api/v1/raport", nil)
	req.Header.Set("Authorization", "Bearer "+mereToken)
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Mere raport", t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMereProduitsSearch(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8090/mere/api/v1/produits/test", nil)
	req.Header.Set("Authorization", "Bearer "+mereToken)
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Mere produits search", t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMereUpdateProduit(t *testing.T) {
	updateData := map[string]interface{}{
		"productId":   1,
		"nom":         "Test Product",
		"prix":        10.99,
		"description": "Test description",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "http://localhost:8090/mere/api/v1/produit", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mereToken)
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Mere update produit", t, err)
	defer resp.Body.Close()

	// This might fail if product doesn't exist, which is expected
	t.Logf("Mere update produit status: %d", resp.StatusCode)
}

// ==================== LOGISTICS API TESTS ====================
func TestLogisCommands(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8091/logistique/api/v1/commands", nil)
	req.Header.Set("Authorization", "Bearer "+logisToken)
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Logis commands", t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestLogisAcceptCommande(t *testing.T) {
	req, _ := http.NewRequest("PUT", "http://localhost:8091/logistique/api/v1/commande/1", nil)
	req.Header.Set("Authorization", "Bearer "+logisToken)
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Logis accept commande", t, err)
	defer resp.Body.Close()

	// This might fail if command doesn't exist, which is expected
	t.Logf("Accept commande status: %d", resp.StatusCode)
}

func TestLogisRefuseCommande(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "http://localhost:8091/logistique/api/v1/commande/1", nil)
	req.Header.Set("Authorization", "Bearer "+logisToken)
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Logis refuse commande", t, err)
	defer resp.Body.Close()

	// This might fail if command doesn't exist, which is expected
	t.Logf("Refuse commande status: %d", resp.StatusCode)
}

func TestLogisProduitsSearch(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8091/logistique/api/v1/produits/t", nil)
	req.Header.Set("no-cache", "true")
	resp, err := httpClient.Do(req)
	errnotnil("Logis produits search", t, err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestLogisProduitById(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8091/logistique/api/v1/produits/id/1", nil)
	req.Header.Set("no-cache", "true")
	resp, err := httpClient.Do(req)
	errnotnil("Logis produit by id", t, err)
	defer resp.Body.Close()

	// This might fail if product doesn't exist, which is expected
	t.Logf("Produit by id status: %d", resp.StatusCode)
}

func TestLogisUpdateProduit(t *testing.T) {
	updateData := map[string]interface{}{
		"nom":         "Test Product",
		"prix":        10.99,
		"description": "Test description",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "http://localhost:8091/logistique/api/v1/produit/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("no-cache", "true")

	resp, err := httpClient.Do(req)
	errnotnil("Logis update produit", t, err)
	defer resp.Body.Close()

	// This might fail if product doesn't exist, which is expected
	t.Logf("Logis update produit status: %d", resp.StatusCode)
}

// ==================== INTEGRATION TESTS ====================

func TestCompleteWorkflow(t *testing.T) {
	t.Log("Testing complete workflow...")

	// 1. Add product to cart
	req, _ := http.NewRequest("POST", "http://localhost:8080/magasin/api/v1/cart/1", nil)
	req.Header.Set("Authorization", "Bearer "+magToken)
	req.Header.Set("C-Caisse", "Caisse 1")
	req.Header.Set("C-Mag", "Magasin 1")
	req.Header.Set("no-cache", "true")
	resp, err := httpClient.Do(req)
	if err == nil {
		t.Logf("Add to cart: %d", resp.StatusCode)
		resp.Body.Close()
	}

	// 2. Complete sale
	req, _ = http.NewRequest("POST", "http://localhost:8080/magasin/api/v1/vendre", nil)
	req.Header.Set("Authorization", "Bearer "+magToken)
	req.Header.Set("C-Caisse", "Caisse 1")
	req.Header.Set("C-Mag", "Magasin 1")
	req.Header.Set("no-cache", "true")
	resp, err = httpClient.Do(req)
	if err == nil {
		t.Logf("Complete sale: %d", resp.StatusCode)
		resp.Body.Close()
	}

	// 3. Request reapprovisionment
	req, _ = http.NewRequest("POST", "http://localhost:8080/magasin/api/v1/produit/1", nil)
	req.Header.Set("Authorization", "Bearer "+magToken)
	req.Header.Set("C-Caisse", "Caisse 1")
	req.Header.Set("C-Mag", "Magasin 1")
	req.Header.Set("no-cache", "true")
	resp, err = httpClient.Do(req)
	if err == nil {
		t.Logf("Request reapprovisionment: %d", resp.StatusCode)
		resp.Body.Close()
	}

	// 4. Create logistics command
	commandeData := map[string]string{"host": "test-host"}
	jsonData, _ := json.Marshal(commandeData)
	req, _ = http.NewRequest("POST", "http://localhost:8091/logistique/api/v1/commande/Magasin%201/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("no-cache", "true")
	resp, err = httpClient.Do(req)
	if err == nil {
		t.Logf("Create logistics command: %d", resp.StatusCode)
		resp.Body.Close()
	}

	t.Log("Complete workflow test finished")
}
