package movie

import (
	"testing"
	"time"

	"github.com/gilcrest/dbaudit"
)

func TestMovie_validate(t *testing.T) {
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
				Audit: dbaudit.Audit{CreateClientID: "123456789",
					CreateUsername:  tt.fields.CreateUsername,
					CreateTimestamp: time.Now(),
					UpdateClientID:  "123456789",
					UpdateUsername:  tt.fields.CreateUsername,
					UpdateTimestamp: time.Now(),
				},
			}
			if err := m.validate(); (err != nil) != tt.wantErr {
				t.Errorf("Movie.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
