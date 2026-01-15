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
