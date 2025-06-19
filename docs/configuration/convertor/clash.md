# Clash

Clash rule provider.

### Source Structure

```json
{
  "source_type": "clash",
  "source_format": "",
  "source_behavior": ""
}
```

### Target Structure

```json
{
  "target_type": "clash",
  "target_format": "",
  "target_behavior": ""
}
```

### Source Fields

#### source_format

==Required==

The format of the input provider, available values are: `text`, `yaml`, `mrs`.

#### source_behavior

==Required==

The format of the input provider, available values are: `domain`, `ipcidr`, `classical`.

### Target Fields

#### target_format

==Required==

The format of the output provider, available values are: `text`, `yaml`, `mrs`.

#### target_behavior

==Required==

The behavior of the output provider, available values are: `domain`, `ipcidr`, `classical`.
