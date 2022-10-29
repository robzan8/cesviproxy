package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

var authHeader string

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT not set")
	}
	authHeader = os.Getenv("AUTH")
	if authHeader == "" {
		log.Fatal("$AUTH not set")
	}

	http.HandleFunc("/forecast/", allowCrossOrigin(restrictMethod(getForecast, http.MethodGet)))

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

type httpHandler = func(http.ResponseWriter, *http.Request)

func restrictMethod(handler httpHandler, method string) httpHandler {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == method {
			handler(w, req)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Unsupported method %s", req.Method)
	}
}

func allowCrossOrigin(handler httpHandler) httpHandler {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if req.Method == http.MethodOptions {
			return // OK
		}
		handler(w, req)
	}
}

func handleInternalErr(w http.ResponseWriter, err error) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, "Internal server error")
}

func getForecast(w http.ResponseWriter, req *http.Request) {
	if m, _ := path.Match("/forecast/*", req.URL.Path); !m {
		http.NotFound(w, req)
		return
	}
	_, regionId := path.Split(req.URL.Path)

	const farmId = "83fbae5a-32e2-44b2-b618-3d3f2df2b4c4"
	url := "https://dacom.farm/api/v3/farms/" + farmId + "/weather_forecasts/" + regionId + "/24h/"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		handleInternalErr(w, err)
		return
	}
	req.Header.Add("Authorization", authHeader)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		handleInternalErr(w, err)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
