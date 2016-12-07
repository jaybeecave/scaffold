package main

import (
	"net/http"

	"bytes"

	"bufio"

	"io/ioutil"

	"os"

	"github.com/jaybeecave/render"
	_ "github.com/mattes/migrate/driver/postgres" //for migrations
	"github.com/mattes/migrate/file"
	"github.com/mattes/migrate/migrate"
	"github.com/urfave/cli"

	"fmt"

	"strings"

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
	bucket.add("TableName", args.First())

	fields := Fields{}
	for _, arg := range args {
		fmt.Printf(arg)
		if args.First() == arg {
			continue // we dont care about the first arg as its the TableName
		}
		if strings.Contains(arg, ":") {
			strSlice := strings.Split(arg, ":")
			field := Field{
				FieldName: strSlice[0],
				FieldType: strSlice[1],
			}
			fields = append(fields, field)
		}
	}
	bucket.add("Fields", fields)

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
