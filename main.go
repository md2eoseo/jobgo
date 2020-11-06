package main

import (
	"fmt"

	"github.com/md2eoseo/jobgo/scrapper"
)

func main() {
	var term string
	fmt.Scan(&term)
	scrapper.Scrape(term)
}
