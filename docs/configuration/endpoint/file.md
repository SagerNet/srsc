# File

The File endpoint converts rule-sets from local or remote files.

### Structure

```json
{
  "type": "file",
  "source": "",
  
  ..., // Source Fields
  ... // Convertor Fields
}
```

### Source Structure

=== "Local"

    ```json
    {
      "source": "local",
      "path": ""
    }
    ```

=== "Remote"

    ```json
    {
      "source": "remote",
      "url": "",
      "user_agent": "",
      "ttl": "",
      "tls": {},
      
      ... // Dial Fields
    }
    ```

### Fields

#### source

==Required==

Source of rule-sets, `local` or `remote`.

### Local Fields

#### path

==Required==

Path to the file to be converted.

Templates in the endpoint path can be used in the path, for example:

```json
{
  ...,
  
  "endpoints": {
    ...,
    
    "/{name}.srs": {
      "type": "file",
      "source": "local",
      "path": "/path/to/{{ name }}.json",
      
      ...
    }
  }
}
```

### Remote Fields

#### url

==Required==

URL of the remote file to be converted.

Templates in the endpoint path can be used in the URL, for example:

```json
{
  ...,
  
  "endpoints": {
    ...,
    
    "/{name}.srs": {
      "type": "file",
      "source": "remote",
      "url": "https://example.com/{{ name }}.json",
      
      ...
    }
  }
}
```

#### user_agent

Custom User-Agent in HTTP requests.

`srsc/$version (sing-box $sing-box-version)` is used by default.

#### ttl

Minimum time interval to check for updates.

`5m` is used by default.

#### tls

Custom TLS configuration, see [TLS](https://sing-box.sagernet.org/configuration/shared/tls/#outbound).

### Dial Fields

Custom dialer options, see [Dial Fields](https://sing-box.sagernet.org/configuration/shared/dial/).

Only basic options are supported, features like `detour`, DNS, multi-network dialing, etc.
that rely on sing-box are not available.

### Convertor Fields

#### source_type

==Required==

See [Convertors](/configuration/convertor/) for more details.
