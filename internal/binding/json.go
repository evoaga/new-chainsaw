package binding

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// BindJSON binds JSON payload to the given object and handles errors.
func BindJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		log.Printf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return err
	}
	return nil
}
