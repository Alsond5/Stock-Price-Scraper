package alert

import (
	"fmt"
	"os"

	"github.com/Alsond5/StockMarketAPIWebScraper/internal/database"
	"github.com/Alsond5/StockMarketAPIWebScraper/internal/dotenv"
	"github.com/Alsond5/StockMarketAPIWebScraper/pkg/email"
)

var headers []string = []string{"MIME-Version: 1.0;", "Content-Type: text/html; charset=\"UTF-8\";"}

var htmlContent string = `<!DOCTYPE html>
<html lang="tr">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: Arial, sans-serif; background-color: #f4f4f4;">
	<div style="width: 100%%; max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 8px; overflow: hidden;">
		<div style="background-color: #007BFF; color: white; padding: 20px; text-align: center;">
			<h1 style="margin: 0;">%s Price Alert</h1>
		</div>
		<div style="padding: 20px;">
			<p>Hello %s,</p>
			<p>%s (%s) is now %s %.2f TL</p>
			<div style="background-color: %s; color: %s; padding: 10px; border-left: 6px solid %s; margin: 20px 0; border-radius: 4px;">
				Currently at <strong>%.2f</strong> TL
			</div>
		</div>
		<div style="text-align: center; padding: 20px; font-size: 12px; color: #777777;">
			<p>This email was sent automatically from our price tracking system.</p>
			<p>&copy; 2024 Price Tracking System</p>
		</div>
	</div>
</body>
</html>
`

func SendAlerts(stockMap map[string]database.Stock, alerts []*database.Alert) error {
	err := dotenv.LoadEnv()
	if err != nil {
		return err
	}

	username := os.Getenv("EMAIL_USERNAME")
	password := os.Getenv("EMAIL_PASSWORD")

	config := email.NewConfigurations("smtp.gmail.com", 587, username, password)

	for _, alert := range alerts {
		stock, ok := stockMap[alert.StockSymbol]
		if !ok {
			continue
		}

		if stock.Price <= alert.LowerLimit && !alert.WasTriggeredBelow {
			emailContent := fmt.Sprintf(htmlContent, stock.StockSymbol, alert.Username, stock.StockName, stock.StockSymbol, "below", alert.LowerLimit, "#ffdddd", "#d8000c", "#d8000c", stock.Price)

			newEmail := email.NewEmail(alert.Email, fmt.Sprintf("%s Price Alert", stock.StockName), emailContent, headers)

			err := newEmail.Send(config)
			if err != nil {
				return err
			}

			alert.WasTriggeredBelow = true
			alert.WasTriggeredAbove = false
		} else if stock.Price >= alert.UpperLimit && !alert.WasTriggeredAbove {
			emailContent := fmt.Sprintf(htmlContent, stock.StockSymbol, alert.Username, stock.StockName, stock.StockSymbol, "above", alert.UpperLimit, "#d5edda", "#158724", "#158724", stock.Price)

			newEmail := email.NewEmail(alert.Email, fmt.Sprintf("%s Price Alert", stock.StockName), emailContent, headers)

			err := newEmail.Send(config)
			if err != nil {
				return err
			}

			alert.WasTriggeredAbove = true
			alert.WasTriggeredBelow = false
		} else if stock.Price > alert.LowerLimit && stock.Price < alert.UpperLimit {
			alert.WasTriggeredBelow = false
			alert.WasTriggeredAbove = false
		}
	}

	return nil
}
