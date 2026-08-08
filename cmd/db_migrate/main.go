package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/logger"
)

func main() {
	var fromDSN, toDSN string

	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))

	flag.StringVar(&fromDSN, "from", "", "source DSN")
	flag.StringVar(&toDSN, "to", "", "destination DSN")

	flag.Parse()

	if fromDSN == "" || toDSN == "" {
		flag.Usage()
		logger.Logger.Error("missing DSN")
		os.Exit(1)
	}

	stopDB := func(d *gorm.DB) {
		if err := db.Stop(d); err != nil {
			logger.Logger.WithError(err).Error("failed to close database")
		}
	}

	logger.Logger.Info("opening source DB...")
	fromDB, err := db.NewDB(fromDSN, false)
	if err != nil {
		logger.Logger.WithError(err).Error("failed to open source database")
		os.Exit(1)
	}
	defer stopDB(fromDB)

	logger.Logger.Info("opening destination DB...")
	toDB, err := db.NewDB(toDSN, false)
	if err != nil {
		logger.Logger.WithError(err).Error("failed to open destination database")
		os.Exit(1)
	}
	defer stopDB(toDB)

	// When migrating to Postgres (e.g. a cloud deployment) the node data must
	// be stored in VSS, since only the database is migrated by this tool.
	if toDB.Dialector.Name() == "postgres" {
		var vssConfig db.UserConfig
		result := fromDB.Where("key = ?", "LdkVssEnabled").First(&vssConfig)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				logger.Logger.Error("LdkVssEnabled config not found in source DB. Migration will not proceed.")
			} else {
				logger.Logger.WithError(result.Error).Error("failed to query LdkVssEnabled config from source DB")
			}
			os.Exit(1)
		}

		if vssConfig.Value != "true" {
			logger.Logger.Error("VSS is not enabled in the source DB (LdkVssEnabled is not 'true'). Migration will not proceed.")
			os.Exit(1)
		}
		logger.Logger.Info("LdkVssEnabled check passed.")
	}

	logger.Logger.Info("migrating...")
	err = db.MigrateDB(fromDB, toDB)
	if err != nil {
		logger.Logger.WithError(err).Error("failed to migrate database")
		os.Exit(1)
	}

	logger.Logger.Info("migration complete")
}
