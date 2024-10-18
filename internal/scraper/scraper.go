package scraper

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Alsond5/StockMarketAPIWebScraper/internal/database"
)

type Scraper struct {
	url    string
	client *http.Client
}

func NewScraper(url string) *Scraper {
	jar, _ := cookiejar.New(nil)

	return &Scraper{
		url: url,
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:      10,
				IdleConnTimeout:   30 * time.Second,
				DisableKeepAlives: false,
			},
			Jar: jar,
		},
	}
}

func (s *Scraper) Connection() error {
	URL, err := url.Parse(s.url)
	if err != nil {
		return err
	}

	rootURL := URL.Scheme + "://" + URL.Host

	req, err := http.NewRequest(http.MethodGet, rootURL, nil)
	if err != nil {
		return err
	}

	res, err := s.client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("Failed to connect to the website with status code: " + strconv.Itoa(res.StatusCode))
	}

	return nil
}

func (s *Scraper) Scrape() ([]database.Stock, error) {
	req, err := http.NewRequest(http.MethodGet, s.url, nil)
	if err != nil {
		return []database.Stock{}, err
	}

	res, err := s.client.Do(req)
	if err != nil {
		return []database.Stock{}, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return []database.Stock{}, errors.New("Failed to scrape the website with status code: " + strconv.Itoa(res.StatusCode))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return []database.Stock{}, err
	}

	htmlContent := string(body)

	htmlContent = regexp.MustCompile(`(?m)^[ \t]*\n`).ReplaceAllString(htmlContent, "")
	htmlContent = regexp.MustCompile(`>[ \t\n\r]*<`).ReplaceAllString(htmlContent, "><")

	tbody := regexp.MustCompile(`<div[^>]*class="tBody\s+ui-unsortable"[^>]*>([\s\S]*?)<\/div>`).FindStringSubmatch(htmlContent)

	stocksContent := regexp.MustCompile(`<ul[^>]*>([\s\S]*?)<\/ul>`).FindAllStringSubmatch(tbody[1], -1)

	liPattern := regexp.MustCompile(`<li[^>]*>([\s\S]*?)<\/li>`)
	aPattern := regexp.MustCompile(`<a[^>]*>([\s\S]*?)<\/a>`)
	aHrefPattern := regexp.MustCompile(`<a[^>]*href="([^"]*)"[^>]*>([^<]*)<\/a>`)

	stocks := make([]database.Stock, 0)

	for _, stockContent := range stocksContent {
		liContent := stockContent[1]
		liContents := liPattern.FindAllStringSubmatch(liContent, -1)

		liContents = liContents[:len(liContents)-1]

		aContent := aPattern.FindAllStringSubmatch(liContents[0][1], -1)
		stockSymbol := strings.TrimSpace(aContent[1][1])

		aHref := aHrefPattern.FindAllStringSubmatch(liContents[0][1], -1)[1][1]
		stockName := regexp.MustCompile(fmt.Sprintf(`%s-(.*?)-detay`, strings.ToLower(stockSymbol))).FindStringSubmatch(strings.TrimSpace(aHref))[1]
		stockName = strings.ToUpper(strings.ReplaceAll(stockName, "-", " "))

		priceString := strings.Replace(strings.TrimSpace(liContents[1][1]), ",", ".", 1)
		price, err := strconv.ParseFloat(priceString, 64)

		if err != nil {
			continue
		}

		stock := database.Stock{
			StockSymbol: stockSymbol,
			StockName:   stockName,
			Price:       price,
		}

		stocks = append(stocks, stock)
	}

	return stocks, nil
}

func (s *Scraper) Close() {
	s.client.CloseIdleConnections()
}
