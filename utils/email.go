package utils

import (
	"net"
	"net/mail"
	"strings"
)

// isValidEmail verifica se o email tem um formato válido e se o domínio possui registros MX.
func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	domain := strings.Split(email, "@")[1]
	mxRecords, err := net.LookupMX(domain)
	return err == nil && len(mxRecords) > 0
}
