package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Configuration struct {
	Secrets struct {
		IEXToken string `yaml:"iextoken"`
	}
	Database struct {
		Host     string `yaml:"host"`
		Port     int16  `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"pass"`
		Name     string `yaml:"name"`
	}
}

type Prices struct {
	Price float64 `json:"price"`
	Time  int     `json:"time"`
}

type OHLC struct {
	O Prices  `json:"open"`
	C Prices  `json:"close"`
	H float64 `json:"high"`
	L float64 `json:"low"`
}

type Company struct {
	Symbol      string `json:"symbol"`
	CompanyName string `json:"companyName"`
	Employees   int    `json:"employees"`
	Exchange    string `json:"exchange"`
	Industry    string `json:"industry"`
	Website     string `json:"website"`
	CEO         string `json:"CEO"`
	Sector      string `json:"sector"`
	SicCode     string `json:"primarySicCode"`
}

type Dividend struct {
	ExDate       string `json:"exDate"`
	PaymentDate  string `json:"paymentDate"`
	RecordDate   string `json:"recordDate"`
	DeclaredDate string `json:"declaredDate"`
	Amount       string `json:"amount"`
	EventType    string `json:"flag"`
	Currency     string `json:"currency"`
	Description  string `json:"description"`
	Frequency    string `json:"frequency"`
}

func getConf() (*Configuration, error) {
	c := &Configuration{}

	// Open config file
	file, err := os.Open("config.yaml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&c); err != nil {
		return nil, err
	}

	return c, nil
}

func iexRequest(uri string, ticker string, token string) ([]byte, error) {
	BASEURL := "https://cloud.iexapis.com/v1/"
	req, err := http.NewRequest(http.MethodGet, BASEURL+uri+"?token="+token, nil)
	if err != nil {
		panic(err)
	}
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("IEXCloud: Ticker Lookup Failed (%s)", ticker))
	}

	return body, nil
}

func checkResponse(body []byte) bool {
	if string(body) == "[]" {
		return false
	}
	return true
}

// ParseFlags will create and parse the CLI flags
// and return the path to be used elsewhere
func ParseFlags() (string, error) {
	var ticker string

	// Set up a CLI flag called "-ticker" to allow users
	// to supply the configuration file
	flag.StringVar(&ticker, "ticker", "aapl", "ticker to lookup")

	// Actually parse the flags
	flag.Parse()

	// Return the ticker
	return ticker, nil
}

func main() {
	c, err := getConf()
	if err != nil {
		panic(err)
	}

	ticker, err := ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.Name)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	log.Println("Successfully connected to database!")

	body, err := iexRequest(fmt.Sprintf("stock/%s/company", ticker), ticker, c.Secrets.IEXToken)
	if err != nil {
		log.Fatal(err)
	}
	company := Company{}
	json.Unmarshal(body, &company)
	log.Println(company.Sector)

	body, err = iexRequest(fmt.Sprintf("stock/%s/dividends/next", ticker), ticker, c.Secrets.IEXToken)
	if err != nil {
		log.Fatal(err)
	}
	dividend := Dividend{}
	json.Unmarshal(body, &dividend)
	if !checkResponse(body) {
		log.Println("No dividend data received")
	} else {
		//amount, err := strconv.ParseFloat(dividend.Amount, 32)

		amount, err := decimal.NewFromString(dividend.Amount)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(amount.StringFixedBank(2))
	}

	body, err = iexRequest(fmt.Sprintf("stock/%s/ohlc", ticker), ticker, c.Secrets.IEXToken)
	if err != nil {
		log.Fatal(err)
	}
	ohlc := OHLC{}
	json.Unmarshal(body, &ohlc)
	closeprice := decimal.NewFromFloat(ohlc.C.Price)
	log.Println(closeprice.StringFixedBank(2))
}
