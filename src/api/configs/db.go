package configs

import (
	"os"
)

const (
	dbStageUser     = ""
	dbStagePassword = ""
	dbStageHost     = ""
	dbStageName     = ""
)

const (
	dbLocalUser     = "root"
	dbLocalPassword = "123456"
	dbLocalHost     = "localhost:3306"
	dbLocalName     = "rpconfigs"
)

func GetDBConnectionParams() []interface{} {
	switch scope := os.Getenv("SCOPE"); scope {
	case "production", "test":
		//dbProductionName := gomelipass.GetEnv("SECRET_CONFIGS_DB_NAME")
		//dbProductionUser := gomelipass.GetEnv("SECRET_CONFIGS_DB_USER")
		//dbProductionPassword := gomelipass.GetEnv("DB_MYSQL_RPCONFIGSAPI00_RPCONFIGS_RPCONFIGS_WPROD")
		//dbProductionHost := gomelipass.GetEnv("DB_MYSQL_RPCONFIGSAPI00_RPCONFIGS_RPCONFIGS_ENDPOINT")
		//return []interface{}{dbProductionUser, dbProductionPassword, dbProductionHost, dbProductionName}
	case "stage":
		return []interface{}{dbStageUser, dbStagePassword, dbStageHost, dbStageName}
	default:
		return []interface{}{dbLocalUser, dbLocalPassword, dbLocalHost, dbLocalName}
	}
}
