package main

import (
	"strings"

	"github.com/golangtest/bbd/scrapper"
	"github.com/labstack/echo"
)

func handleScrape(c echo.Context) error {
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	scrapper.Scrape(term)
	return c.Attachment("jobs.csv", "job.csv")
}

func HandleHome(c echo.Context) error {
	return c.File("home.html")
}

func main() {
	e := echo.New()
	e.GET("/", HandleHome)
	e.POST("/scrape", handleScrape)
	e.Logger.Fatal(e.Start(":1323"))
}
