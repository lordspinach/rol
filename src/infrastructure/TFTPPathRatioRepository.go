package infrastructure

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"rol/app/interfaces"
	"rol/domain"
)

//TFTPPathRatioRepository repository for TFTPPathRatio entity
type TFTPPathRatioRepository struct {
	*GormGenericRepository[domain.TFTPPathRatio]
}

//NewTFTPPathRatioRepository constructor for domain.TFTPPathRatio GORM generic repository
//Params
//	db - gorm database
//	log - logrus logger
//Return
//	generic.IGenericRepository[domain.TFTPPathRatio] - new tftp server repository
func NewTFTPPathRatioRepository(db *gorm.DB, log *logrus.Logger) interfaces.IGenericRepository[domain.TFTPPathRatio] {
	genericRepository := NewGormGenericRepository[domain.TFTPPathRatio](db, log)
	return TFTPPathRatioRepository{
		genericRepository,
	}
}
