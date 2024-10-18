
# Stock Price Scraper

A Go-based web scraper that collects real-time stock prices from **[Hürriyet Bigpara](https://bigpara.hurriyet.com.tr/borsa/canli-borsa/)**. It automates data scraping at regular intervals, storing the results in a SQL Server database. This project is designed with modular components for scraping, scheduling, logging, and database operations.

## Features

-   **Real-time Stock Scraper:** Fetches stock prices from `https://bigpara.hurriyet.com.tr`.
-   **Scheduler:** Automates scraping every 3 minutes.
-   **Database Upsert Logic:** Uses SQL Server to store and update stock prices.
-   **Logger Module:** Provides color-coded logs for better readability.
-   **Graceful Shutdown:** Stops the scheduler safely on termination signals.

## Technologies

-   **Language:** Go
-   **Database:** SQL Server
-   **HTTP Scraping:** Native Go HTTP client
-   **Scheduling:** Custom-built job scheduler

## Getting Started

### Prerequisites

-   **Go 1.18+** installed
-   **SQL Server** running on your machine or accessible remotely
-   Basic knowledge of Go, SQL, and web scraping

### Project Structure
```
Stock-Price-Scraper/
│
├── internal/
│   ├── database/      # Database operations (SQL queries, temp tables, etc.)
│   ├── logger/        # Logger with ANSI color support
│   ├── scheduler/     # Custom job scheduler
│   └── scraper/       # Web scraper logic
├── main.go            # Entry point of the application
├── go.mod             # Go module dependencies
└── LICENSE            # License file
```

## Installation

1.  **Clone the repository:**
    
    ```bash
    git clone https://github.com/Alsond5/Stock-Price-Scraper.git
    cd Stock-Price-Scraper
    ``` 
    
2.  **Install dependencies:**
    
    ```bash
    go mod tidy
    ``` 
    
3.  **Set up SQL Server:**  
    Make sure your SQL Server is running, and update the connection string in `internal/database/database.go`:
    
    ```go
    connectionString := "server=localhost;database=StockMarketDB;trusted_connection=true;trustservercertificate=true"
    ```
    

## Usage

1.  **Run the application:**
    
    ```bash
    go run main.go
    ```
    
2.  **Scheduler Output:**  
    The application will scrape stock prices every 3 minutes and log the results to the console.
    
3.  **Stopping the Scheduler:**  
    Use `Ctrl+C` to stop the scheduler gracefully.
    

## Code Overview

### Scheduler Example

```go
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
```

### Logging Example

```go
logger.Info("This is an info message.")
logger.Success("Operation was successful!")
logger.Warning("This is a warning.")
logger.Error("An error occurred.")
```

### Scraper Logic

-   **Connection Setup:** Initializes the scraper by opening an HTTP connection.
-   **Scraping Data:** Collects stock symbols, names, and prices.
-   **Error Handling:** Logs errors if the scraping process fails.

### Database Operations

-   **Temp Tables:** Uses temp tables to store scraped stock data temporarily.
-   **Upsert Logic:** Merges the new data into the `Stocks` table, updating existing records or inserting new ones.

## License

This project is licensed under the **MIT License** – see the LICENSE file for details.

## Contributing

Contributions are welcome! Please open an issue or a pull request for any feature requests or improvements.

## Contact

For questions or collaboration inquiries, reach out through GitHub issues.