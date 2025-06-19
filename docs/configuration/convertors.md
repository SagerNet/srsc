# Convertors

#### source

sing-box rule-set source.

As an intermediate format for all formats, it can be converted from any all other formats
except `adguard` and `bianry` which is converted from `adguard`.

#### binary

sing-box rule-set binary.

As the original purpose of this project, it should be possible to convert from all other formats.

#### adguard

AdGuard DNS Filer.

Special rule-set supported by sing-box, covering most of the syntax supported by AdGuard Home.

Since it contains many other rules that cannot be represented in any other format,
it can only be converted to and from `bianry`.

#### clash-text

Clash rule provider, in plain text format.

Only `domain` and `ipcidr` formats are supported, that is, there is no support for `classical` format yet.

Only `domain`, `domain_suffix` and `ip_cidr` rule items that can be expressed by sing-box will be converted,
that is, only plain text domain names and plain text domain names starting with `+.` are supported, 
but other domain names containing special matching symbols `+` and `*` will be ignored.

#### clash-yaml

Clash rule provider, in YAML format.

Same compatibility as `clash-text`.

#### mrs

Clash.Meta (mihomo) binary rule-set.

Same compatibility as `clash-text`.
