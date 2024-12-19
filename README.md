## Organic Views counter in Go

An example how you can integrate viewscounter into Gin API. You should create a middleware handler

```go

package middlewares

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mikebionic/viewscount"
	"time"
)

var viewTracker *viewscount.ViewTracker

func InitializeViewTracker(db interface{}, minutesGap time.Duration) {
	viewTracker = viewscount.NewViewTracker(db, minutesGap)
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

In your main app you should import created middleware and configure the database (supports sql.DB and pgxscann.DB)

```go
func main() {
	//...
	middlewares.InitializeViewTracker(database.DB, 10)	

	// some routers controllers:
	router.GET("/:id", middlewares.ViewCounterMiddleware("tbl_driver"), services.GetDriver)
	//...
}
```