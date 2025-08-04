# go-gin-api-example

# 1 Cria o volume
docker volume create notepad_on

# 2 Sobe o container
docker run -d \
  --name notepad_on_db \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=user \
  -e POSTGRES_DB=notepad_on \
  -v notepad_on:/var/lib/postgresql/data \
  -p 5432:5432 \
  postgres:15


psql -h localhost -U user -d notepad_on