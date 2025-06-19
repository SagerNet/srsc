---
icon: material/docker
---

# Docker

## :material-console: Command

```bash
docker run -d \
  -v /etc/srsc:/etc/srsc/ \
  --name=srsc \
  --restart=always \
  ghcr.io/sagernet/srsc \
  -D /var/lib/srsc \
  -C /etc/srsc/ run
```

## :material-box-shadow: Compose

```yaml
services:
  srsc:
    image: ghcr.io/sagernet/srsc
    container_name: srsc
    restart: always
    volumes:
      - /etc/srsc:/etc/srsc/
    command: -D /var/lib/srsc -C /etc/srsc/ run
```
