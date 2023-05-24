package db_impl

import (
	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/internal/db_impl/ent_db"
)

func NewEntDB() db.ClientInterface {
	return ent_db.NewEntDB()
}
