package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var baseURL string = "https://kr.indeed.com/jobs?q=python&limit=50"

type jobObj struct {
	jk       string
	title    string
	company  string
	location string
	salary   string
	summary  string
}

func main() {
	var result []jobObj
	c := make(chan []jobObj)
	totalPages := getPages()

	for i := 0; i < totalPages; i++ {
		go getPage(i, c)
	}
	for i := 0; i < totalPages; i++ {
		pageResult := <-c
		result = append(result, pageResult...)
	}

	createCSV(result)
}

func createCSV(result []jobObj) {
	file, err := os.Create("jobs.csv")
	if err != nil {
		panic(err)
	}

	wr := csv.NewWriter(bufio.NewWriter(file))
	defer wr.Flush()

	headers := []string{"URL", "Title", "Company", "Location", "Salary", "Summary"}
	wr.Write(headers)

	for _, job := range result {
		err := wr.Write([]string{"https://kr.indeed.com/viewjob?jk=" + job.jk, job.title, job.company, job.location, job.salary, job.summary})
		checkErr(err)
	}

	fmt.Println("complete")
}

func extractJob(job *goquery.Selection) jobObj {
	jk, _ := job.Attr("data-jk")
	title := cleanString(job.Find(".jobtitle").Text())
	company := cleanString(job.Find(".company").Text())
	location := cleanString(job.Find(".location").Text())
	salary := cleanString(job.Find(".salaryText").Text())
	summary := cleanString(job.Find(".summary").Text())
	return jobObj{jk, title, company, location, salary, summary}
}

// TODO: goroutine here
func getPage(pageNum int, c chan []jobObj) {
	var result []jobObj
	pageURL := baseURL + "&start=" + strconv.Itoa(pageNum)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkStatusCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	jobs := doc.Find(".jobsearch-SerpJobCard")

	jobs.Each(func(i int, s *goquery.Selection) {
		result = append(result, extractJob(s))
	})

	c <- result
}

func getPages() int {
	pages := 0
	res, err := http.Get(baseURL)
	checkErr(err)
	checkStatusCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination-list").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("li").Length() - 1
	})
	// TODO: if there is no pagination-list class, pages is 1

	return pages
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkStatusCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status:", res.StatusCode)
	}
}

func cleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}
