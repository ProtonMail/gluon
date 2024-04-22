package sqlite3

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestClient_DBConnectionSpecialCharacterPath(t *testing.T) {
	dbDirs := []string{
		"#test",
		"test_test",
		"test#test#test",
		"test$test$test",
	}

	testingDir := t.TempDir()

	for _, dirName := range dbDirs {
		path := filepath.Join(testingDir, dirName)
		if err := os.MkdirAll(path, 0777); err != nil {
			fmt.Println("Could not create testing directory, error: ", err)
			t.FailNow()
		}

		filePath := filepath.Join(path, "test.db")

		client, err := sql.Open("sqlite3", getDatabaseConn("test", "test", filePath))
		if err != nil {
			fmt.Println("Could not connect to test database, error: ", err)
			t.FailNow()
		}

		if err := client.Ping(); err != nil {
			fmt.Println("Could not ping test database, error: ", err)
			client.Close()
			t.FailNow()
		}

		if err := client.Close(); err != nil {
			fmt.Println("Could not close test database, error: ", err)
			t.FailNow()
		}
	}
}
