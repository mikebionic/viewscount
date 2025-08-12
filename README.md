## Organic Views counter in Go

**A unified, secure, and framework-agnostic view counter for any Golang web application.**
Tracks organic views across multiple tables without creating separate APIs for each one.
Protects against DoS-style repeated increments while remaining lightweight and easy to integrate.

![ViewsCounter](https://miro.medium.com/v2/resize:fit:720/format:webp/1*gvBUK2uyc335LpHQJzbiDQ.png)

## Why ViewCounter?

In many applications, you need to track views for different entities — articles, videos, products, or profiles.
Traditionally, you’d have to write **separate APIs** for each table with a `view_count` column.
This approach is repetitive, error-prone, and hard to maintain.

**ViewCounter solves this by:**

* Providing **one unified middleware** to track views for any table.
* Working with **any Golang web framework** (Gin, Chi, net/http, etc.).
* Offering **DoS protection** by ignoring rapid repeat requests from the same user/device.
* Being **database-agnostic** (PostgreSQL support included).

[Medium Article](https://medium.com/@mecreate/how-i-developed-the-viewscount-library-solving-the-problem-of-organic-view-counting-in-golang-0b1f6e1b19d7)

---

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
