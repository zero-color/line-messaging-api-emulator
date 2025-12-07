package db

import (
	"testing"
)

func TestMain(m *testing.M) {
	closeDB := SetupTestDB()
	defer closeDB()

	m.Run()
}
