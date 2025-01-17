package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/trustwallet/blockatlas/observer"
	observerStorage "github.com/trustwallet/blockatlas/observer/storage"
	"github.com/trustwallet/blockatlas/platform"
	"net/http"
	"strconv"
)

func setupObserverAPI(router gin.IRouter) {
	router.Use(requireAuth)
	router.POST("/webhook/register", addCall)
	router.DELETE("/webhook/register", deleteCall)
	router.GET("/status", statusCall)
}

func requireAuth(c *gin.Context) {
	auth := fmt.Sprintf("Bearer %s", viper.GetString("observer.auth"))
	if c.GetHeader("Authorization") == auth {
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

func addCall(c *gin.Context) {
	var req struct {
		Subscriptions map[string][]string `json:"subscriptions"`
		Webhook       string              `json:"webhook"`
	}
	if c.BindJSON(&req) != nil {
		return
	}

	if len(req.Subscriptions) == 0 {
		c.String(http.StatusOK, "Added")
		return
	}

	var subs []observer.Subscription
	for coinStr, perCoin := range req.Subscriptions {
		coin, _ := strconv.Atoi(coinStr)
		if coin == 0 {
			continue
		}
		for _, addr := range perCoin {
			subs = append(subs, observer.Subscription{
				Coin:     uint(coin),
				Address:  addr,
				Webhooks: []string{req.Webhook},
			})
		}
	}

	err := observerStorage.App.Add(subs)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.String(http.StatusOK, "Added")
}

func deleteCall(c *gin.Context) {
	var req struct {
		Subscriptions map[string][]string `json:"subscriptions"`
		Webhook       string              `json:"webhook"`
	}
	if c.BindJSON(&req) != nil {
		return
	}

	if len(req.Subscriptions) == 0 {
		c.String(http.StatusOK, "Deleted")
		return
	}

	var subs []observer.Subscription
	for coinStr, perCoin := range req.Subscriptions {
		coin, _ := strconv.Atoi(coinStr)
		if coin == 0 {
			continue
		}
		for _, addr := range perCoin {
			subs = append(subs, observer.Subscription{
				Coin:     uint(coin),
				Address:  addr,
				Webhooks: []string{req.Webhook},
			})
		}
	}

	err := observerStorage.App.Delete(subs)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.String(http.StatusOK, "Deleted")
}

func statusCall(c *gin.Context) {
	type coinStatus struct {
		Height int64  `json:"height"`
		Error  string `json:"error,omitempty"`
	}

	result := make(map[string]coinStatus)

	for _, api := range platform.BlockAPIs {
		coin := api.Coin()
		num, err := observerStorage.App.GetBlockNumber(coin.ID)
		var status coinStatus
		if err != nil {
			status = coinStatus{Error: err.Error()}
		} else if num == 0 {
			status = coinStatus{Error: "no blocks"}
		} else {
			status = coinStatus{Height: num}
		}
		result[coin.Handle] = status
	}

	c.JSON(http.StatusOK, result)
}
