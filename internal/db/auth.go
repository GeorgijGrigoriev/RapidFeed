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
