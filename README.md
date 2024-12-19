
Organic Views counter in Go

An example how to integrate into Gin API. You should create a middleware handler

```go

package middlewares

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

var viewTracker *ViewTracker

func InitializeViewTracker(db interface{}, minutesGap time.Duration) {
	viewTracker = NewViewTracker(db, minutesGap)
}

func ViewCounterMiddleware(tableName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.Next()
			return
		}

		idInt := 0
		fmt.Sscanf(id, "%d", &idInt)

		if err := viewTracker.HandleView(c.Request, tableName, idInt); err != nil {
			fmt.Printf("Error tracking view: %v\n", err)
		}
		c.Next()
	}
}


```