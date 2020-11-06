package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/md2eoseo/jobgo/scrapper"
)

func main() {
	// Echo instance
	e := echo.New()

	// Routes
	e.GET("/", handleHome)
	e.POST("/search", handleSearch)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

func handleSearch(c echo.Context) error {
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	fileName := term + "_" + strconv.Itoa(time.Now().Year()) + "_" + time.Now().Month().String() + ".csv"
	scrapper.Scrape(term)
	defer os.Remove("jobs.csv")
	return c.Attachment("jobs.csv", fileName)
}

func handleHome(c echo.Context) error {
	return c.File("home.html")
}
