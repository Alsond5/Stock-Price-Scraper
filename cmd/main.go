package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alsond5/StockMarketAPIWebScraper/internal/database"
	"github.com/Alsond5/StockMarketAPIWebScraper/internal/logger"
	"github.com/Alsond5/StockMarketAPIWebScraper/internal/scheduler"
	"github.com/Alsond5/StockMarketAPIWebScraper/internal/scraper"
)

func main() {
	sourceUrl := "https://bigpara.hurriyet.com.tr/borsa/canli-borsa/"

	scraper := scraper.NewScraper(sourceUrl)
	s := scheduler.NewScheduler()

	defer scraper.Close()

	scraper.Connection()

	s.AddJob(3*time.Minute, func() {
		logger.Info("Scraping stocks...")
		stocks, err := scraper.Scrape()
		if err != nil {
			logger.Error(err.Error())
			return
		}

		err = database.Save(stocks)
		if err != nil {
			logger.Error(err.Error())
			return
		}

		logger.Success("Stocks have been upserted successfully.")
	})

	go s.Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Press Ctrl+C to stop the scheduler.")
	<-sigs

	logger.Warning("Stopping the scheduler...")
	s.Stop()
	logger.Success("The scheduler has been stopped.")
}
