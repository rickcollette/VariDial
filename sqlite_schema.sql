CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	active INT NOT NULL,
	username TEXT NOT NULL,
	number INT NOT NULL,
	password TEXT NOT NULL,
	level INT NOT NULL,
	channel INT NOT NULL
);