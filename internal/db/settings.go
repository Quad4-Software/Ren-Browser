package db

func (d *DB) GetSetting(key string) (string, error) {
	var value string
	err := d.sql.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (d *DB) SetSetting(key, value string) error {
	_, err := d.sql.Exec(
		`INSERT INTO settings (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	return err
}
