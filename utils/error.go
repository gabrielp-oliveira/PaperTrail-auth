package utils

import (
	"github.com/gin-gonic/gin"
)

// CustomErrorResponse define a estrutura da resposta de erro personalizada
type CustomErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// RespondWithError é uma função utilitária para enviar respostas de erro personalizadas
func RespondWithError(ctx *gin.Context, statusCode int, message string, err error) {
	response := CustomErrorResponse{
		Message: message,
	}
	if err != nil {
		response.Error = err.Error()
	}
	ctx.JSON(statusCode, response)
}
