package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
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

	//yamlFile, err := ioutil.ReadFile("config.yaml")
	//if err != nil {
	//	log.Printf("yamlFile.Get err   #%v ", err)
	//}
	//err = yaml.Unmarshal(yamlFile, c)
	//if err != nil {
	//	log.Fatalf("Unmarshal: %v", err)
	//}
	//log.Println(c.Database.Port)
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

	log.Println(c.Secrets.IEXToken)
	log.Println(c.Database.Host)
	return c, nil
}

func iexRequest(uri string, token string) []byte {
	BASEURL := "https://cloud.iexapis.com/v1/"
	//TOKEN := "sk_23be4443d656473298137c788414d1cd"
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

	return body
}

func checkResponse(body []byte) bool {
	if string(body) == "[]" {
		return false
	}
	return true
}

func main() {
	c, err := getConf()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Config:\n %s", c.Database.Port)

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

	fmt.Println("Successfully connected!")

	body := iexRequest("stock/v/company", c.Secrets.IEXToken)
	company := Company{}
	json.Unmarshal(body, &company)
	fmt.Printf("%v", string(body))
	fmt.Println("\n\n-----\n\n")
	fmt.Println(company.Sector)

	body = iexRequest("stock/v/dividends/next", c.Secrets.IEXToken)
	dividend := Dividend{}
	json.Unmarshal(body, &dividend)
	if !checkResponse(body) {
		fmt.Println("No dividend data received")
	} else {
		fmt.Printf("%v", string(body))
		amount, err := strconv.ParseFloat(dividend.Amount, 32)
		if err != nil {
			panic(err)
		}
		fmt.Println("\n\n------\n\nDiv amount: $", fmt.Sprintf("%.2f", amount))
	}
}
