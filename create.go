// Package movie has the Movie struct and associated methods to
// create, modify, delete and get movies
package movie

import (
	"context"
	"database/sql"
	"time"

	"github.com/gilcrest/dbaudit"
	"github.com/gilcrest/errors"
	"github.com/gilcrest/servertoken"
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
		return errors.E(op, errors.MissingField("Title"))
	case m.Year < 1878:
		return errors.E(op, "The first film was in 1878, Year must be >= 1878")
	case len(m.Rated) == 0:
		return errors.E(op, errors.MissingField("Rated"))
	case m.Released.IsZero() == true:
		return errors.E(op, "Released must have a value")
	case m.RunTime <= 0:
		return errors.E(op, "Run time must be greater than zero")
	case len(m.Director) == 0:
		return errors.E(op, errors.MissingField("Director"))
	case len(m.Writer) == 0:
		return errors.E(op, errors.MissingField("Writer"))
	}

	return nil
}

// Create creates a Movie and stores it in the database
func (m *Movie) Create(ctx context.Context, log zerolog.Logger, tx *sql.Tx) (*sql.Tx, error) {
	const op errors.Op = "movie/Movie.Create"

	srvToken := servertoken.FromCtx(ctx)

	tx, err := m.createDB(ctx, log, tx, srvToken)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return tx, nil
}

// createDB creates a record in the user table using a stored function
func (m *Movie) createDB(ctx context.Context, log zerolog.Logger, tx *sql.Tx, srvrToken string) (*sql.Tx, error) {
	const op errors.Op = "movie/Movie.createDB"

	var (
		createTimestamp time.Time
	)

	// Prepare the sql statement using bind variables
	stmt, err := tx.PrepareContext(ctx, `select demo.create_movie (
		p_title => $1,
		p_year => $2,
		p_rated => $3,
		p_released => $4,
		p_run_time => $5,
		p_director => $6,
		p_writer => $7,
		p_create_server_token => $8,
		p_username => $9)`)

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Execute stored function that returns the create_date timestamp,
	// hence the use of QueryContext instead of Exec
	rows, err := stmt.QueryContext(ctx,
		m.Title,          //$1
		m.Year,           //$2
		m.Rated,          //$3
		m.Released,       //$4
		m.RunTime,        //$5
		m.Director,       //$6
		m.Writer,         //$7
		srvrToken,        //$8
		m.CreateUsername) //$9

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&createTimestamp); err != nil {
			return nil, err
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// set the CreateDate field to the create_date set as part of the insert in
	// the stored function call above
	m.CreateTimestamp = createTimestamp

	return tx, nil

}
