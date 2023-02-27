package database

import (
	"database/sql"
	"github.com/go-faster/errors"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

// NewDatabase will open a new database
func NewDatabase(filename string) (Database, error) {
	// Create the database and set it up
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return Database{}, errors.Wrap(err, "cannot open database")
	}
	db.SetMaxOpenConns(1)
	err = db.Ping()
	if err != nil {
		return Database{}, errors.Wrap(err, "cannot ping database")
	}
	// Create initial tables
	query := `CREATE TABLE IF NOT EXISTS users (tg_id INTEGER NOT NULL PRIMARY KEY, email TEXT) WITHOUT ROWID;
CREATE TABLE IF NOT EXISTS lists (id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, inline_id TEXT NOT NULL UNIQUE, title TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS lists_users_pivot (user_id INTEGER NOT NULL, list_id INTEGER NOT NULL, PRIMARY KEY (user_id, list_id), FOREIGN KEY(user_id) REFERENCES users(tg_id), FOREIGN KEY(list_id) REFERENCES lists(id)) WITHOUT ROWID;`
	_, err = db.Exec(query)
	return Database{db}, nil
}

// UpdateEmail will update the email of a user in database.
func (db Database) UpdateEmail(userID int64, email string) error {
	_, err := db.db.Exec("REPLACE INTO users (tg_id, email) VALUES (?, ?)", userID, email)
	return err
}

// GetEmail will get the email of user. If the user has no email, it will return an
// empty string.
func (db Database) GetEmail(userID int64) (string, error) {
	var email string
	err := db.db.QueryRow("SELECT email FROM users WHERE tg_id=?", userID).Scan(&email)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	return email, nil
}

// CreateList will create a list with given name and return its id.
func (db Database) CreateList(id, name string) error {
	_, err := db.db.Exec("INSERT INTO lists (inline_id, title) VALUES (?, ?)", id, name)
	return err
}

// GetListEmails gets the emails signed up in a list seperated by a new line.
// Also gets the list name.
func (db Database) GetListEmails(id string) (string, []string, error) {
	// Get the list name
	var listName string
	err := db.db.QueryRow("SELECT title FROM lists WHERE inline_id = ?", id).Scan(&listName)
	if err != nil {
		return "", nil, errors.Wrap(err, "cannot read title of list")
	}
	// Get the emails
	rows, err := db.db.Query("SELECT u.email FROM lists_users_pivot p INNER JOIN users u on u.tg_id = p.user_id INNER JOIN lists l on l.id = p.list_id WHERE l.inline_id = ? AND u.email IS NOT NULL", id)
	if err != nil {
		return "", nil, errors.Wrap(err, "cannot read emails")
	}
	defer rows.Close()
	// Read all emails
	var emails []string
	for rows.Next() {
		var email string
		err = rows.Scan(&email)
		if err != nil {
			return "", nil, errors.Wrap(err, "cannot scan row")
		}
		emails = append(emails, email)
	}
	// Done
	return listName, emails, nil
}

// AddToList will add a user to a list
func (db Database) AddToList(userID int64, listID string) error {
	_, err := db.db.Exec("INSERT INTO lists_users_pivot (user_id, list_id) VALUES (?, (SELECT id FROM lists WHERE inline_id = ?))", userID, listID)
	return err
}
