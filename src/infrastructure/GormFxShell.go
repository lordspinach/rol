package infrastructure

import "gorm.io/gorm"

//GormFxShell Shell for second database passing in FX DI framework
type GormFxShell struct {
	dbShell *gorm.DB
}

//GetDb gets database from shell
//Return
//	*gorm.DB - gorm database
func (g GormFxShell) GetDb() *gorm.DB {
	return g.dbShell
}
