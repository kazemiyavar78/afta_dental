package migrate

import "gorm.io/gorm"

// migrate tables
func Migrate(db *gorm.DB , tables []interface{})  error {

	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			return err
		}
	}

	return nil
}