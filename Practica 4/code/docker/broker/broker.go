package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	//"os"
	"strings"
)



var CertDirGlobal="../../cert/"
var globalVersion =1.0
func main() {

	//MANEJO DE PETICIONES
	http.HandleFunc("/signup", redirectAuth(1))
	http.HandleFunc("/login", redirectAuth(1))
	http.HandleFunc("/", manageContenbyId)
	http.HandleFunc("/version", getVersion)
	//INICIO DEL SERVIDOR
	serverAddr := "myserver.local:5000"
	fmt.Printf("Server listening on https://%s\n", serverAddr)
	cert:=fmt.Sprintf("%sCert.crt",CertDirGlobal)
	key:=fmt.Sprintf("%sKey.key",CertDirGlobal)
	log.Fatal(http.ListenAndServeTLS(serverAddr, cert, key, nil))
}


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
		if RequestToAuthToken(token,username){
			print("Authorize")
			if(documentName=="_all_docs"&& r.Method==http.MethodGet){
				//&file.getAllDocsHandler(w, r, username)
			}else{
				//switch r.Method {
				//case http.MethodGet:
				//	// Manejar el método GET
				//	&file.getContentByIdHandler(w, r, username, documentName)
				//case http.MethodPost:
				//	// Manejar el método POST
				//	&file.postContentByIdHandler(w, r, username, documentName)
				//case http.MethodPut:
				//	// Manejar el método PUT
				//	&file.putContentByIdHandler(w, r, username, documentName)
				//case http.MethodDelete:
				//	// Manejar el método DELETE
				//	&file.deleteContentByIdHandler(w, r, username, documentName)
				//default:
				//	w.WriteHeader(http.StatusMethodNotAllowed)
				//	fmt.Fprint(w, "Method not allowed\n")
				//}
			}
		}else{
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Unauthorized\n")
		}
	}
}

func getVersion(w http.ResponseWriter, r *http.Request){
	fmt.Fprint(w, globalVersion,"\n")
}



func redirectAuth(option int) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        client := aux()
		var resp *http.Response
        var err error
        switch option {
        case 1:
            resp, err = client.Post("https://myserver.local:5001" + r.URL.Path, "application/json", r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
        }
        }
        defer resp.Body.Close()
        body, _ := io.ReadAll(resp.Body)
        w.Write(body)
    }
}
func RequestToAuthToken(token string, username string) (bool) {
    // Crear un cliente HTTP
    client := aux()
	
	requestBody, err := json.Marshal(map[string]string{
        "token":    token,
        "username": username,
    })
    if err != nil {
        fmt.Println("Error creating request body:", err)
        return false
    }
    // Crear una nueva solicitud HTTP
    req, err := http.NewRequest("POST", "https://myserver.local:5001/token", bytes.NewBuffer(requestBody))
    if err != nil {
        return false
    }

    req.Header.Set("Content-Type", "application/json")

    // Enviar la solicitud y obtener la respuesta
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error sending request:", err)
        return false
    }
    defer resp.Body.Close()

    // Leer el cuerpo de la respuesta
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return false
    }

    // Comprobar si la respuesta es "OK"
    if string(body) == "OK" {
        return true
    }

    return false
}
func aux() (*http.Client) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	var client = &http.Client{Transport: tr}
	return client
}