package infrastructure

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"rol/app/interfaces"
	"rol/domain"
)

//TFTPConfigRepository repository for TFTPServer entity
type TFTPConfigRepository struct {
	*GormGenericRepository[domain.TFTPConfig]
}

//NewTFTPConfigRepository constructor for domain.TFTPServer GORM generic repository
//Params
//	db - gorm database
//	log - logrus logger
//Return
//	generic.IGenericRepository[domain.TFTPServer] - new tftp server repository
func NewTFTPConfigRepository(db *gorm.DB, log *logrus.Logger) interfaces.IGenericRepository[domain.TFTPConfig] {
	genericRepository := NewGormGenericRepository[domain.TFTPConfig](db, log)
	return TFTPConfigRepository{
		genericRepository,
	}
}
