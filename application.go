package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"./simpleCache"
	"github.com/rs/cors"
)

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

		amount, convert, validateErr := validateForm(r)

		if validateErr != "" {
			http.Error(w, validateErr, http.StatusInternalServerError)
			return
		}

		var f FixerResponse
		savedCache, err := simpleCache.GetCache()

		if err != nil {
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

			err = json.Unmarshal(fixedBody, &f)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			savedCache.Rates = f.Rates
			savedCache.Base = f.Base
			savedCache.Expires = time.Now().Add(time.Hour * 24).Unix()

			simpleCache.SaveCache(savedCache)
		} else {
			f.Rates = savedCache.Rates
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

func validateForm(r *http.Request) (float64, string, string) {
	r.ParseForm()
	var err string
	log.Print(r.Body)
	for key, values := range r.PostForm {
		log.Print(key)
		log.Print(values)
	}

	amount, amountErr := strconv.ParseFloat(r.FormValue("amount"), 64)
	if amountErr != nil {
		log.Print("Unable to convert amount to float")
		err += "Amount submitted missing/invalid."
	}
	convert := r.FormValue("convert")
	if convert == "" {
		log.Print("Conversion currency missing")
		err += "Conversion currency not submitted."
	}

	availableCurrencies := []string{"eur", "gbp", "cad", "usd"}
	var found bool

	for i := 0; i < len(availableCurrencies); i++ {

		if strings.ContainsAny(convert, availableCurrencies[i]) {
			found = true
		}
	}

	if !found {
		err += "Submitted conversion currency not valid. Available currencies: " + strings.Join(availableCurrencies, ",")
	}

	return amount, convert, err

}

func returnResponse(w http.ResponseWriter, data ExchangeResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	handler := cors.Default().Handler(mux)
	log.Fatal(http.ListenAndServe(":5000", handler))
}
