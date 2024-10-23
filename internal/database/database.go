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

type Alert struct {
	AlertId           int
	LowerLimit        float64
	UpperLimit        float64
	WasTriggeredBelow bool
	WasTriggeredAbove bool
	StockSymbol       string
	StockName         string
	Username          string
	Email             string
}

func connect() (*sql.DB, error) {
	connectionString := "server=localhost;database=StockMarketDB;trusted_connection=true;trustservercertificate=true"

	db, err := sql.Open("sqlserver", connectionString)
	if err != nil {
		return nil, err
	}

	return db, nil
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
	db, err := connect()
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

func GetAlerts() ([]*Alert, error) {
	db, err := connect()
	if err != nil {
		return nil, err
	}

	defer db.Close()

	query := `
		SELECT
			Alerts.AlertId,
			Alerts.LowerLimit,
			Alerts.UpperLimit,
			Alerts.WasTriggeredBelow,
			Alerts.WasTriggeredAbove,
			Stocks.StockSymbol,
			Stocks.StockName,
			Users.Username,
			Users.Email
		FROM Alerts
		INNER JOIN Users
		ON Alerts.UserId = Users.UserId
		INNER JOIN Stocks
		ON Alerts.StockId = Stocks.StockId
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var alerts []*Alert
	for rows.Next() {
		var alert Alert
		err := rows.Scan(&alert.AlertId, &alert.LowerLimit, &alert.UpperLimit, &alert.WasTriggeredBelow, &alert.WasTriggeredAbove, &alert.StockSymbol, &alert.StockName, &alert.Username, &alert.Email)
		if err != nil {
			return nil, err
		}

		alerts = append(alerts, &alert)
	}

	return alerts, nil
}

func UpdateAlerts(alerts []*Alert) error {
	db, err := connect()
	if err != nil {
		return err
	}

	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		UPDATE Alerts
		SET
			WasTriggeredBelow = @WasTriggeredBelow,
			WasTriggeredAbove = @WasTriggeredAbove
		WHERE AlertId = @AlertId;
	`)
	if err != nil {
		tx.Rollback()
		return err
	}

	defer stmt.Close()

	for _, alert := range alerts {
		_, err := stmt.Exec(
			sql.Named("WasTriggeredBelow", alert.WasTriggeredBelow),
			sql.Named("WasTriggeredAbove", alert.WasTriggeredAbove),
			sql.Named("AlertId", alert.AlertId))
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
