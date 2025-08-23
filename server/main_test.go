package server_test

import (
	"testing"

	"github.com/zero-color/line-messaging-api-emulator/db"
)

func TestMain(m *testing.M) {
	closeDB := db.SetupTestDB()
	defer closeDB()

	m.Run()
}
