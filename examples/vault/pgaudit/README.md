# vault-log-audit PostgreSQL and PGAudit

This docker-compose example shows how to use vault-log-audit to collect pgaudit PostgreSQL logs.

## Running

```bash
docker-compose up
```

## Overview 

The docker-compose contains of:

Containers:

- PostgreSQL, configured to store logs in json format and has enabled pgaudit extension. For purpose of this example, logs are rotated every minute. This can be changed to week, month etc, depending on a need and log volume. 
- vault-log-audit

Init containers:

- volume_init, to set proper permissions for log volume
- vault-log-audit-init, to create vault if it did not exist

Volumes:

- postgresql_data, holds PostgreSQL database
- postgresql_logs, holds PostgresQL logs
- vault-log-audit_data, holds montored log file tracking registry file