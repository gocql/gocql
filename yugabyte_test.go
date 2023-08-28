//go:build yugabyte
// +build yugabyte

package gocql

import (
	"bytes"
	"testing"
)

func TestJSONB(t *testing.T) {
	session := createSession(t)
	defer session.Close()

	defer func() {
		err := createTable(session, "DROP TABLE IF EXISTS gocql_test.jsonb")
		if err != nil {
			t.Logf("failed to delete jsonb table: %v", err)
		}
	}()

	if err := createTable(session, "CREATE TABLE gocql_test.jsonb (id int, my_jsonb jsonb, PRIMARY KEY (id))"); err != nil {
		t.Fatalf("failed to create table with error '%v'", err)
	}

	b := session.NewBatch(LoggedBatch)
	b.Query("INSERT INTO gocql_test.jsonb(id, my_jsonb) VALUES (?,?)", 1, []byte("true"))
	b.Query("INSERT INTO gocql_test.jsonb(id, my_jsonb) VALUES (?,?)", 2, []byte(`{"foo":"bar"}`))

	if err := session.ExecuteBatch(b); err != nil {
		t.Fatalf("query failed. %v", err)
	} else {
		if b.Attempts() < 1 {
			t.Fatal("expected at least 1 attempt, but got 0")
		}
		if b.Latency() <= 0 {
			t.Fatalf("expected latency to be greater than 0, but got %v instead.", b.Latency())
		}
	}

	var id int
	var myJSONB []byte
	if err := session.Query("SELECT id, my_jsonb FROM gocql_test.jsonb WHERE id = 1;").Scan(&id, &myJSONB); err != nil {
		t.Fatalf("Failed to select with err: %v", err)
	} else if id != 1 {
		t.Fatalf("Expected id = 1, got %v", id)
	} else if !bytes.Equal(myJSONB, []byte("true")) {
		t.Fatalf("Expected my_jsonb = true, got %v", string(myJSONB))
	}

	if err := session.Query("SELECT id, my_jsonb FROM gocql_test.jsonb WHERE id = 2;").Scan(&id, &myJSONB); err != nil {
		t.Fatalf("Failed to select with err: %v", err)
	} else if id != 2 {
		t.Fatalf("Expected id = 2, got %v", id)
	} else if !bytes.Equal(myJSONB, []byte(`{"foo":"bar"}`)) {
		t.Fatalf(`Expected my_jsonb = {"foo":"bar"}, got %v`, string(myJSONB))
	}

	if rd, err := session.Query("SELECT id, my_jsonb FROM gocql_test.jsonb;").Iter().RowData(); err != nil {
		t.Fatalf("Failed to select with err: %v", err)
	} else if len(rd.Columns) != 2 || rd.Columns[0] != "id" || rd.Columns[1] != "my_jsonb" {
		t.Fatalf("Expected [id, my_jsonb], got %v", rd.Columns)
	} else if len(rd.Values) != 2 {
		t.Fatalf("Expected 2 values, got %v", rd.Values)
	} else if _, ok := rd.Values[0].(*int); !ok {
		t.Fatalf("Expected values[0] = *int, got %T", rd.Values[0])
	} else if _, ok := rd.Values[1].(*[]byte); !ok {
		t.Fatalf("Expected values[1] = *[]byte, got %T", rd.Values[1])
	}
}
