package models

import (
	"errors"

	"github.com/shkuran/go-library/db"
	"github.com/shkuran/go-library/utils"
)

type User struct {
	Id       int64  `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Email    string `json:"email" db:"email" binding:"required"`
	Password string `json:"password" db:"password" binding:"required"`
}

func SaveUser(user User) error {
	query := `
	INSERT INTO users (name, email, password) 
	VALUES (?, ?, ?)
	`

	hashedPass, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}

	_, err = db.DB.Exec(query, user.Name, user.Email, hashedPass)
	if err != nil {
		return err
	}

	return nil
}

func GetUsers() ([]User, error) {
	rows, err := db.DB.Query("SELECT * FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.Password)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func GetUserById(id int64) (User, error) {
	var user User

	row := db.DB.QueryRow("SELECT * FROM users WHERE id = ?", id)
	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.Password)
	if err != nil {
		return user, err
	}

	return user, nil
}

func ValidateCredentials(u *User) error {
	query := "SELECT id, password FROM users WHERE email = ?"
	row := db.DB.QueryRow(query, u.Email)

	var passFromDB string
	err := row.Scan(&u.Id, &passFromDB)
	if err != nil {
		return errors.New("invalid credentials")
	}

	passwordIsValid := utils.CheckPasswordHash(u.Password, passFromDB)

	if !passwordIsValid {
		return errors.New("invalid credentials")
	}
	return nil
}
