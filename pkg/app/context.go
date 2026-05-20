package app

import "github.com/gin-gonic/gin"

// FromContext retrieves *App from a gin.Context. Returns nil if missing.
func FromContext(c *gin.Context) *App {
	v, ok := c.Get("app")
	if !ok {
		return nil
	}
	a, _ := v.(*App)
	return a
}
