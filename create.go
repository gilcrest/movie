// Package movie has the Movie struct and associated methods to
// create, modify, delete and get movies
package movie

import (
	"context"
	"database/sql"
	"time"

	"github.com/gilcrest/apiclient"
	"github.com/gilcrest/dbaudit"
	"github.com/gilcrest/errors"
	"github.com/rs/zerolog"
)

// Movie holds details of a movie
type Movie struct {
	Title    string
	Year     int
	Rated    string
	Released time.Time
	RunTime  int
	Director string
	Writer   string
	dbaudit.Audit
}

// Validate does basic input validation and ensures the struct is
// properly constructed
func (m *Movie) Validate() error {
	const op errors.Op = "movie/Movie.validate"

	switch {
	case len(m.Title) == 0:
		return errors.E(op, errors.Validation, errors.Parameter("Title"), errors.MissingField("Title"))
	case m.Year < 1878:
		return errors.E(op, errors.Validation, errors.Parameter("Title"), "The first film was in 1878, Year must be >= 1878")
	case len(m.Rated) == 0:
		return errors.E(op, errors.Validation, errors.Parameter("Rated"), errors.MissingField("Rated"))
	case m.Released.IsZero() == true:
		return errors.E(op, errors.Validation, errors.Parameter("ReleaseDate"), "Released must have a value")
	case m.RunTime <= 0:
		return errors.E(op, errors.Validation, errors.Parameter("RunTime"), "Run time must be greater than zero")
	case len(m.Director) == 0:
		return errors.E(op, errors.Validation, errors.Parameter("Director"), errors.MissingField("Director"))
	case len(m.Writer) == 0:
		return errors.E(op, errors.Validation, errors.Parameter("Writer"), errors.MissingField("Writer"))
	}

	return nil
}

// createDB creates a record in the user table using a stored function
func (m *Movie) createDB(ctx context.Context, log zerolog.Logger, tx *sql.Tx) error {
	const op errors.Op = "movie/Movie.createDB"

	createClient, err := apiclient.ViaServerToken(ctx, tx)
	if err != nil {
		return errors.E(op, err)
	}

	// Prepare the sql statement using bind variables
	stmt, err := tx.PrepareContext(ctx, `
	select o_create_timestamp,
		   o_update_timestamp
	  from demo.create_movie (
		p_title => $1,
		p_year => $2,
		p_rated => $3,
		p_released => $4,
		p_run_time => $5,
		p_director => $6,
		p_writer => $7,
		p_create_client_num => $8,
		p_create_username => $9)`)

	if err != nil {
		return errors.E(op, err)
	}
	defer stmt.Close()

	var (
		createTime time.Time
		updateTime time.Time
	)

	// Execute stored function that returns the create_date timestamp,
	// hence the use of QueryContext instead of Exec
	rows, err := stmt.QueryContext(ctx,
		m.Title,             //$1
		m.Year,              //$2
		m.Rated,             //$3
		m.Released,          //$4
		m.RunTime,           //$5
		m.Director,          //$6
		m.Writer,            //$7
		createClient.Number, //$8
		"gilcrest")          //$9

	if err != nil {
		return errors.E(op, err)
	}
	defer rows.Close()

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&createTime, &updateTime); err != nil {
			return errors.E(op, err)
		}
	}

	if err := rows.Err(); err != nil {
		return errors.E(op, err)
	}

	// set the dbaudit fields
	m.CreateClient = *createClient
	m.CreateTimestamp = createTime
	m.UpdateClient = *createClient
	m.UpdateTimestamp = updateTime

	return nil

}
