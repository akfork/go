package horizon

import (
	"github.com/stellar/go/services/horizon/internal/db2/core"
	"github.com/stellar/go/services/horizon/internal/db2/history"
	"github.com/stellar/go/support/db"
	"github.com/stellar/go/support/log"
)

func initHorizonDb(app *App) {
	session, err := db.Open("postgres", app.config.DatabaseURL)
	if err != nil {
		log.Panic(err)
	}

	session.DB.SetMaxIdleConns(app.config.HorizonDBMaxIdleConnections)
	session.DB.SetMaxOpenConns(app.config.HorizonDBMaxOpenConnections)

	app.historyQ = &history.Q{session}
}

func initCoreDb(app *App) {
	session, err := db.Open("postgres", app.config.StellarCoreDatabaseURL)
	if err != nil {
		log.Panic(err)
	}

	session.DB.SetMaxIdleConns(app.config.CoreDBMaxIdleConnections)
	session.DB.SetMaxOpenConns(app.config.CoreDBMaxOpenConnections)

	app.coreQ = &core.Q{session}
}

func init() {
	appInit.Add("horizon-db", initHorizonDb, "app-context", "log")
	appInit.Add("core-db", initCoreDb, "app-context", "log")
}
