## Organic Views counter in Go

Imagine you have tables with a **view_count** column, but you don't want to write separate api for them and you want to prevent DOS type increment. This library is a solution for you.

```sql
CREATE TABLE tbl_driver
(
    id              SERIAL PRIMARY KEY,
    first_name      VARCHAR(100) NOT NULL DEFAULT '',
    view_count      INT          NOT NULL DEFAULT 0,
    deleted         INT          NOT NULL DEFAULT 0
);

CREATE TABLE tbl_vehicle
(
    id              SERIAL PRIMARY KEY,
    numberplate      VARCHAR(100) NOT NULL DEFAULT '',
    view_count      INT          NOT NULL DEFAULT 0,
    deleted         INT          NOT NULL DEFAULT 0
);


```

---

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

![](https://go.dev/blog/gophergala/fancygopher.jpg)
