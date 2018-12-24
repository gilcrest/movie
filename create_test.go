// Package movie has the User type and associated methods to
// create, modify and delete application users
package movie

import (
	"testing"
	"time"
)

func TestMovie_validate2(t *testing.T) {
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
				t.Errorf("Movie.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
