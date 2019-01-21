package main

import (
	"GoCrypt/crypt"
	"GoCrypt/cryptstore"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
)

type jsondata struct {
	File string
	Hash string
}

func getdata() {
	var w http.ResponseWriter
	connStr := "postgres://docker:docker@db/filehashes?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, fileid, ipfshash FROM filemaps")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var fileid string
		var ipfshash string
		err = rows.Scan(&id, &fileid, &ipfshash)
		if err != nil {
			log.Fatal(err)
		}
		data := jsondata{fileid, ipfshash}
		js, err := json.Marshal(data)
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// This needs to have better URL parsing implemented...
	keys := strings.Split(r.URL.Path, "/")[2]
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	data := crypt.Encrypt(bodyBytes, strings.Split(keys, "=")[1])
	w.Write([]byte("Received File, Encrypting Data.\n"))
	fileid := strings.Split(keys, "=")[0] + ":" + crypt.Signature(data)
	if cryptstore.Get(fileid) == true {
		fmt.Println("File has already been uploaded to database..")
	} else {
		fmt.Println("Didnt find file Signature.. Uploading..")
		cryptstore.Store(fileid, data)
	}
}
func apicall(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		switch r.RequestURI {
		case "/api/":
			getdata()
		}
	} else {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
}

func main() {
	http.HandleFunc("/api/", apicall)
	http.HandleFunc("/upload/", uploadHandler)
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}
