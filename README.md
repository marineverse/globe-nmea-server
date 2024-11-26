# Globe NMEA Server

## How to configure OpenCPN

![OpenCPN](docs/nema_server.jpg)

Open Options, select Connections, Add Connection:

Network Connection
TCP
NMEA 0183
Address: localhost
DataPort: 3006

List position: 1
Receive Input on this Port ( ticked yes )

## Usage:

Get your [BOAT_UUID] from https://www.marineverse.com/globe

```
./globe-nmea-server-mac [BOAT_UUID]
```
