package database

import (
	"jira-clone-api/database/mongo"
)

func InitDatabase() {
	mongo.InitDatabase()
}

func DisconnectDatabase() {
	mongo.DisconnectDatabase()
}
