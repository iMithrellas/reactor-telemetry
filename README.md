# Reactor Telemetry Dashboard

A Minecraft + Go + Grafana system for monitoring in-game reactors(Extreme/Bigger/Big) in real time.

## Features

- 🐢 Data pulled from ComputerCraft reactors
- 📦 HTTP API server written in Go
- 📈 Time-series data pushed to InfluxDB
- 📊 Grafana dashboard visualization


### Getting Started

1. Copy the example environment file:

cp .env.example .env

1. Edit `.env` to configure your InfluxDB settings(NOTE: Ensure InfluxDB endoint coresponds to what is set in `grafana/datasources/influxdb.yaml`):

```env

INFLUX_URL=http://influxdb:8086
INFLUX_TOKEN=your-token
INFLUX_ORG=your-org
INFLUX_BUCKET=your-bucket
```

1. Start the system with Docker Compose:

`docker-compose up --build`

2. The Go API server will listen on all network interfaces on port 8080 by default.
3. In-game reactor stats will be sent via HTTP POST. Set the endpoint in your ComputerCraft in the variable serverIP, default is localhost.
4. The server also supports GET requests at the same endpoint to fetch the latest stats.
5. Grafana will be available (as configured via docker-compose) to visualize the data stored in InfluxDB, included a simple premade dashboard.

### Contributing
Contributions are welcome! Please fork the repository and submit a pull request with new features or bug fixes.

