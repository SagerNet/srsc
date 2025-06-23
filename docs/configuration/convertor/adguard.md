# AdGuard

AdGuard DNS Filer.

Special rule-set supported by sing-box, covering most of the syntax supported by AdGuard Home.

Since it contains many other rules that cannot be represented in any other format,
it can only be converted to and from `bianry`.

### Source Structure

```json
{
  "source_type": "adguard",
  "accept_extended_rules": false
}
```

### Target Structure

```json
{
  "target_type": "adguard"
}
```

### Source Fields

#### accept_extended_rules

If not enabled, only rule items that can be expressed by `domain`, `domain_suffix`, `domain_regex` will be parsed, and other items will be ignored.

Otherwise, most rules supported by AdGuard DNS Filter will be supported, but can only be converted to and from sing-box rule-set binary.

For compatibility, see [AdGuard DNS Filter](https://sing-box.sagernet.org/configuration/rule-set/adguard/).
