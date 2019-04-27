package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var templates = template.Must(template.ParseFiles("../frontend/index.html"))
var config Config

//GalleonToEuro conversion
const GalleonToEuro = 22

//SickleToEuro conversion
const SickleToEuro = 1.32

//KnutToEuro conversion
const KnutToEuro = 0.044

//Config holds the configuration
type Config struct {
	FixerKey string `json:"fixer_key"`
}

//FixerResponse contains the structure for the fixer API response
type FixerResponse struct {
	Rates map[string]float64 `json:"rates"`
	Base  string             `json:"base"`
}

//ExchangeResponse holds the response to an exchange request
type ExchangeResponse struct {
	Galleons float64
	Sickles  float64
	Knuts    float64
	Success  bool
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var data ExchangeResponse
		r.ParseForm()
		amount, amountErr := strconv.ParseFloat(r.FormValue("amount"), 64)
		if amountErr != nil {
			log.Print("Unable to convert amount to float")
		}

		convert := r.FormValue("convert")

		var urlBuilder strings.Builder
		urlBuilder.WriteString("http://data.fixer.io/api/latest?&access_key=")
		urlBuilder.WriteString(config.FixerKey)
		fixedResp, err := http.Get(urlBuilder.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer fixedResp.Body.Close()
		var fixedBody []byte
		fixedBody, err = ioutil.ReadAll(fixedResp.Body)

		var f FixerResponse
		err = json.Unmarshal(fixedBody, &f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		//Convert to Euros
		switch convert {
		case "eur":
			amount = amount / f.Rates["EUR"]
		case "gbp":
			amount = amount / f.Rates["GBP"]
		case "cad":
			amount = amount / f.Rates["CAD"]
		case "usd":
			amount = amount / f.Rates["USD"]
		}

		//Convert to wizarding money
		var galleons = math.Floor(amount / GalleonToEuro)
		var sickles float64
		var knuts float64
		if math.Mod(amount, GalleonToEuro) != 0.0 {
			remainder := math.Mod(amount, GalleonToEuro)
			sickles = math.Floor(remainder / SickleToEuro)
			if math.Mod(remainder, SickleToEuro) != 0.0 {
				remainder = math.Mod(remainder, SickleToEuro)
				knuts = math.Floor(remainder / KnutToEuro)
			}
		}

		data.Galleons = galleons
		data.Sickles = sickles
		data.Knuts = knuts
		data.Success = true

		returnResponse(w, data)
	} else if r.Method == "GET" {
		http.NotFound(w, r)
	}
}

func returnResponse(w http.ResponseWriter, data ExchangeResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
}

func loadConfigurationFile(file string) Config {
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func main() {
	config = loadConfigurationFile("config.json")
	http.HandleFunc("/", rootHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
