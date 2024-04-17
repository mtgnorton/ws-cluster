docker run -d --name redis-container \
  -p 6389:6379 \
  -v ./redis.conf:/usr/local/etc/redis/redis.conf \
  -v ./data:/data \
  redis:6.0 redis-server /usr/local/etc/redis/redis.conf

