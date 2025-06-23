# Introduction

srsc uses JSON for configuration files.

### Structure

```json
{
  "log": {},
  "listen": "",
  "listen_port": 0,
  "endpoints": {},
  "tls": {},
  "cache": {},
  "resources": {}
}
```

### Fields

#### log

Log configuration, see [Log](https://sing-box.sagernet.org/configuration/log/).

#### listen

==Required==

Listen address.

#### listen_port

==Required==

Listen port.

#### endpoints

HTTP endpoint configuration, see [Endpoint](./endpoint/).

#### tls

TLS configuration, see [TLS](https://sing-box.sagernet.org/configuration/shared/tls/#inbound).

#### cache

Cache configuration, see [Cache](./cache/).

#### resources

Resource configuration, see [Resources](./resources/).

### Check

```bash
srsc check
```

### Format

```bash
srsc format -w -c config.json -D config_directory
```
