package db

import "fmt"

func GetUserHash(username string) ([]byte, error) {
	var storedHash []byte

	err := DB.QueryRow("select password from users where username = ?", username).Scan(&storedHash)
	if err != nil {
		return nil, fmt.Errorf("username not found %w", err)
	}

	return storedHash, nil
}

func CheckUserBlocked(username string) (bool, error) {
	var role string

	err := DB.QueryRow("select role from users where username = ?", username).Scan(&role)
	if err != nil {
		return false, fmt.Errorf("user select error %w", err)
	}

	return role == "blocked", nil
}
