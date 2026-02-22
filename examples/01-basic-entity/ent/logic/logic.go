package logic

import "github.com/google/uuid"

func GetUuidStr() string {
	return uuid.New().String()
}
