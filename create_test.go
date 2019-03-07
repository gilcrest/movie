package movie

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gilcrest/servertoken"
	"github.com/gilcrest/srvr"
	"github.com/gilcrest/srvr/datastore"
	"github.com/rs/zerolog"
)

func TestMovie_Validate(t *testing.T) {
	type fields struct {
		Title    string
		Year     int
		Rated    string
		Released time.Time
		RunTime  int
		Director string
		Writer   string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"All Valid", fields{"Repo Man", 1984, "R", time.Now(), 92, "Alex Cox", "Alex Cox"}, false},
		{"Invalid Year", fields{"Repo Man", 1800, "R", time.Now(), 92, "Alex Cox", "Alex Cox"}, true},
		{"Missing Title", fields{"", 1800, "R", time.Now(), 92, "Alex Cox", "Alex Cox"}, true},
		{"Missing Rated", fields{"Repo Man", 1984, "", time.Now(), 92, "Alex Cox", "Alex Cox"}, true},
		{"Zero ReleaseDate", fields{"Repo Man", 1984, "R", time.Time{}, 92, "Alex Cox", "Alex Cox"}, true},
		{"Invalid Runtime", fields{"Repo Man", 1984, "R", time.Now(), 0, "Alex Cox", "Alex Cox"}, true},
		{"Missing Director", fields{"Repo Man", 1984, "R", time.Now(), 92, "", "Alex Cox"}, true},
		{"Missing Writer", fields{"Repo Man", 1984, "R", time.Now(), 92, "Alex Cox", ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Movie{
				Title:    tt.fields.Title,
				Year:     tt.fields.Year,
				Rated:    tt.fields.Rated,
				Released: tt.fields.Released,
				RunTime:  tt.fields.RunTime,
				Director: tt.fields.Director,
				Writer:   tt.fields.Writer,
			}
			if err := m.validate(); (err != nil) != tt.wantErr {
				t.Errorf("Movie.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMovie_CreateDB(t *testing.T) {
	type fields struct {
		Title    string
		Year     int
		Rated    string
		Released time.Time
		RunTime  int
		Director string
		Writer   string
	}
	type args struct {
		ctx context.Context
		log zerolog.Logger
		tx  *sql.Tx
	}

	srvr, err := srvr.NewServer(zerolog.DebugLevel)
	if err != nil {
		t.Errorf("Error from Newserver = %v", err)
	}
	token := servertoken.ServerToken(os.Getenv("TEST_SERVER_TOKEN"))
	ctx := context.Background()
	ctx = token.Add2Ctx(ctx)

	f1 := fields{"Repo Man", 1984, "R", time.Now(), 92, "Alex Cox", "Alex Cox"} //, aud}

	tx, err := srvr.DS.BeginTx(ctx, nil, datastore.AppDB)
	if err != nil {
		t.Errorf("Error from BeginTx = %v", err)
	}
	arg := args{ctx, srvr.Logger, tx}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"Valid Test", f1, arg, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Movie{
				Title:    tt.fields.Title,
				Year:     tt.fields.Year,
				Rated:    tt.fields.Rated,
				Released: tt.fields.Released,
				RunTime:  tt.fields.RunTime,
				Director: tt.fields.Director,
				Writer:   tt.fields.Writer,
			}
			err := m.createDB(tt.args.ctx, tt.args.log, tt.args.tx)
			if err != nil {
				t.Errorf("Movie.createDB() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !m.CreateTimestamp.IsZero() {
				fmt.Printf("Timestamp = %v", m.CreateTimestamp)
				err := tx.Commit()
				if err != nil {
					t.Errorf("Movie.createDB() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				err = tx.Rollback()
				if err != nil {
					t.Errorf("Movie.createDB() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

		})
	}
}
