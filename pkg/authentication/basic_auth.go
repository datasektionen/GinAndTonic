package authentication

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func BasicAuthMiddleware(username, password string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, pass, hasAuth := c.Request.BasicAuth()
		println(user, pass, hasAuth)
		if hasAuth && user == username && pass == password {
			c.Next()
		} else {
			c.Header("WWW-Authenticate", `Basic realm="Metrics"`)
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
