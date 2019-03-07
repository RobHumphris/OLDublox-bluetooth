# Ublox-Bluetooth

This implements a small subset of the AT commands defined in the [u-blox Short Range Modules](https://www.u-blox.com/sites/default/files/u-blox-SHO_ATCommands_%28UBX-14044127%29.pdf) documents

Its a Linux only implementation.

sudo network-manager.nmcli con add type gsm ifname cdc-wdm0 con-name VODAPHONE apn wap.vodafone.co.uk user wap password wap
sudo network-manager.nmcli con add type gsm ifname cdc-wdm0 con-name VODAPHONE apn pp.bundle.internet user wap password wap
(didnt work)

sudo network-manager.nmcli con add type gsm ifname cdc-wdm0 con-name VODAPHONE apn internet user web password web


sudo network-manager.nmcli con up VODAPHONE
sudo network-manager.nmcli con down VODAPHONE




sudo network-manager.nmcli con show VODAPHONE

sudo network-manager.nmcli con del VODAPHONE

sudo network-manager.nmcli con add type gsm ifname cdc-wdm0 con-name three apn three.co.uk
