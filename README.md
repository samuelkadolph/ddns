## ddns

Simple server for accepting DDNS requests and updating the matching records in
OVHcloud with authentication.

### Configuration

#### OVHcloud API Key

Go to `https://auth.ca.ovhcloud.com/api/createToken`. Create API Key with the
following Rights:

```
GET /domain/zone/*
PUT /domain/zone/*
POST /domain/zone/*
DELETE /domain/zone/*"
```

### Usage

#### OPNsense

Under `Services > Dynamic DNS > Settings`, click + (`Add`).

| Setting | Value |
| :---: | --- |
| Service Type | `custom` |
| Protocol | `DynDNS 2` |
| Server | `server.com or IP` |
| Username | `username` |
| Password | `password` |
| Hostname(s) | `` |
