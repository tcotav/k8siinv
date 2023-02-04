package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/spf13/viper"
	"github.com/tcotav/k8siinv/services"
	"github.com/tcotav/k8siinv/types"
)

func saveImages(c echo.Context) error {
	clusterImages := new(types.ClusterInventory)
	if err := c.Bind(clusterImages); err != nil {
		return err
	}
	err := clusterImageService.SaveClusterInventory(clusterImages)
	if err != nil {
		return err
	}
	ret := types.HttpReturn{Status: fmt.Sprint(http.StatusCreated), Message: "OK"}
	return c.JSON(http.StatusCreated, ret)
}

var clusterImageService *services.ClusterInventoryService

type ServerConfig struct {
	DatabaseDSN  string `json:"databaseDSN"`
	HTTPListener struct {
		Port int `json:"port"`
	} `json:"httplistener"`
}

func main() {

	e := echo.New()
	e.Use(middleware.Logger())

	viper.SetConfigName("servconfig")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/k8siinv/")
	viper.AddConfigPath("$HOME/.k8siinv")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		e.Logger.Fatal("fatal error config file: %w", err)
	}

	var config ServerConfig
	viper.Unmarshal(&config)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/v1/images", saveImages)

	dsn := fmt.Sprintf("%s?parseTime=true", config.DatabaseDSN)
	db, err := openDB(dsn)
	if err != nil {
		e.Logger.Fatal(err)
	}

	clusterImageService = services.NewClusterInventoryService(db)

	// do our defer here for cleanup
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			e.Logger.Fatal(err)
		}
	}(db)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", config.HTTPListener.Port)))
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// give it a quick test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
