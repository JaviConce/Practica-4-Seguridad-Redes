package main

import (
	"encoding/json"
	"path/filepath"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"encoding/base64"
	"crypto/sha256"
	"encoding/hex"
	"strings"

)
//VARIABLES GLOBALES
var MainDirGlobal="data/"
var CertDirGlobal="cert/"
var globalVersion =1.0

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type Document struct {
	Content string `json:"content"`
}

var userDatabase map[string]User
var mutex sync.Mutex

func init() {
	//Inicializamos las estructuras de datos que usaremos
	userDatabase = make(map[string]User)
	path:= fmt.Sprintf("%suser.json",MainDirGlobal)
	loadUsersFromJSON(path)
	for username := range userDatabase {
		checkDirectories((username))
	}

}

func main() {

	//MANEJO DE PETICIONES
	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/", manageContenbyId)
	http.HandleFunc("/version", getVersion)
	//INICIO DEL SERVIDOR
	serverAddr := "myserver.local:5000"
	fmt.Printf("Server listening on https://%s\n", serverAddr)
	cert:=fmt.Sprintf("%scert.pem",CertDirGlobal)
	key:=fmt.Sprintf("%skey.pem",CertDirGlobal)
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
	newUser.Password=transpass(newUser.Password)
	if _, exists := userDatabase[newUser.Username]; exists {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, "Username already exists\n")
		return
	}
	mutex.Lock()
	userDatabase[newUser.Username] = newUser
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
	path:=fmt.Sprintf("%suser.json",MainDirGlobal)
	writeJsonFile(userDatabase,path)
	defer mutex.Unlock()
}
func loginHandler(w http.ResponseWriter, r *http.Request){

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Method not allowed\n")
		return
	}

	var loginUser User
	err := json.NewDecoder(r.Body).Decode(&loginUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid JSON format\n")
		return
	}
	mutex.Lock()
	loginUser.Password=transpass(loginUser.Password)

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
func getContentByIdHandler(w http.ResponseWriter, r *http.Request,username string, docname string ){
	dir:=MainDirGlobal
	path := fmt.Sprintf("%s%s/%s%s",dir,username,docname,".json")
	jsonData, err := os.ReadFile(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error reading JSON file\n")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
func getAllDocsHandler(w http.ResponseWriter, r *http.Request,username string){
	dir:=MainDirGlobal
	path := fmt.Sprintf("%s%s",dir,username,)
	archivos, err := os.ReadDir(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error reading files from directory\n")
		return
	}
	id:=1
	var docsDatabase map[string]Document
	docsDatabase= make(map[string]Document)
	for _, archivo:= range archivos{
		pathfile:= filepath.Join(path, archivo.Name())
		content, err := os.ReadFile(pathfile)
		if err != nil {
			fmt.Printf("Error reading file %s: %s\n", archivo.Name(), err)
			continue
		}
		var doc Document
		err = json.Unmarshal(content, &doc)
		if err != nil {
			fmt.Printf("Error decoding JSON in %s\n: %s\n", archivo.Name(), err)
			continue
		}
		iddoc:=fmt.Sprintf("id%d",id)
		id++
		docsDatabase[iddoc]=doc
	}
	jsonData, err := json.Marshal(docsDatabase)
	if err != nil {
		http.Error(w, "Error converting to JSON\n", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
func postContentByIdHandler(w http.ResponseWriter, r *http.Request,username string, docname string ){
	dir:=MainDirGlobal
	path := fmt.Sprintf("%s%s/%s%s",dir,username,docname,".json")
	
	var newDoc Document
	err := json.NewDecoder(r.Body).Decode(&newDoc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid JSON format")
		return
	}

	mutex.Lock()

	data, err := json.Marshal(newDoc)
	if err != nil {
    	w.WriteHeader(http.StatusInternalServerError)
    	fmt.Fprint(w, "Error encoding document to JSON\n")
    	return
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.WriteFile(path, data, 0644)
		if err != nil {
    		w.WriteHeader(http.StatusInternalServerError)
    		fmt.Fprintf(w, "Error writing to file: %v\n", err)
    		return
		}
		n_bytes_str:=n_bytesFile(path)
		defer mutex.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"size": n_bytes_str})
	}else{
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, "The document already exists\n")
	}
}

func putContentByIdHandler(w http.ResponseWriter, r *http.Request,username string, docname string ){
	dir:=MainDirGlobal
	path := fmt.Sprintf("%s%s/%s%s",dir,username,docname,".json")
	var newDoc Document
	err := json.NewDecoder(r.Body).Decode(&newDoc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid JSON format\n")
		return
	}
	mutex.Lock()
	data, err := json.Marshal(newDoc)
	if err != nil {
    	w.WriteHeader(http.StatusInternalServerError)
    	fmt.Fprint(w, "Error encoding document to JSON\n")
    	return
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
    	w.WriteHeader(http.StatusInternalServerError)
    	fmt.Fprintf(w, "Error writing to the file: %v\n", err)
    	return
	}
	n_bytes_str:=n_bytesFile(path)
	defer mutex.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"size": n_bytes_str})
}
func deleteContentByIdHandler(w http.ResponseWriter, r *http.Request,username string, docname string ){
	dir:=MainDirGlobal
	path := fmt.Sprintf("%s%s/%s%s",dir,username,docname,".json")
	mutex.Lock()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "The document does not exist\n")
	}else{
		os.Remove(path)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "[]\n")
	}
	mutex.Unlock()
}

