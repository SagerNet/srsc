# Cache

### Structure

=== "Memory"

    ```json
    {
      "type": "",
      "expiration": ""
    }
    ```

=== "Redis"

    ```json
    {
      "type": "redis",
      "address": [],
      "username": "",
      "password": "",
      "db": 0,
      "protocol": 0,
      "pool_size": 0,
      "tls": {},
      "expiration": ""
    }
    ```

### Fields

#### type

| Type               | Description       |
|--------------------|-------------------|
| `memory` (default) | Use memory cache. |
| `redis`            | Use Redis cache.  |

#### expiration

Cache expiration time, in Go duration format (e.g., `5m`, `1h`, `24h`).

Never expire if not set.

### Redis Fields

#### address

Either a single address or a seed list of host:port addresses of cluster/sentinel nodes.

#### username

Username is used to authenticate the current connection with one of the connections defined in the ACL list when
connecting to a Redis 6.0 instance, or greater, that is using the Redis ACL system.

#### password

Password is an optional password. Must match the password specified in the `requirepass` server configuration option
(if connecting to a Redis 5.0 instance, or lower), or the User Password when connecting to a Redis 6.0 instance,
or greater, that is using the Redis ACL system.

#### db

DB is the database to be selected after connecting to the server.

#### protocol

Protocol `2` or `3`. Use the version to negotiate RESP version with redis-server.

`3` will be used by default.

#### pool_size

PoolSize is the base number of socket connections. 

`10` connections per every available CPU as reported by `runtime.GOMAXPROCS` will be used by default.

#### tls

TLS configuration, see [TLS](https://sing-box.sagernet.org/configuration/shared/tls/#outbound).
