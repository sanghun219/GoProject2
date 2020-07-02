package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var baseURL = "https://kr.indeed.com/jobs?q=python"

type extractedJob struct {
	id       string
	title    string
	location string
	salary   string
	summary  string
}

func main() {
	var Jobs []extractedJob
	c := make(chan []extractedJob)
	TotalPages := GetPages()
	for i := 0; i < TotalPages; i++ {
		go getPage(i, c)

	}

	for i := 0; i < TotalPages; i++ {
		extractedJobs := <-c
		Jobs = append(Jobs, extractedJobs...)
	}

	writeJobs(Jobs)
	fmt.Println("Done, extracted jobs")
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	CheckErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"ID", "Title", "Location", "salary", "summary"}
	wErr := w.Write(headers)
	CheckErr(wErr)
	for _, job := range jobs {
		jobslice := []string{"kr.indeed.com/viewjobs?jk=" + job.id, job.title, job.location, job.salary, job.summary}
		go w.Write(jobslice)

	}
}

func GetPages() int {
	pages := 0
	res, err := http.Get(baseURL)
	CheckErr(err)
	CheckCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	CheckErr(err)
	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages
}

func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func CheckCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request Failed With Status", res.StatusCode)
	}
}

func getPage(page int, mainC chan<- []extractedJob) {
	var jobs []extractedJob
	pageURL := baseURL + "&start=" + strconv.Itoa(page*10)
	fmt.Println("Requesting", pageURL)
	res, err := http.Get(pageURL)
	CheckErr(err)
	CheckCode(res)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	CheckErr(err)
	c := make(chan extractedJob)
	searchCards := doc.Find(".jobsearch-SerpJobCard")
	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)

	})

	for i := 0; i < searchCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}
	mainC <- jobs
}
func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	id, _ := card.Attr("data-jk")
	title := CleanString(card.Find(".title>a").Text())
	location := CleanString(card.Find(".sjcl").Text())
	salary := CleanString(card.Find("salaryText").Text())
	summary := CleanString(card.Find(".summary").Text())
	c <- extractedJob{id: id, title: title, location: location, salary: salary, summary: summary}
}
func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")

}
