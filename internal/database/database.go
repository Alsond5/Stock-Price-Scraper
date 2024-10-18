package database

import (
	"database/sql"
	"fmt"

	_ "github.com/microsoft/go-mssqldb"
)

type Stock struct {
	StockSymbol string
	StockName   string
	Price       float64
}

func createTempTable(tx *sql.Tx, tableName string) error {
	query := fmt.Sprintf(`
		CREATE TABLE %s (
			StockId INT IDENTITY(1, 1) PRIMARY KEY,
			StockSymbol NVARCHAR(10),
			StockName NVARCHAR(100),
			Price DECIMAL(18, 2),
			LastUpdated DATETIME
		);
	`, tableName)

	_, err := tx.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func upsert(db *sql.DB, stocks []Stock) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	tableName := "#tempStocks"

	err = createTempTable(tx, tableName)
	if err != nil {
		return err
	}

	insertQuery := fmt.Sprintf(`
		INSERT INTO %s (StockSymbol, StockName, Price, LastUpdated)
		VALUES (@StockSymbol, @StockName, @Price, GETDATE());
	`, tableName)

	for _, stock := range stocks {
		_, err = tx.Exec(insertQuery,
			sql.Named("StockSymbol", stock.StockSymbol),
			sql.Named("StockName", stock.StockName),
			sql.Named("Price", stock.Price))
		if err != nil {
			return err
		}
	}

	mergeQuery := fmt.Sprintf(`
		MERGE INTO Stocks AS target
		USING (SELECT StockSymbol, StockName, Price FROM %s) AS source
		ON target.StockSymbol = source.StockSymbol
		WHEN MATCHED THEN
			UPDATE SET 
				target.Price = source.Price,
				target.LastUpdated = GETDATE()
		WHEN NOT MATCHED THEN
			INSERT (StockSymbol, StockName, Price, Quantity, IsActive, LastUpdated)
			VALUES (source.StockSymbol, source.StockName, source.Price, 10000, 1, GETDATE());
	`, tableName)

	_, err = tx.Exec(mergeQuery)
	if err != nil {
		return err
	}

	insertHistoriesQuery := fmt.Sprintf(`
		INSERT INTO Histories (StockId, Price, Date)
		SELECT StockId, Price, LastUpdated
		FROM %s;
	`, tableName)

	_, err = tx.Exec(insertHistoriesQuery)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func Save(stocks []Stock) error {
	connectionString := "server=localhost;database=StockMarketDB;trusted_connection=true;trustservercertificate=true"

	db, err := sql.Open("sqlserver", connectionString)
	if err != nil {
		return err
	}

	defer db.Close()

	err = upsert(db, stocks)
	if err != nil {
		return err
	}

	defer db.Exec(fmt.Sprintf("DROP TABLE %s;", "#tempStocks"))

	return nil
}
