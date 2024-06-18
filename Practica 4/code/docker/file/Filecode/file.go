package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type Document struct {
	Content string `json:"content"`
}

var userDatabase map[string]User
var mutex sync.Mutex
var MainDirGlobal="/app/data/"
var CertDirGlobal="/app/"
var globalVersion =1.0

func main(){


	http.HandleFunc("/", manageContenbyId)
	http.HandleFunc("/file",createDir)
    serverAddr := "10.0.2.4:5000"
	cert:=fmt.Sprintf("%sCert.crt",CertDirGlobal)
	key:=fmt.Sprintf("%sKey.key",CertDirGlobal)
	print("Running...")
	log.Fatal(http.ListenAndServeTLS(serverAddr, cert, key, nil))
}

func createDir (w http.ResponseWriter, r *http.Request){
	dir := MainDirGlobal

    var user User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    dir = dir + "/" + user.Username
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        os.Mkdir(dir, os.ModePerm)
    }

}
func manageContenbyId(w http.ResponseWriter, r *http.Request){
		path:=r.URL.Path
		parts := strings.Split(path, "/")
			username:=parts[1]
			documentName:=parts[2]
		if(documentName=="_all_docs" && r.Method==http.MethodGet){
			getAllDocsHandler(w, r,username)
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
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Method not allowed\n")
		return
	}
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
	print(jsonData)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
func postContentByIdHandler(w http.ResponseWriter, r *http.Request,username string, docname string ){
	dir:=MainDirGlobal
	path := fmt.Sprintf("%s%s/%s%s",dir,username,docname,".json")
	dirPath := fmt.Sprintf("%s/%s", dir, username)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
        errDir := os.MkdirAll(dirPath, 0755)
        if errDir != nil {
            log.Fatal(err)
        }
    }
	
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


//********Funciones Auxiliares*************

func n_bytesFile(path string)string{
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	n_bytes:=fileInfo.Size()
	n_bytes_str:=fmt.Sprintf("%d",n_bytes)
	return n_bytes_str
}
