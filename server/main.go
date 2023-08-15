package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"net/http"
	"time"
)

type ExchangeValue struct {
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

type Quote struct {
	ExchangeValue ExchangeValue `json:"USDBRL"`
}

const url = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

func main() {
	migrate()
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", handleQuotation)

	http.ListenAndServe(":8080", mux)
}

func migrate() {

	db, err := getDB()
	if err != nil {
		panic(err)
	}
	sql := `CREATE TABLE IF NOT EXISTS quotes (
    id integer not null primary key autoincrement,
    code text not null,
    codein text not null,
    name text not null,
    high text not null,
    low text not null,
    var_bid text not null,
    pct_change text not null,
    bid text not null,
    ask text not null)`
	_, err = db.Exec(sql)
	if err != nil {
		panic(err.Error())
	}
}

func getDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./quote.db")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func SaveQuote(quote *ExchangeValue) error {

	db, err := getDB()
	if err != nil {
		return err
	}
	timeout := time.Millisecond * 10
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	insertQuery := `INSERT INTO quotes(code,
                   codein,
                   name,
                   high,
                   low,
                   var_bid,
                   pct_change,
                   bit,
                   ask) VALUES (?,?,?,?,?,?,?,?,?)`
	stmt, err := db.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(
		ctx,
		quote.Code,
		quote.Codein,
		quote.Name,
		quote.High,
		quote.Low,
		quote.VarBid,
		quote.PctChange,
		quote.Bid,
		quote.Ask)
	if err != nil {
		return err
	}
	contextTimeoutHandle(ctx, "Database timeout")
	return nil
}

func handleQuotation(writer http.ResponseWriter, request *http.Request) {

	quote, err := GetQuotation()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = SaveQuote(quote)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Header().Set("content-type", "application/json")
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(quote)

}

func GetQuotation() (*ExchangeValue, error) {
	timeout := time.Millisecond * 200
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(client)
	if err != nil {
		return nil, err
	}
	contextTimeoutHandle(ctx, "Request timeout")
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var quote Quote
	err = json.Unmarshal(body, &quote)
	if err != nil {
		return nil, err
	}
	return &quote.ExchangeValue, nil
}

func contextTimeoutHandle(ctx context.Context, message string) {

	select {
	case <-ctx.Done():
		fmt.Println(message)
	}
}
