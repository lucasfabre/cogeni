CREATE TABLE users (
  id INTEGER PRIMARY KEY,
  username TEXT,
  email TEXT
);

CREATE TABLE products (
  id INTEGER PRIMARY KEY,
  name TEXT,
  price REAL
);

CREATE TABLE todos (
  id INTEGER PRIMARY KEY,
  title TEXT,
  completed BOOLEAN
);