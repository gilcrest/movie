# movie

This repo is a demo repo, holding the business logic for CRUD operations (Create, Read, Update, Delete) of a movie database.

The main business struct used in this module is Movie.

```go
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
```

Embedded within the type, is the `Audit` struct from the [dbaudit module](https://github.com/gilcrest/dbaudit).

```go
// Audit represents fields meant to track when
// an action was taken and by whom
type Audit struct {
    CreateClient    apiclient.Client
    CreateUsername  string
    CreateTimestamp time.Time
    UpdateClient    apiclient.Client
    UpdateUsername  string
    UpdateTimestamp time.Time
}
```

Within the `dbaudit.Audit` struct is the `apiclient.Client` struct from the [apiclient module](https://github.com/gilcrest/apiclient) which represents the Client of an API.

```go
type Client struct {
    Number             int
    ID                 string
    Name               string
    ServerToken        string
    HomeURL            string
    Description        string
    RedirectURI        string
    PrimaryUserID      string
    Secret             string
    CreateClientNumber int
    CreateTimestamp    time.Time
    UpdateClientNumber int
    UpdateTimestamp    time.Time
}
```

## Operations

- [Create](#Create)

### Create

To create a movie, the `Create` method should be called, sending in a context, logger and a pointer to a sql transaction. Within `Create`, the first step is to validate all input data.

```go
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
```

Input validation is done via the `validate` method. Here basic validations (missing fields, etc.) and business input validations are performed.

```go
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
```

Next, the api client is retrieved using the `ViaServerToken` method of the [apiclient module](https://github.com/gilcrest/apiclient) using the servertoken held within the context (see the [servertoken module](https://github.com/gilcrest/servertoken) to understand where the servertoken is set to the context). The client is retrieved for auditing purposes. In the database table, the create client, create username, create timestamp, update client, update username and update timestamp are all kept on the table.

```go
    // Pull client information from Server token and set
    createClient, err := apiclient.ViaServerToken(ctx, tx)
    if err != nil {
        return errors.E(op, errors.Internal, err)
    }
    m.CreateClient.Number = createClient.Number
    m.UpdateClient.Number = createClient.Number
```

Next, the `createDB` method is called passing through the context, logger and pointer to the database transaction. I am also handling rollback here and returning errors from the sent up from the `createDB` method. The errors fit the pattern discussed in the [errors section](https://github.com/gilcrest/go-api-basic#errors) of the `go-api-basic` module, but also follows the a pattern seen in the [Go standard library](https://golang.org/pkg/database/sql/#example_Tx_Rollback). There are several ways of handling rollbacks I've seen, but I think this way is the cleanest. Notice the type assertion below is used to pull out the Kind, Code and Param types as within createDB, I'm setting up a particular Code for an exists error thrown from the PostgreSQL database driver.

```go
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
```

Inside the `createDB` method, the first step is to use the `PrepareContext` method of the `sql.Tx` to prepare a sql statement which calls a stored function in the database to insert the movie data. I check for errors from the statement preparation and defer closing the statement. You can certainly choose to just write a straight insert here (it's much faster to code that way), but I like using PostgreSQL's PLpgSQL language for certain things.

```go
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
```

The stored function in the database `demo.create_movie` is below. Note that this function returns a `TABLE`. I love this feature of `PLpgSQL`.

```sql
create function create_movie(p_title character varying, p_year integer, p_rated character varying, p_released date, p_run_time integer, p_director character varying, p_writer character varying, p_create_client_num integer, p_create_username character varying) returns TABLE(o_create_client_num integer, o_create_username character varying, o_create_timestamp timestamp without time zone, o_update_client_num integer, o_update_username character varying, o_update_timestamp timestamp without time zone)
  language plpgsql
as
$$
DECLARE
  v_dml_timestamp TIMESTAMP;
  v_create_client_num integer;
  v_create_username character varying;
  v_create_timestamp timestamp;
  v_update_client_num integer;
  v_update_username character varying;
  v_update_timestamp timestamp;
BEGIN

  v_dml_timestamp := now() at time zone 'utc';

  INSERT INTO demo.movie (title,
                          year,
                          rated,
                          released,
                          run_time,
                          director,
                          writer,
                          create_client_num,
                          create_username,
                          create_timestamp,
                          update_client_num,
                          update_username,
                          update_timestamp)
  VALUES (p_title,
          p_year,
          p_rated,
          p_released,
          p_run_time,
          p_director,
          p_writer,
          p_create_client_num,
          p_create_username,
          v_dml_timestamp,
          p_create_client_num,
          p_create_username,
          v_dml_timestamp)
      RETURNING create_client_num, create_username, create_timestamp, update_client_num, update_username, update_timestamp
        into v_create_client_num, v_create_username, v_create_timestamp, v_update_client_num, v_update_username, v_update_timestamp;

      o_create_client_num := v_create_client_num;
      o_create_username := v_create_username;
      o_create_timestamp := v_create_timestamp;
      o_update_client_num := v_update_client_num;
      o_update_username := v_update_username;
      o_update_timestamp := v_update_timestamp;

      RETURN NEXT;

END;

$$;
```

To talk about a couple points in the `PLpgSQL` above, the timestamp is set into the `v_dml_timestamp` variable for the dml (insert). I like relying on the database time for capturing when an event occurs rather than distributed server time.

```sql
  v_dml_timestamp := now() at time zone 'utc';
```

I use the `RETURNING` clause into variables to capture some of the data returned from the insert (I really only need the timestamps here)

```sql
      RETURNING create_client_num, create_username, create_timestamp, update_client_num, update_username, update_timestamp
        into v_create_client_num, v_create_username, v_create_timestamp, v_update_client_num, v_update_username, v_update_timestamp;
```

Then set the outbound variables for the TABLE and return.

```sql
      o_create_client_num := v_create_client_num;
      o_create_username := v_create_username;
      o_create_timestamp := v_create_timestamp;
      o_update_client_num := v_update_client_num;
      o_update_username := v_update_username;
      o_update_timestamp := v_update_timestamp;


      RETURN NEXT;

```

Back to `Go` and inside the `createDB` method, the actual execution of the prepared statement happens using the `QueryContext` method of the Stmt struct from above. The `QueryContext` returns a pointer to a `Rows` struct. If no errors, the `Close()` method of the `Rows` struct is deferred.

```go
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
```

Two local variables (createTime and updateTime) are defined to store the create and update time returned from the function.

```go
    var (
        createTime time.Time
        updateTime time.Time
    )
```

The `Rows` struct returned above stores the results of the query execution. The `Next` method of the `Rows` struct prepares a row from the results as well as returns a true boolean if there is a row to process. The `for rows.Next()` statement checks if there is a row to process and if so, moves into the `Scan` method, which copies the result columns into the variables defined above.

```go
    // Iterate through the returned record(s)
    for rows.Next() {
        if err := rows.Scan(&createTime, &updateTime); err != nil {
            return errors.E(op, err)
        }
    }
```

If any error was encountered while iterating through rows.Next above it will be returned here.

```go
    if err := rows.Err(); err != nil {
        return errors.E(op, err)
    }
```

Finally, the timestamp fields are set to the `Movie` struct from the times returned from the database and the method returns a nil error back to `Create`

```go
    // set the dbaudit fields with timestamps from the database
    m.CreateTimestamp = createTime
    m.UpdateTimestamp = updateTime

    return nil
}
```

Back inside `Create`, the `sql.Tx` is committed and a nil error is passed back to the caller, ending the Create flow!

```go
    if err := tx.Commit(); err != nil {
        return errors.E(op, errors.Database, err)
    }

    return nil
}
```