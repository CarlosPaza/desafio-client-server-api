package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ExchangeRate struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type ReturnError struct {
	Message string `json:"error_message"`
}

type ExchangeRateBid struct {
	Value string `json:"value"`
}

func main() {
	http.HandleFunc("/cotacao", handler)
	http.ListenAndServe(":8080", nil)

}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
	}

	exchangeRate, err := getExchangeRate("USD-BRL")
	if err != nil {
		log.Print(err)
		writeResponse(w, http.StatusInternalServerError, ReturnError{err.Error()})
		return
	}

	db, err := conexaoDB()
	if err != nil {
		log.Print(err)
		writeResponse(w, http.StatusInternalServerError, ReturnError{err.Error()})
		return
	}

	err = insert(db, exchangeRate)
	if err != nil {
		log.Print(err)
		writeResponse(w, http.StatusInternalServerError, ReturnError{err.Error()})
		return
	}

	writeResponse(w, http.StatusOK, ExchangeRateBid{exchangeRate.Bid})
}

func writeResponse(w http.ResponseWriter, statusCode int, content any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(content)
}

func getExchangeRate(currency string) (*ExchangeRate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/"+currency, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]ExchangeRate
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	exchangeRate := data["USDBRL"]
	return &exchangeRate, nil
}

func conexaoDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "../config/cotacao.db")
	if err != nil {
		return nil, err
	}
	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS exchange_rate (id INTEGER PRIMARY KEY, code TEXT,codein TEXT, name TEXT, high TEXT, low TEXT, varbid TEXT, pctchange TEXT, bid TEXT, ask TEXT, timestamp TEXT, createdate TEXT)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func insert(db *sql.DB, exchangeRate *ExchangeRate) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	stmt, err := db.Prepare("insert into exchange_rate(code, codein,name,high,low,varBid,pctChange,bid,ask,timestamp,createDate) values(?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, exchangeRate.Code, exchangeRate.Codein, exchangeRate.Name, exchangeRate.High, exchangeRate.Low, exchangeRate.VarBid, exchangeRate.PctChange, exchangeRate.Bid, exchangeRate.Ask, exchangeRate.Timestamp, exchangeRate.CreateDate)
	if err != nil {
		return err
	}

	return nil
}
