package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func NewUserFile(userId string) error {
	url := "http://localhost:9090/CreateEmptyFolder" // URL da API

	// Dados a serem enviados no corpo da requisição
	payload := map[string]string{"userId": userId}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Erro ao converter payload para JSON: %v", err)
	}

	// Fazer a requisição POST
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Fazer a requisição POST
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Erro ao fazer requisição POST: %v", err)
	}
	if err != nil {
		return err
	}
	defer resp.Body.Close() // Garantir que o corpo da resposta será fechado

	// Ler o corpo da resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err

	}

	// Imprimir o corpo da resposta
	fmt.Println(string(body))
	return nil
}
