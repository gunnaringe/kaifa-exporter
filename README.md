# Prometheus Exporter for Kaifa meter
env GOOS=linux GOARCH=arm GOARM=5 go build

https://github.com/roarfred/AmsToMqttBridge/blob/master/Samples/Kaifa/obisdata.md

https://www.nek.no/info-ams-han-utviklere/
https://www.nek.no/wp-content/uploads/2018/10/AMS-HAN-Port-Smart-Hus-og-Smart-Bygg-Gj%C3%B8r-det-selv-og-Pilotprosjekter-ver-1.16.pdf
## OBIS list information
https://www.nek.no/wp-content/uploads/2018/11/Kaifa-KFM_001.pdf

https://drive.google.com/drive/folders/0B3ZvFI0Dg1TDbDBzMU02cnU0Y28

https://www.hjemmeautomasjon.no/forums/topic/390-ny-str%C3%B8mm%C3%A5ler-med-han-interface/?page=5


'\x7E' is used as start and end delimiter
Can maybe be found as part of message as well...

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


https://github.com/roarfred/AmsToMqttBridge/blob/master/Code/Arduino/HanReader/src/DlmsReader.cpp

Data types
0x0A OBIS code value
0x09 string value (next byte is length)
0x02 byte value (1 byte)
0x12 integer value (2 bytes) uint16
0x06 integer value (4 bytes) uint32


// TODO: Checksum
