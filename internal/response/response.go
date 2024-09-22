package response

import (
	"github.com/gin-gonic/gin"
	"log"
	"new-chainsaw/internal/config"
)

func JSONResponse(c *gin.Context, status int, message string, data interface{}, err error) {
	response := gin.H{
		"message": message,
		"data":    data,
	}

	if err != nil {
		if config.EnvVars["ENV"] == "production" {
			response["error"] = err.Error()
		}
		log.Println(err)
	}

	c.JSON(status, response)
}

func LogErrorAndRespond(c *gin.Context, statusCode int, logMessage string, errorMessage string, err error) {
	log.Println(logMessage)
	JSONResponse(c, statusCode, errorMessage, nil, err)
}
