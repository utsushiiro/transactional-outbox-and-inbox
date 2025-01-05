FROM postgres:16

RUN apt-get update && apt-get -y --no-install-recommends install postgresql-16-cron

EXPOSE 5432
CMD ["postgres", "-c", "config_file=/etc/postgresql/postgresql.conf"]
