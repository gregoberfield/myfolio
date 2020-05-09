package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type Configuration struct {
	Token string `yaml: "token"`
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
	ExDate      string `json:"exDate"`
	PaymentDate string `json:"paymentDate"`
	RecordDate	string `json:"recordDate"`
	DeclaredDate	string `json:"declaredDate"`
	Amount	string `json:"amount"`
	EventType	string `json:"flag"`
	Currency 	string `json:"currency"`
	Description string `json:"description"`
	Frequency	string `json:"frequency"`
}

func (c *Configuration) getConf() *Configuration {

	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
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
	var c Configuration
	c.getConf()
	fmt.Println(c)
	body := iexRequest("stock/v/company", c.Token)

	company := Company{}
	json.Unmarshal(body, &company)
	fmt.Printf("%v", string(body))
	fmt.Println("\n\n-----\n\n")
	fmt.Println(company.Sector)

	body = iexRequest("stock/v/dividends/next", c.Token)
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
