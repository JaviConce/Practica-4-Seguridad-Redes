package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// VARIABLES GLOBALES
var MainDirGlobal = "/app/"

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var userDatabase map[string]User
var mutex sync.Mutex

func init() {
	userDatabase = make(map[string]User)
	path := fmt.Sprintf("%suser.json", MainDirGlobal)
	loadUsersFromJSON(path)
}

func main() {
	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/token", checkdecodeToken)
	serverAddr := "10.0.2.3:5000"
	cert := fmt.Sprintf("%sCert.crt", MainDirGlobal)
	key := fmt.Sprintf("%sKey.key", MainDirGlobal)
	print("Running...")
	log.Fatal(http.ListenAndServeTLS(serverAddr, cert, key, nil))
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Method not allowed\n")
		return
	}
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid JSON format\n")
		return
	}
	newUser.Password = transpass(newUser.Password)
	if _, exists := userDatabase[newUser.Username]; exists {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, "Username already exists\n")
		return
	}
	mutex.Lock()
	jsonData, err := json.Marshal(newUser)
	if err != nil {
		fmt.Println(err)
		return
	}

	userDatabase[newUser.Username] = newUser
	client := aux()
	resp, err := client.Post("https://10.0.1.4:5000/file", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Response status:", resp.Status)
	checkDirectories(newUser.Username)
	token, err := generateAccessToken(newUser.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error generating token\n")
		return
	}
	// Devolver el token al cliente
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"access_token": token})
	path := fmt.Sprintf("%suser.json", MainDirGlobal)
	writeJsonFile(userDatabase, path)
	defer mutex.Unlock()
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Method not allowed\n")
		return
	}
	println("Entra en login")

	var loginUser User
	err := json.NewDecoder(r.Body).Decode(&loginUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid JSON format\n")
		return
	}
	mutex.Lock()
	loginUser.Password = transpass(loginUser.Password)

	user, exists := userDatabase[loginUser.Username]
	if exists && user.Password == loginUser.Password {
		token, err := generateAccessToken(loginUser.Username)
		if err != nil {
			http.Error(w, "Error generating access token\n", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"access_token": token})
	} else {
		http.Error(w, "Incorrect username or password\n", http.StatusUnauthorized)
	}
	mutex.Unlock()
}
func writeJsonFile(data map[string]User, filePath string) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Errorf("Error converting to JSON: %v", err)
	}

	// Crear el directorio si no existe
	dataDir := filepath.Dir(filePath)
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		os.Mkdir(dataDir, os.ModePerm)
	}

	// Escribir el JSON en el archivo especificado
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		fmt.Errorf("Error writing to the file: %v\n", err)
	}
}
func generateAccessToken(username string) (string, error) {
	expirationTime := time.Now().Add(time.Minute * 5)

	tokenString := username + "|" + expirationTime.Format(time.RFC3339)
	tokenBytes := []byte(tokenString)
	encodedToken := base64.StdEncoding.EncodeToString(tokenBytes)

	return encodedToken, nil
}
func checkdecodeToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type requestBody struct {
		Token    string `json:"token"`
		Username string `json:"username"`
	}

	// Leer el cuerpo de la solicitud
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	// Deserializar el cuerpo de la solicitud en la estructura
	var data requestBody
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusInternalServerError)
		return
	}

	decodedToken, err := base64.StdEncoding.DecodeString(data.Token)
	if err != nil {
		http.Error(w, "Error decoding token", http.StatusInternalServerError)
		return
	}
	parts := strings.Split(string(decodedToken), "|")
	if len(parts) != 2 {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	username := parts[0]
	expirationTime, err := time.Parse(time.RFC3339, parts[1])
	if err != nil {
		http.Error(w, "Error parsing expiration time", http.StatusInternalServerError)
		return
	}

	if username != data.Username || !expirationTime.After(time.Now()) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}
func transpass(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func loadUsersFromJSON(filePath string) error {
	// Leer el contenido del archivo JSON
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Error reading JSON file: %v\n", err)
	}

	// Deserializar el contenido JSON en un mapa de usuarios
	err = json.Unmarshal(jsonData, &userDatabase)
	if err != nil {
		return fmt.Errorf("Error when deserializing the JSON file: %v\n", err)
	}

	return nil
}
func checkDirectories(username string) {
	path := fmt.Sprintf("%s%s", MainDirGlobal, username)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Errorf("Error creating the directory: %v\n", err)
		}
	}
}

func aux() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	var client = &http.Client{Transport: tr}
	return client
}
