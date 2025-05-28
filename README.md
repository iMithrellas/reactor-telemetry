# Reactor Telemetry Dashboard

A Minecraft + Go + Grafana system for monitoring in-game reactors in real time.

## Features

- 🐢 Data pulled from ComputerCraft reactors
- 📦 HTTP API server written in Go
- 📈 Time-series data pushed to InfluxDB
- 📊 Grafana dashboard visualization

## Getting Started

```bash
cp .env.example .env
docker-compose up --build

