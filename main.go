//go:build ignore
// +build ignore

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

var status map[string](chan string) = make(map[string](chan string))

var addr = flag.String("addr", ":8000", "http service address")

type Page struct {
	FileName string
}

type BodyData struct {
	Value string `json:"value"`
}

func Home(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("tmpl.html")
	if err != nil {
		panic(err)
	}
	data := []byte{}
	err = tmpl.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

func Get(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	log.Printf("key %s", key)
	status[key] = make(chan string) // No need buffer

	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// https://www.itzhimei.com/archives/6548.html
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		w.WriteHeader(http.StatusGatewayTimeout)
	case value := <-status[key]:
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/plain;charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(value))
	}
}

func Set(w http.ResponseWriter, r *http.Request) {
	var bodyData BodyData
	err := json.NewDecoder(r.Body).Decode(&bodyData)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Decode error! please check your JSON formating.")
		return
	}
	key := r.URL.Query().Get("key")
	status[key] <- bodyData.Value
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stdout)

	flag.Parse()
	http.HandleFunc("/status/get", Get)
	http.HandleFunc("/status/set", Set)
	http.HandleFunc("/", Home)
	log.Printf("starting server at 8000")
	srv := &http.Server{
		// ReadTimeout:  5 * time.Second,
		// WriteTimeout: 10 * time.Second,
		Addr: *addr,
	}
	log.Fatal(srv.ListenAndServe())
}
