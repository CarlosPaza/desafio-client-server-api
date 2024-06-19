package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type ReturnExchangeRate struct {
	Value string `json:"value"`
}

func main() {
	value, err := getExchangeRateValue()
	if err != nil {
		panic(err)
	}

	err = writeOnFile(value)
	if err != nil {
		panic(err)
	}

	fmt.Println("Dolar: ", value)
}

func getExchangeRateValue() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return "", fmt.Errorf("erro ao montar requisição: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer requisição: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", fmt.Errorf("erro na API de cotação")
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta: %v", err)
	}

	var data ReturnExchangeRate

	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer parse da resposta: %v", err)
	}

	return data.Value, nil
}

func writeOnFile(value string) error {
	var file *os.File
	defer file.Close()
	_, err := os.Stat("cotacao.txt")
	if os.IsNotExist(err) {
		file, err = os.Create("cotacao.txt")
		if err != nil {
			return fmt.Errorf("erro ao criar arquivo: %v", err)
		}

	} else {
		file, err = os.OpenFile("cotacao.txt", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("erro ao abrir arquivo: %v", err)
		}
	}

	_, err = file.WriteString(fmt.Sprintf("Dolar: %s\n", value))
	if err != nil {
		return fmt.Errorf("erro ao escrever no arquivo arquivo: %v", err)
	}
	return nil
}
