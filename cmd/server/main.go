package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/tcotav/imginv/services"
	"github.com/tcotav/imginv/types"
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

func main() {
	dsn := "imginv:password@tcp(127.0.0.1:3306)/testimginv?parseTime=true"
	e := echo.New()
	e.Use(middleware.Logger())
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/v1/images", saveImages)

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

	e.Logger.Fatal(e.Start(":8080"))
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
