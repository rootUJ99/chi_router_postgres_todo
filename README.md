# chi_router_postgres_todo
chi router with postgres todo

add .env file with the following variables:

```cmd
DATABASE_URL=postgres://user:password@host:port/dbname
```
make migrations using goose

```cmd
cd migrations
```

```cmd
goose postgres "postgres://user:password@host:port/dbname" up 
```