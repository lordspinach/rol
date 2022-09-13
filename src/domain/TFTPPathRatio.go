package domain

import "github.com/google/uuid"

//TFTPPathRatio TFTP path ratio entity
type TFTPPathRatio struct {
	Entity
	//TFTPServerID TFTP config entity ID
	TFTPServerID uuid.UUID
	//ActualPath actual file path
	ActualPath string
	//VirtualPath virtual file path
	VirtualPath string
}
