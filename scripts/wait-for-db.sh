#!/bin/sh

echo "⏳ Waiting for PostgreSQL..."

until pg_isready -h timescaledb -p 5432 -U gpsgo
do
  sleep 2
done

echo "✅ PostgreSQL is ready!"

# Extra safety (important for Timescale restart)
sleep 5

echo "🚀 Running migrations..."

exec migrate -path=/migrations \
  -database=postgres://gpsgo:gpsgo@timescaledb:5432/gpsgo?sslmode=require \
  up