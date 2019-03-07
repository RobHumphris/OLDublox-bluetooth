# Ublox-Bluetooth

This implements a small subset of the AT commands defined in the [u-blox Short Range Modules](https://www.u-blox.com/sites/default/files/u-blox-SHO_ATCommands_%28UBX-14044127%29.pdf) documents

Its a Linux only implementation.

sudo network-manager.nmcli con add type gsm ifname cdc-wdm0 con-name three apn three.co.uk
sudo network-manager.nmcli con up three
sudo network-manager.nmcli con show three
sudo network-manager.nmcli con down three