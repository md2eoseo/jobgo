package scrapper

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

var limit string = "50"

type jobObj struct {
	jk       string
	title    string
	company  string
	location string
	salary   string
	summary  string
}

func Scrape(term string) {
	var baseURL string = "https://kr.indeed.com/jobs?" + "q=" + term + "&limit=" + limit
	var result []jobObj
	c := make(chan []jobObj)
	totalPages := getPages(baseURL)

	for i := 0; i < totalPages; i++ {
		go getPage(i, baseURL, c)
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

func extractJob(job *goquery.Selection, c chan<- jobObj) {
	jk, _ := job.Attr("data-jk")
	title := CleanString(job.Find(".jobtitle").Text())
	company := CleanString(job.Find(".company").Text())
	location := CleanString(job.Find(".location").Text())
	salary := CleanString(job.Find(".salaryText").Text())
	summary := CleanString(job.Find(".summary").Text())
	c <- jobObj{jk, title, company, location, salary, summary}
}

func getPage(pageNum int, baseURL string, mainC chan<- []jobObj) {
	var result []jobObj
	c := make(chan jobObj)
	limitInt, _ := strconv.Atoi(limit)
	pageURL := baseURL + "&start=" + strconv.Itoa(pageNum*limitInt)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkStatusCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	jobs := doc.Find(".jobsearch-SerpJobCard")

	jobs.Each(func(i int, s *goquery.Selection) {
		go extractJob(s, c)

	})

	for i := 0; i < jobs.Length(); i++ {
		job := <-c
		result = append(result, job)
	}

	mainC <- result
}

func getPages(baseURL string) int {
	pages := 0
	res, err := http.Get(baseURL)
	checkErr(err)
	checkStatusCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	pages = doc.Find(".pagination-list li").Length() - 1

	if pages < 0 {
		pages = 1
	}

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

func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}
