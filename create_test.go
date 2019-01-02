package movie

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gilcrest/dbaudit"
	"github.com/gilcrest/srvr"
	"github.com/gilcrest/srvr/datastore"
	"github.com/rs/zerolog"
)

func TestMovie_Validate(t *testing.T) {
	type fields struct {
		Title          string
		Year           int
		Rated          string
		Released       time.Time
		RunTime        int
		Director       string
		Writer         string
		CreateUsername string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"All Valid", fields{"Repo Man", 1984, "R", time.Now(), 92, "Alex Cox", "Alex Cox", "gilcrest"}, false},
		{"Invalid Year", fields{"Repo Man", 1800, "R", time.Now(), 92, "Alex Cox", "Alex Cox", "gilcrest"}, true},
		{"Missing Title", fields{"", 1800, "R", time.Now(), 92, "Alex Cox", "Alex Cox", "gilcrest"}, true},
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
				Audit: dbaudit.Audit{
					CreateClientID:  "123456789",
					CreateUsername:  tt.fields.CreateUsername,
					CreateTimestamp: time.Now(),
					UpdateClientID:  "123456789",
					UpdateUsername:  tt.fields.CreateUsername,
					UpdateTimestamp: time.Now(),
				},
			}
			if err := m.Validate(); (err != nil) != tt.wantErr {
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
		Audit    dbaudit.Audit
	}
	type args struct {
		ctx       context.Context
		log       zerolog.Logger
		tx        *sql.Tx
		srvrToken string
	}

	srvr, err := srvr.NewServer(zerolog.DebugLevel)
	if err != nil {
		t.Errorf("Error from Newserver = %v", err)
	}
	ctx := context.Background()

	aud := dbaudit.Audit{CreateClientID: "FakeClientID",
		CreateUsername:  "gilcrest",
		CreateTimestamp: time.Now(),
		UpdateClientID:  "FakeClientID",
		UpdateUsername:  "gilcrest",
		UpdateTimestamp: time.Now()}

	f1 := fields{"Repo Man", 1984, "R", time.Now(), 92, "Alex Cox", "Alex Cox", aud}

	tx, err := srvr.DS.BeginTx(ctx, nil, datastore.AppDB)
	if err != nil {
		t.Errorf("Error from BeginTx = %v", err)
	}
	arg := args{ctx, srvr.Logger, tx, os.Getenv("TEST_SERVER_TOKEN")}

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
				Audit:    tt.fields.Audit,
			}
			tx, err := m.createDB(tt.args.ctx, tt.args.log, tt.args.tx, tt.args.srvrToken)
			if err != nil {
				t.Errorf("Movie.createDB() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !m.CreateTimestamp.IsZero() {
				fmt.Printf("Timestampe = %v", m.CreateTimestamp)
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