func getVersion(w http.ResponseWriter, r *http.Request){
	fmt.Fprint(w, globalVersion,"\n")
}


//********************** FUNCIONES AUXILIARES **********************
func manageContenbyId(w http.ResponseWriter, r *http.Request){
	path:=r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts)!=3{
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Error, the structure is not correct\n"})
	}else{
		username := parts[1]
		documentName := parts[2]
		token:=r.Header.Get("Authorization")
		if checkdecodeToken(token,username){
			if(documentName=="_all_docs"&& r.Method==http.MethodGet){
				getAllDocsHandler(w, r, username)
			}else{
				switch r.Method {
				case http.MethodGet:
					// Manejar el método GET
					getContentByIdHandler(w, r, username, documentName)
				case http.MethodPost:
					// Manejar el método POST
					postContentByIdHandler(w, r, username, documentName)
				case http.MethodPut:
					// Manejar el método PUT
					putContentByIdHandler(w, r, username, documentName)
				case http.MethodDelete:
					// Manejar el método DELETE
					deleteContentByIdHandler(w, r, username, documentName)
				default:
					w.WriteHeader(http.StatusMethodNotAllowed)
					fmt.Fprint(w, "Method not allowed\n")
				}
			}
		}else{
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Unauthorized\n")
		}
	}
}
func checkDirectories(username string){
	path := fmt.Sprintf("%s%s", MainDirGlobal, username)
	_, err:=os.Stat(path)
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Errorf("Error creating the directory: %v\n", err)
		}
	}
}
func generateAccessToken(username string) (string, error) {
	expirationTime := time.Now().Add(time.Minute *5)

	tokenString := username + "|" + expirationTime.Format(time.RFC3339)
	tokenBytes := []byte(tokenString)
	encodedToken := base64.StdEncoding.EncodeToString(tokenBytes)

	return encodedToken, nil
}
func checkdecodeToken(token string, user string) bool{
	decodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		fmt.Errorf("Error decoding token: %v\n", err)
	}
	parts := strings.Split(string(decodedToken), "|")
	if len(parts)!=2{
		return false
	}
	username := parts[0]
	expirationTime, err := time.Parse(time.RFC3339, parts[1])
	if(username==user){
		
		return expirationTime.After(time.Now())
	
	}
	return false
}
func writeJsonFile(data map[string]User, filePath string){
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

func transpass(password string)string{
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func n_bytesFile(path string)string{
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	n_bytes:=fileInfo.Size()
	n_bytes_str:=fmt.Sprintf("%d",n_bytes)
	return n_bytes_str
}
