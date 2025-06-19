---
icon: material/package
---

# Package Manager

## :material-tram: Repository Installation

=== ":material-debian: Debian / APT"

    ```bash
    sudo mkdir -p /etc/apt/keyrings &&
       sudo curl -fsSL https://sing-box.app/gpg.key -o /etc/apt/keyrings/sagernet.asc &&
       sudo chmod a+r /etc/apt/keyrings/sagernet.asc &&
       echo '
    Types: deb
    URIs: https://deb.sagernet.org/
    Suites: *
    Components: *
    Enabled: yes
    Signed-By: /etc/apt/keyrings/sagernet.asc
    ' | sudo tee /etc/apt/sources.list.d/sagernet.sources &&
       sudo apt-get update &&
       sudo apt-get install srsc # or srsc-beta
    ```

=== ":material-redhat: Redhat / DNF 5"

    ```bash
    sudo dnf config-manager addrepo --from-repofile=https://sing-box.app/sing-box.repo &&
    sudo dnf install srsc # or srsc-beta
    ```

=== ":material-redhat: Redhat / DNF 4"

    ```bash
    sudo dnf config-manager --add-repo https://sing-box.app/sing-box.repo &&
    sudo dnf -y install dnf-plugins-core &&
    sudo dnf install srsc # or srsc-beta
    ```

## :material-download-box: Manual Installation

The script download and install the latest package from GitHub releases
for deb or rpm based Linux distributions, ArchLinux and OpenWrt.

```shell
curl -fsSL https://sing-box.app/srsc/install.sh | sh
```

or latest beta:

```shell
curl -fsSL https://sing-box.app/srsc/install.sh | sh -s -- --beta
```

or specific version:

```shell
curl -fsSL https://sing-box.app/srsc/install.sh | sh -s -- --version <version>
```
    ```

## :material-book-multiple: Service Management

For Linux systems with [systemd][systemd], usually the installation already includes a srsc service,
you can manage the service using the following command:

| Operation | Command                                   |
|-----------|-------------------------------------------|
| Enable    | `sudo systemctl enable srsc`              |
| Disable   | `sudo systemctl disable srsc`             |
| Start     | `sudo systemctl start srsc`               |
| Stop      | `sudo systemctl stop srsc`                |
| Kill      | `sudo systemctl kill srsc`                |
| Restart   | `sudo systemctl restart srsc`             |
| Logs      | `sudo journalctl -u srsc --output cat -e` |
| New Logs  | `sudo journalctl -u srsc --output cat -f` |

[systemd]: https://systemd.io/