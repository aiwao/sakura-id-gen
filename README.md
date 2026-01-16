# Sakura ID Generator
さくらインターネットID（アカウント）ジェネレーター

## Start the ID generator
```
docker compose --env-file .env up
```
Only start the database
```
docker compose up db
```

## Connect to the database (example)
```
psql --host localhost --port 5432 --username user --password password --dbname accountdb
```

## Get accounts data as json array
```
psql --host localhost --port 5432 --username user --password password --dbname accountdb -t -A -c "SELECT json_agg(row_to_json(t)) FROM (SELECT * FROM accounts) t;" > accounts.json
```
