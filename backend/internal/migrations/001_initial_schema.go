package migrations

import (
"github.com/guigui42/go-react-starter/internal/models"
"gorm.io/gorm"
)

func init() {
Register(Migration{
Version: "001",
Name:    "initial_schema",
Up: func(db *gorm.DB) error {
return db.AutoMigrate(models.AllModels()...)
},
Down: func(db *gorm.DB) error {
// Drop all tables in reverse order
allModels := models.AllModels()
for i := len(allModels) - 1; i >= 0; i-- {
if err := db.Migrator().DropTable(allModels[i]); err != nil {
return err
}
}
return nil
},
})
}
