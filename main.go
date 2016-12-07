package main

import (
	"database/sql"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	dat "gopkg.in/mgutz/dat.v1"
	runner "gopkg.in/mgutz/dat.v1/sqlx-runner"

	"github.com/jaybeecave/render"
	dotenv "github.com/joho/godotenv"
	log "github.com/mgutz/logxi/v1"
	"github.com/urfave/cli"
)

func main() {
	dotenv.Load() // load from .env file where scaffold is run
	render := getRenderer()
	db := getDBConnection()
	app := cli.NewApp()
	app.Name = "scaffold"
	app.Usage = "generate models & migrations using dat"
	app.Commands = []cli.Command{
		{
			Name:    "table",
			Aliases: []string{"t"},
			Usage:   "Create a new table [tablename] [fieldname:fieldtype]",
			Action: func(c *cli.Context) error {
				return createTable(c, render, db)
			},
		},
	}

	app.Run(os.Args)
}

func getRenderer() *render.Render {
	r := render.New(render.Options{
		Directory: "./models/templates",
	})
	return r
}

func getDBConnection() *runner.DB {
	//get url from ENV in the following format postgres://user:pass@192.168.8.8:5432/spaceio")
	dbURL := os.Getenv("DATABASE_URL")
	u, err := url.Parse(dbURL)
	if err != nil {
		panic(err)
	}

	username := u.User.Username()
	pass, isPassSet := u.User.Password()
	if !isPassSet {
		log.Error("no database password")
	}
	host, port, _ := net.SplitHostPort(u.Host)
	dbName := strings.Replace(u.Path, "/", "", 1)

	db, _ := sql.Open("postgres", "dbname="+dbName+" user="+username+" password="+pass+" host="+host+" port="+port+" sslmode=disable")
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	log.Info("database running")
	// ensures the database can be pinged with an exponential backoff (15 min)
	runner.MustPing(db)

	// set to reasonable values for production
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(16)

	// set this to enable interpolation
	dat.EnableInterpolation = true

	// set to check things like sessions closing.
	// Should be disabled in production/release builds.
	dat.Strict = false

	// Log any query over 10ms as warnings. (optional)
	runner.LogQueriesThreshold = 10 * time.Millisecond

	// db connection
	return runner.NewDB(db, "postgres")
}
