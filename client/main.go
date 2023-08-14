package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Quote struct {
	Dollar string `json:"bid"`
}

type QuoteFile struct {
	Dollar string `json:"Dollar"`
}

func NewQuoteFileFromQuote(quote Quote) *QuoteFile {
	return &QuoteFile{Dollar: quote.Dollar}
}

func main() {
	quote := GetQuote()
	SaveQuote(quote)
}

func SaveQuote(quote *Quote) error {
	file, err := os.Create("quotes.txt")
	if err != nil {
		return err
	}
	_, err = json.Marshal(*quote)
	if err != nil {
		return err
	}
	err = json.NewEncoder(file).Encode(NewQuoteFileFromQuote(*quote))
	if err != nil {
		return err
	}
	return nil
}

func GetQuote() *Quote {
	url := "http://localhost:8080/cotacao"
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var quote Quote
	json.Unmarshal(body, &quote)
	select {
	case <-ctx.Done():
		log.Println("Timeout reached")
	}
	return &quote
}
