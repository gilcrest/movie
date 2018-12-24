// Package movie has the Movie struct and associated methods to
// create, modify, delete and get movies
package movie

import (
	"fmt"
	"time"

	"github.com/gilcrest/dbaudit"
	"github.com/gilcrest/errors"
)

// // ErrNoUser is an error when a user is passed
// // that does not exist in the db
// var ErrNoUser = errors.Str("User does not exist")

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

func (m *Movie) validate() error {
	const op errors.Op = "movie/validate"

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
	default:
		fmt.Println("This movie is so valid!!!!")
	}

	return nil
}

// func (u *User) hashPassword() error {
// 	const op errors.Op = "usr.User.hashPassword"

// 	// Salt and hash the password using the bcrypt algorithm
// 	passHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), 8)
// 	if err != nil {
// 		return err
// 	}

// 	u.Password = string(passHash)

// 	return nil
// }

// func (u *User) validateEmail() error {
// 	const op errors.Op = "usr.User.validateEmail"
// 	_, err := mail.ParseAddress(u.Email)
// 	if err != nil {
// 		return errors.E(op, err)
// 	}
// 	return nil
// }

// // Create performs business validations prior to writing to the db
// func (u *User) Create(ctx context.Context, log zerolog.Logger) error {
// 	const op errors.Op = "usr.User.Create"

// 	err := u.validate()
// 	if err != nil {
// 		return errors.E(op, err)
// 	}

// 	err = u.hashPassword()

// 	return nil
// }

// // CreateDB creates a record in the user table using a stored function
// func (u *User) CreateDB(ctx context.Context, log zerolog.Logger, tx *sql.Tx) error {
// 	const op errors.Op = "usr.User.CreateDB"

// 	var (
// 		updateTimestamp time.Time
// 	)

// 	// Prepare the sql statement using bind variables
// 	stmt, err := tx.PrepareContext(ctx, `select auth.create_user (
// 		p_pgm => $1,
// 		p_username => $2,
// 		p_password => $3,
// 		p_first_name => $4,
// 		p_last_name => $5,
// 		p_email => $6,
// 		p_mobile_id => $7,
// 		p_client_id => $8,
// 		p_create_username => $9)`)

// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()

// 	// Execute stored function that returns the create_date timestamp,
// 	// hence the use of QueryContext instead of Exec
// 	rows, err := stmt.QueryContext(ctx,
// 		0,                //$1
// 		u.Username,       //$2
// 		u.Password,       //$3
// 		u.FirstName,      //$4
// 		u.LastName,       //$5
// 		u.Email,          //$6
// 		u.MobileID,       //$7
// 		u.CreateClientID, //$8
// 		u.CreateUsername) //$9

// 	if err != nil {
// 		return err
// 	}
// 	defer rows.Close()

// 	// Iterate through the returned record(s)
// 	for rows.Next() {
// 		if err := rows.Scan(&updateTimestamp); err != nil {
// 			return err
// 		}
// 	}

// 	if err := rows.Err(); err != nil {
// 		return err
// 	}

// 	// set the CreateDate field to the create_date set as part of the insert in
// 	// the stored function call above
// 	u.UpdateTimestamp = updateTimestamp

// 	return nil

// }

// // UserFromUsername constructs a User given a username
// func UserFromUsername(ctx context.Context, log zerolog.Logger, tx *sql.Tx, username string) (*User, error) {
// 	const op errors.Op = "appuser.UserFromUsername"

// 	// Prepare the sql statement using bind variables
// 	row := tx.QueryRowContext(ctx,
// 		`select username,
// 				password,
// 				mobile_id,
// 				email_address,
// 				first_name,
// 				last_name,
// 				update_client_id,
// 				update_user_id,
// 				update_timestamp
//   		   from demo.user
//           where username = $1`, username)

// 	usr := new(User)
// 	err := row.Scan(&usr.Username,
// 		&usr.Password,
// 		&usr.MobileID,
// 		&usr.Email,
// 		&usr.FirstName,
// 		&usr.LastName,
// 		&usr.UpdateClientID,
// 		&usr.UpdateUsername,
// 		&usr.UpdateTimestamp,
// 	)
// 	if err == sql.ErrNoRows {
// 		return nil, errors.E(op, ErrNoUser)
// 	} else if err != nil {
// 		return nil, errors.E(op, err)
// 	}

// 	return usr, nil
// }
