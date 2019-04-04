package main

import (
	"context"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"strings"
	"testing"
)

func Test_parseRateFormat(t *testing.T) {
	for _, usercase := range []struct {
		in string
		out float64
	} {
		{"", 0},
		{"0", 0},
		{"1k", 1000},
		{"1m", 1000000},
		{"1b", 1000000000},
		{"1t", 1000000000000},
		{"2.4m", 2400000},
	} {
		r := parseRateFormat(usercase.in)
		if r != usercase.out {
			t.Errorf("%s: '%f' != '%f'", usercase.in, r, usercase.out)
			t.Fail()
		}
	}
}

func test_htmlProcess(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{}
	callHTML(ctx, client, "http://localhost:9000/clusters/default/brokers", "table", func(selection *goquery.Selection) {
		//s, _ := goquery.OuterHtml(selection.First().Find("td:nth-child(5)"))
		selection.First().Find("tr").EachWithBreak(func(i int, selection *goquery.Selection) bool {
			c := selection.Children()
			if i == 0 {
				if c.Eq(1).Text() != "Host" {
					return false
				}
				return true
			}

			log.Print(strings.TrimSpace(c.Eq(1).Text()))
			log.Print(strings.TrimSpace(c.Eq(4).Text()))
			log.Print(parseRateFormat(strings.TrimSpace(c.Eq(4).Text())))
			log.Print(strings.TrimSpace(c.Eq(5).Text()))
			log.Print(parseRateFormat(strings.TrimSpace(c.Eq(5).Text())))
			return true
		})
		log.Printf("=====")
		selection.Last().Find("tr").EachWithBreak(func(i int, selection *goquery.Selection) bool {
			c := selection.Children()
			if i == 0 {
				if c.Eq(0).Text() != "Rate" {
					return false
				}
				return true
			}

			label := strings.TrimSpace(c.Eq(0).Text())
			log.Print(label)
			log.Print(labelToClusterMetrics[label])
			log.Print(strings.TrimSpace(c.Eq(2).Text()))
			log.Print(parseRateFormat(strings.TrimSpace(c.Eq(2).Text())))
			return true
		})
	})
}
