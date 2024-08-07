package emailHandler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SendEmail manipula a requisição para enviar um e-mail

type EmailStruct struct {
	Subject string   `json:"subject"`
	To      []string `json:"to"`
	Name    string   `json:"name"`
	Data    string   `json:"data"`
}

func SendEmail(c *gin.Context, emailData EmailStruct) (string, error) {
	url := "http://localhost:7070/sendEmail/text"
	method := "POST"

	jsonPayload, err := json.Marshal(emailData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return "", err
	}

	// fmt.Sprintf("http://localhost:4200/dashboard?accessToken=%s&expiry=%s", token, user.TokenExpiry)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err

	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return "", err

	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err

	}

	return string(body), nil
}
