package dtos

import "github.com/google/uuid"

//TFTPPathDto TFTP path dto
type TFTPPathDto struct {
	//ID entity identifier
	ID uuid.UUID
	//TFTPServerID TFTP config entity ID
	TFTPServerID uuid.UUID
	TFTPPathBaseDto
}
