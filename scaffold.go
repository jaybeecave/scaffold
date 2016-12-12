package main

import (
	"net/http"

	"bytes"

	"bufio"

	"io/ioutil"

	"os"

	"github.com/jaybeecave/render"
	errors "github.com/kataras/go-errors"
	_ "github.com/mattes/migrate/driver/postgres" //for migrations
	"github.com/mattes/migrate/file"
	"github.com/mattes/migrate/migrate"
	"github.com/urfave/cli"

	runner "gopkg.in/mgutz/dat.v1/sqlx-runner"
)

type description struct {
	Name        string
	Method      string
	URL         string
	Description string
	Function    http.HandlerFunc
}

type descriptions []description

func (slice descriptions) Len() int {
	return len(slice)
}

func (slice descriptions) Less(i int, j int) bool {
	return slice[i].Name < slice[j].Name
}

func (slice descriptions) Swap(i int, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type Field struct {
	FieldName string
	FieldType string
}

type Fields []Field

func createTable(c *cli.Context, r *render.Render, db *runner.DB) error {
	// setup
	bucket := newViewBucket()
	args := c.Args()

	if !args.Present() {
		// no args
		return cli.NewExitError("ERROR: No tablename defined", 1)
	}

	// add variables for template
	bucket.addFieldDataFromContext(c)

	file, err := migrate.Create(os.Getenv("DATABASE_URL")+"?sslmode=disable", "./models/migrations", "create_"+bucket.getStr("TableName"))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	err = fromTemplate(r, "create-table", file.UpFile, bucket)
	if err != nil {
		// render.JSON(w, http.StatusInternalServerError, err.Error())
		return cli.NewExitError(err.Error(), 1)
	}
	err = fromTemplate(r, "drop-table", file.DownFile, bucket)
	if err != nil {
		// render.JSON(w, http.StatusInternalServerError, err.Error())
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func addFields(c *cli.Context, r *render.Render, db *runner.DB) error {
	// setup
	bucket := newViewBucket()
	if !c.Args().Present() {
		// no args
		return cli.NewExitError("ERROR: No tablename defined", 1)
	}

	// add variables for template
	bucket.addFieldDataFromContext(c)

	file, err := migrate.Create(os.Getenv("DATABASE_URL")+"?sslmode=disable", "./models/migrations", "fields_"+bucket.getStr("TableName"))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	err = fromTemplate(r, "add-fields", file.UpFile, bucket)
	if err != nil {
		// render.JSON(w, http.StatusInternalServerError, err.Error())
		return cli.NewExitError(err.Error(), 1)
	}
	err = fromTemplate(r, "remove-fields", file.DownFile, bucket)
	if err != nil {
		// render.JSON(w, http.StatusInternalServerError, err.Error())
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func doMigration(c *cli.Context, r *render.Render, db *runner.DB) error {
	errs, ok := migrate.UpSync(os.Getenv("DATABASE_URL")+"?sslmode=disable", "./models/migrations")
	finalError := ""
	if !ok {
		for _, err := range errs {
			finalError += err.Error() + "\n"
		}
		return errors.New(finalError)
	}
	return nil
}

func fromTemplate(r *render.Render, templateName string, file *file.File, data *viewBucket) error {
	template := r.TemplateLookup(templateName)
	buffer := bytes.NewBuffer(file.Content)
	wr := bufio.NewWriter(buffer)
	err := template.Execute(wr, data)
	if err != nil {
		return err
	}
	wr.Flush()
	err = ioutil.WriteFile(file.Path+"/"+file.FileName, buffer.Bytes(), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
