### To start

First
```bash
docker build -t batched-orders .
```
then
```bash
docker run -p 8000:8000 --env-file .env batched-orders
```