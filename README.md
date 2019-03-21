# Prometheus Exporter for Kaifa meter

Prometheus Exporter for Kaifa Power Meters used in Norway

The project is written i Go

## Setup
- Rasberry PI Zero W
- USB MBUS Slave Module, example from [AliExpress](https://www.aliexpress.com/item/Freeshipping-USB-to-MBUS-slave-module-discrete-component-non-TSS721-circuit-M-BUS-bus-data-monitor/32814808312.html)

Meter-Bus uses two wires for communication. The Kaifa meter has a RJ45 plug, where the two left-most ones are used 
(orange cables in the T568B standard).


## Background
This project is inspired by https://github.com/roarfred/AmsToMqttBridge

Relevant reading:
- https://www.nek.no/info-ams-han-utviklere/
- https://www.nek.no/wp-content/uploads/2018/10/AMS-HAN-Port-Smart-Hus-og-Smart-Bygg-Gj%C3%B8r-det-selv-og-Pilotprosjekter-ver-1.16.pdf
- https://www.nek.no/wp-content/uploads/2018/11/Kaifa-KFM_001.pdf
- https://drive.google.com/drive/folders/0B3ZvFI0Dg1TDbDBzMU02cnU0Y28
- https://drive.google.com/open?id=1c3f0D52ZxRLzoG60Sj68kE0U0XDaYWQI 

### OBIS list information

'\x7E' is used as start and stop bit

Example of byte array
```
/*
7e                                                     : Flag (0x7e)
a0 87                                                  : Frame Format Field
01 02                                                  : Source Address
01                                                     : Destination Address
10                                                     : Control Field = R R R P/F S S S 0 (I Frame)
9e 6d                                                  : HCS
e6 e7 00                                               : DLMS/COSEM LLC Addresses
0f 40 00 00 00                                         : DLMS HEADER?
09 0c 07 d0 01 03 01 0e 00 0c ff 80 00 03              : Information
02 0e                                                  : Information
09 07 4b 46 4d 5f 30 30 31                             : Information
09 10 36 39 37 30 36 33 31 34 30 30 30 30 30 39 35 30  : Information
09 08 4d 41 31 30 35 48 32 45                          : Information
06 00 00 00 00                                         : Information
06 00 00 00 00                                         : Information
06 00 00 00 00                                         : Information
06 00 00 00 00                                         : Information
06 00 00 00 0e                                         : Information
06 00 00 09 01                                         : Information
09 0c 07 d0 01 03 01 0e 00 0c ff 80 00 03              : Information
06 00 00 00 00                                         : Information
06 00 00 00 00                                         : Information
06 00 00 00 00                                         : Information
06 00 00 00 00                                         : Information
97 35                                                  : FCS
7e                                                     : Flag
*/
```

### Data fields

All data fields consists of a single byte giving the type preceding the actual data.

```
| Type | Length    |                      |
| ---- | --------- | -------------------- |
| 0x0A | OBIS code value                  |
| 0x09 | Variable* | ASCII value          |
| 0x02 | 1 byte    | byte value           |
| 0x12 | 2 bytes   | integer value uint16 |
| 0x06 | 4 bytes   | integer value uint32 |

* Length is value of first byte
```

## Compile and run

### Dependencies
```bash
go get github.com/prometheus/client_golang/prometheus/promhttp
go get github.com/tarm/serial
```

### Compile
```bash
# Rasberry PI Zero W
env GOOS=linux GOARCH=arm GOARM=5 go build
```

##### Tip

For smaller binary, strip debug info:
```
env GOOS=linux GOARCH=arm GOARM=5 go build -ldflags="-s -w"
```

[upx](https://github.com/upx/upx) may also be applied

### systemd
Binary is located at /usr/bin/kaifa-exporter

```
sudo cat << EOF > /etc/systemd/system/kaifa-exporter.service
[Unit] 
Description=kaifa-exporter 
After=network-online.target

[Service] 
ExecStart=/usr/usr/bin/kaifa-exporter
Restart = always
[Install] 
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable kaifa-exporter.service
sudo systemctl start kaifa-exporter.service
```

### My setup
I am running this in a Rasberry Pi Zero W.

This is running autossh to create a reverse tunnel to a t3.nano instance in AWS, where my Prometheus instance is running.
