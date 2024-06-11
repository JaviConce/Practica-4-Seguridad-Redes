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
var globalVersion =2.0
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
			if(documentName=="_all_docs"&& r.Method==http.MethodGet){
				redirectFile(1)(w,r)
			}else{
				switch r.Method {
				case http.MethodGet:
					// Manejar el método GET
					redirectFile(1)(w,r)
				case http.MethodPost:
					// Manejar el método POST
					redirectFile(2)(w,r)
				case http.MethodPut:
					// Manejar el método PUT
					redirectFile(3)(w,r)
				case http.MethodDelete:
					// Manejar el método DELETE
					redirectFile(4)(w,r)
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
func redirectFile (option int) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        client := aux()
		var resp *http.Response
        var err error
		switch option {
		case 1:
			resp, err = client.Get("https://myserver.local:5002" + r.URL.Path)
		case 2:
			resp, err = client.Post("https://myserver.local:5002" + r.URL.Path, "application/json", r.Body)
		case 3:
			req, _ := http.NewRequest("DELETE", "https://myserver.local:5002"+r.URL.Path, r.Body)
			req.Header.Set("Content-Type", "application/json")
			resp, err = client.Do(req)
		case 4:
			req, _ := http.NewRequest("PUT", "https://myserver.local:5002"+r.URL.Path, r.Body)
			req.Header.Set("Content-Type", "application/json")
			resp, err = client.Do(req)
		}
		
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			print("Error\n")
			return
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