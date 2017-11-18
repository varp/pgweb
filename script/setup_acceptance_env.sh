#!/usr/bin/env bash

set -e

export PGWEB_VERSION=0.9.10
export PGWEB_PORT=8081
export PGHOST=${PGHOST:-localhost}
export PGUSER="postgres"
export PGVERSION="9.6"
export PGPASSWORD="postgres"
export PGDATABASE="booktown"
export PGPORT="15432"



if [[ $# -gt 0 &&  $1 == "-c" ]]; then
    echo "--------------- CLEANING ACCEPTANCE ENV ---------------"
    docker rm -f postgres || true
    docker rm -f pgweb_${PGWEB_VERSION} || true
    killall pgweb || true
    echo "------------ CLEANING ACCEPTANCE ENV DONE -------------"
    exit 0
fi


echo "---------------- RUN PG 9.6 ----------------"
docker rm -f postgres || true
docker run -p $PGPORT:5432 --name postgres -e POSTGRES_PASSWORD=$PGPASSWORD -d postgres:$PGVERSION
sleep 5
docker cp ./data/booktown.sql postgres:/booktown.sql
docker exec postgres psql -U postgres -f /booktown.sql
echo "-------------- PG 9.6 RUN END --------------"


echo "-------------- RUN PGWEB GUI-------------------"
if [[ -f ./pgweb ]]; then
  echo "Starting dev build of pgweb"
  killall pgweb || true
  ./pgweb -s 2>&1 > /dev/null &
else
  docker rm -f pgweb_${PGWEB_VERSION} || true
  docker build -t pgweb:${PGWEB_VERSION} .
  docker run --name pgweb_${PGWEB_VERSION} --link=postgres -p ${PGWEB_PORT}:${PGWEB_PORT} -d pgweb:${PGWEB_VERSION}
fi
echo "------------ PGWEB GUI IS READY ---------------"
