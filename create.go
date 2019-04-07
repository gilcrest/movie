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

// Create performs business validations prior to writing to the db
func (m *Movie) Create(ctx context.Context, log zerolog.Logger, tx *sql.Tx) error {
	const op errors.Op = "movie/Movie.Create"

	// Validate input data
	err := m.validate()
	if err != nil {
		if e, ok := err.(*errors.Error); ok {
			return errors.E(errors.Validation, e.Param, err)
		}
		// should not get here, but just in case
		return errors.E(errors.Validation, err)
	}

	// Pull client information from Server token and set
	createClient, err := apiclient.ViaServerToken(ctx, tx)
	if err != nil {
		return errors.E(op, errors.Internal, err)
	}
	m.CreateClient.Number = createClient.Number
	m.UpdateClient.Number = createClient.Number

	// Create the user record in the database
	err = m.createDB(ctx, log, tx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.E(op, errors.Database, err)
		}
		// Kind could be Database or Exist from db, so
		// use type assertion and send both up
		if e, ok := err.(*errors.Error); ok {
			return errors.E(e.Kind, e.Code, e.Param, err)
		}
		// Should not actually fall to here, but including as
		// good practice
		return errors.E(op, errors.Database, err)
	}

	if err := tx.Commit(); err != nil {
		return errors.E(op, errors.Database, err)
	}

	return nil
}

// Validate does basic input validation and ensures the struct is
// properly constructed
func (m *Movie) validate() error {
	const op errors.Op = "movie/Movie.validate"

	switch {
	case m.Title == "":
		return errors.E(op, errors.Validation, errors.Parameter("Title"), errors.MissingField("Title"))
	case m.Year < 1878:
		return errors.E(op, errors.Validation, errors.Parameter("Year"), "The first film was in 1878, Year must be >= 1878")
	case m.Rated == "":
		return errors.E(op, errors.Validation, errors.Parameter("Rated"), errors.MissingField("Rated"))
	case m.Released.IsZero() == true:
		return errors.E(op, errors.Validation, errors.Parameter("ReleaseDate"), "Released must have a value")
	case m.RunTime <= 0:
		return errors.E(op, errors.Validation, errors.Parameter("RunTime"), "Run time must be greater than zero")
	case m.Director == "":
		return errors.E(op, errors.Validation, errors.Parameter("Director"), errors.MissingField("Director"))
	case m.Writer == "":
		return errors.E(op, errors.Validation, errors.Parameter("Writer"), errors.MissingField("Writer"))
	}

	return nil
}

// CreateDB creates a record in the user table using a stored function
func (m *Movie) createDB(ctx context.Context, log zerolog.Logger, tx *sql.Tx) error {
	const op errors.Op = "movie/Movie.createDB"

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

	// Execute stored function that returns the create_date timestamp,
	// hence the use of QueryContext instead of Exec
	rows, err := stmt.QueryContext(ctx,
		m.Title,               //$1
		m.Year,                //$2
		m.Rated,               //$3
		m.Released,            //$4
		m.RunTime,             //$5
		m.Director,            //$6
		m.Writer,              //$7
		m.CreateClient.Number, //$8
		m.CreateUsername)      //$9

	if err != nil {
		return errors.E(op, err)
	}
	defer rows.Close()

	var (
		createTime time.Time
		updateTime time.Time
	)

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&createTime, &updateTime); err != nil {
			return errors.E(op, err)
		}
	}

	// If any error was encountered while iterating through rows.Next above
	// it will be returned here
	if err := rows.Err(); err != nil {
		return errors.E(op, err)
	}

	// set the dbaudit fields with timestamps from the database
	m.CreateTimestamp = createTime
	m.UpdateTimestamp = updateTime

	return nil
}
