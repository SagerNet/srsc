# Convertor

| Convertor | Format                |
|-----------|-----------------------|
| `source`  | [Source](./source/)   |
| `binary`  | [Binary](./target/)   |
| `adguard` | [AdGuard](./adguard/) |
| `clash`   | [Clash](./clash/)     |
| `surge`  | [Surge](./surge/)     |

### Source Structure

```json
{
  "source_type": "",
  
  ... // Type Specific Fields
}
```

### Target Structure

```json
{
  "target_type": "",
  
  ... // Type Specific Fields
}
```

### Source Fields

#### source_type

==Required==

The type of the source convertor.

### Target Fields

#### target_type

==Required==

The type of the target convertor.
