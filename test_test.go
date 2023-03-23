package main

import (
	"database/sql"
	"testing"
)

func TestCreateTable(t *testing.T) {
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = createTable(db)
	if err != nil {
		t.Errorf("createTable() returned an error: %v", err)
	}

	// Check that the table was created
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='mibs'")
	err = row.Scan(&count)
	if err != nil {
		t.Errorf("Failed to query table count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected table count to be 1, but got %d", count)
	}
}

func TestGetLastRowFetched(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS mibs (LastRowFetched INTEGER)"); err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	tests := []struct {
		name             string
		initialLastRow   int
		expectedLastRow  int
		expectedErrIsNil bool
	}{
		{
			name:             "no last row",
			initialLastRow:   0,
			expectedLastRow:  0,
			expectedErrIsNil: true,
		},
		{
			name:             "last row exists",
			initialLastRow:   10,
			expectedLastRow:  10,
			expectedErrIsNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := db.Exec("DELETE FROM mibs"); err != nil {
				t.Fatalf("failed to clear table: %v", err)
			}

			if _, err := db.Exec("INSERT INTO mibs (LastRowFetched) VALUES (?)", tc.initialLastRow); err != nil {
				t.Fatalf("failed to insert row: %v", err)
			}

			lastRow, err := getLastRowFetched(db)
			if (err == nil) != tc.expectedErrIsNil {
				t.Fatalf("expected err == nil to be %v, but got %v", tc.expectedErrIsNil, err)
			}
			if lastRow != tc.expectedLastRow {
				t.Errorf("expected last row to be %d, but got %d", tc.expectedLastRow, lastRow)
			}
		})
	}
}