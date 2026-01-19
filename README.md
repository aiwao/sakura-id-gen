# Sakura ID Generator
さくらインターネットID（アカウント）ジェネレーター

## 実行
独立したデータベースで実行
```
docker compose --profile local-db --env-file .env up
```

共有ネットワーク(``docker-compose.yml:networks``参照)内のデータベースで実行
```
docker compose --env-file .env up
```

データベースのみ実行(データ取り出し用)
```
docker compose --profile local-db up db
```

## データベースに接続 (デフォルト環境の場合) (psqlが必要)
```shell
psql --host localhost --port 5432 --username user --password password --dbname accountdb
```

## データをjsonとして書き出す
```shell
psql --host localhost --port 5432 --username user --password password --dbname accountdb -t -A -c "SELECT json_agg(row_to_json(t)) FROM (SELECT * FROM accounts) t;" > accounts.json
```
