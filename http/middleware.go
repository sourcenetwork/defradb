package http

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
)

// DatabaseMiddleware sets the database for the current request context.
func DatabaseMiddleware(db client.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

// TransactionMiddleware sets the transaction for the current request context.
func TransactionMiddleware(txs *sync.Map) gin.HandlerFunc {
	return func(c *gin.Context) {
		txValue := c.GetHeader(txHeaderName)
		if txValue == "" {
			c.Next()
			return
		}
		id, err := strconv.ParseUint(txValue, 10, 64)
		if err != nil {
			c.Next()
			return
		}
		tx, ok := txs.Load(id)
		if !ok {
			c.Next()
			return
		}

		c.Set("tx", tx)
		c.Next()
	}
}

func LensMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := c.MustGet("db").(client.DB)

		tx, ok := c.Get("tx")
		if ok {
			c.Set("lens", db.LensRegistry().WithTxn(tx.(datastore.Txn)))
		} else {
			c.Set("lens", db.LensRegistry())
		}

		c.Next()
	}
}

// StoreMiddleware sets the store for the current request
func StoreMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := c.MustGet("db").(client.DB)

		tx, ok := c.Get("tx")
		if ok {
			c.Set("store", db.WithTxn(tx.(datastore.Txn)))
		} else {
			c.Set("store", db)
		}

		c.Next()
	}
}

// CollectionMiddleware sets the collection for the current request context.
func CollectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := c.MustGet("db").(client.DB)

		col, err := db.GetCollectionByName(c.Request.Context(), c.Param("name"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		tx, ok := c.Get("tx")
		if ok {
			col = col.WithTxn(tx.(datastore.Txn))
		}

		c.Set("col", col)
		c.Next()
	}
}
