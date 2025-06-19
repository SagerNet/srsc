---
icon: material/package
---

# 包管理器

## :material-tram: 仓库安装

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
       sudo apt-get install sing-box # or sing-box-beta
    ```

=== ":material-redhat: Redhat / DNF 5"

    ```bash
    sudo dnf config-manager addrepo --from-repofile=https://sing-box.app/sing-box.repo &&
    sudo dnf install sing-box # or sing-box-beta
    ```

=== ":material-redhat: Redhat / DNF 4"

    ```bash
    sudo dnf config-manager --add-repo https://sing-box.app/sing-box.repo &&
    sudo dnf -y install dnf-plugins-core &&
    sudo dnf install sing-box # or sing-box-beta
    ```

## :material-download-box: 手动安装

该脚本从 GitHub 发布中下载并安装最新的软件包，适用于基于 deb 或 rpm 的 Linux 发行版、ArchLinux 和 OpenWrt。

```shell
curl -fsSL https://sing-box.app/srsc/install.sh | sh
```

或最新测试版：

```shell
curl -fsSL https://sing-box.app/srsc/install.sh | sh -s -- --beta
```

或指定版本：

```shell
curl -fsSL https://sing-box.app/srsc/install.sh | sh -s -- --version <version>
```

## :material-book-multiple: 服务管理

对于带有 [systemd][systemd] 的 Linux 系统，通常安装已经包含 serenity 服务，
您可以使用以下命令管理服务：

| 行动   | 命令                                            |
|------|-----------------------------------------------|
| 启用   | `sudo systemctl enable serenity`              |
| 禁用   | `sudo systemctl disable serenity`             |
| 启动   | `sudo systemctl start serenity`               |
| 停止   | `sudo systemctl stop serenity`                |
| 强行停止 | `sudo systemctl kill serenity`                |
| 重新启动 | `sudo systemctl restart serenity`             |
| 查看日志 | `sudo journalctl -u serenity --output cat -e` |
| 实时日志 | `sudo journalctl -u serenity --output cat -f` |

[systemd]: https://systemd.io/