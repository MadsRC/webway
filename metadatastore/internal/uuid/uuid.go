package uuid

import "github.com/hashicorp/go-uuid"

type UUIDService struct{}

func NewUUIDService() (*UUIDService, error) {
	return &UUIDService{}, nil
}

func (u UUIDService) RandomUUID() (string, error) {
	return uuid.GenerateUUID()
}

func (u UUIDService) ParseUUID(in string) ([]byte, error) {
	return uuid.ParseUUID(in)
}

func (u UUIDService) FormatUUID(in []byte) (string, error) {
	return uuid.FormatUUID(in)
}
