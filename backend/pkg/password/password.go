package password

import (
	"golang.org/x/crypto/bcrypt"
)

func Hash(password string) (string, error){
	bytePassword := []byte(password)
	hashPass, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashPass), nil
}

func Compare(password string, hash string) error{
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
}
