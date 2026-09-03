Project Setup:

Installation of required Software:
- install postgres: sudo apt install postgresql postgresql-contrib
- start Postgres database service: sudo service postgresql start
- install goose for migrations: go install github.com/pressly/goose/v3/cmd/goose@latest
- install sqlc for sql handling in Go: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

Create database and configure it:
- set Postgres password: sudo passwd your_username      (I usually use postgres as username)
- Open psql to access postgres: sudo -u postgres psql
    - In postgres create the database: CREATE DATABASE your_database_name;
    - Connect to your Database: \c your_database_name
    - Set your password for the database: ALTER USER your_username WITH PASSWORD 'your_password';
- You can exit by typing: \q

Connection String:
- postgres://your_username:your_password@localhost:5432/your_database_name

Use goose:
- To run an Up migration, navigate into the folder of the navigation and type: goose postgres your_Connection_String up

Configure sqlc and use it:
- In this project sqlc is already setup und you can create new GoFunctions from sql Queries: sqlc generate
- If anything fails check the sqlc.yaml and check the paths in it.
