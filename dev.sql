PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
  CREATE TABLE books(
    pk integer primary key autoincrement,
    title text,
    author text,
    id text,
    classification text
  );
COMMIT;
