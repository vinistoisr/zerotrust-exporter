# Zero Trust Exporter

Zero Trust Exporter is a Prometheus exporter written in Go that collects and exposes metrics from Cloudflare's Zero Trust API. It is designed to provide visibility into devices, users, and tunnels managed by Cloudflare Zero Trust.

## Features
- Collects metrics for devices, users, and tunnels from Cloudflare Zero Trust API
- Provides detailed scrape duration and API call metrics
- Supports both command-line flags and environment variables for configuration
- Docker support for containerized deployments

## Go Libraries Used

- [VictoriaMetrics/metrics](https://github.com/VictoriaMetrics/metrics) - Metrics library for Prometheus.
- [cloudflare/cloudflare-go](https://github.com/cloudflare/cloudflare-go) - Cloudflare API client for Go.

## Metrics Collected

| Metric Name                                          | Description                                     | Labels                                     | Type      |
| ---------------------------------------------------- | ----------------------------------------------- | ------------------------------------------ | --------- |
| `zerotrust_exporter_up`                              | Exporter up status                              | -                                          | Gauge     |
| `zerotrust_exporter_scrape_duration_seconds`         | Duration of the scrape in seconds               | -                                          | Histogram |
| `zerotrust_exporter_api_calls_total`                 | Total number of API calls made                  | -                                          | Counter   |
| `zerotrust_exporter_api_errors_total`                | Total number of API errors encountered          | -                                          | Counter   |
| `zerotrust_devices_up`                               | Device up status                                | device_type, id, ip, user_id, user_email, name | Gauge     |
| `zerotrust_users_access_seat`                        | User access seat status                         | email, id, gateway_seat                    | Gauge     |
| `zerotrust_tunnels_status`                           | Tunnel status                                   | id, name                                   | Gauge     |


## Environment Variables

| Variable      | Description                                    | Default Value |
| ------------- | ---------------------------------------------- | ------------- |
| `API_KEY`     | Cloudflare API key (required)                  | -             |
| `ACCOUNT_ID`  | Cloudflare account ID (required)               | -             |
| `DEBUG`       | Enable debug mode (true/false)                 | false         |
| `DEVICES`     | Enable devices metrics (true/false)            | false         |
| `USERS`       | Enable users metrics (true/false)              | false         |
| `TUNNELS`     | Enable tunnels metrics (true/false)            | false         |
| `INTERFACE`   | Listening interface (default: any)             | ""            |
| `PORT`        | Listening port (default: 9184)                 | 9184          |

## Command-Line Flags

| Flag          | Description                                    | Default Value |
| ------------- | ---------------------------------------------- | ------------- |
| `-apikey`     | Cloudflare API key (required)                  | -             |
| `-accountid`  | Cloudflare account ID (required)               | -             |
| `-debug`      | Enable debug mode (true/false)                 | false         |
| `-devices`    | Enable devices metrics (true/false)            | false         |
| `-users`      | Enable users metrics (true/false)              | false         |
| `-tunnels`    | Enable tunnels metrics (true/false)            | false         |
| `-interface`  | Listening interface (default: any)             | ""            |
| `-port`       | Listening port (default: 9184)                 | 9184          |

## Usage

### Docker Deployment

1. Build the Docker image:

    ```sh
    docker build -t zerotrust-exporter .
    ```

2. Run the Docker container:

    ```sh
    docker run -d -p 9184:9184 --env-file .env zerotrust-exporter
    ```
### Running Locally

1. Clone the repository:

    ```sh
    git clone https://github.com/yourusername/zerotrust-exporter.git
    cd zerotrust-exporter
    ```

2. Set up environment variables in a `.env` file or export them directly:

    ```plaintext
    API_KEY=your_api_key
    ACCOUNT_ID=your_account_id
    DEBUG=true
    DEVICES=true
    USERS=true
    TUNNELS=true
    INTERFACE=0.0.0.0
    PORT=9184
    ```

3. Run the exporter:

    ```sh
    go run .
    ```

### Building and Running with Go

1. Build the Go binary:

    ```sh
    go build -o zerotrust-exporter .
    ```

2. Run the binary:

    ```sh
    ./zerotrust-exporter -apikey=your_api_key -accountid=your_account_id -debug=true -devices=true -users=true -tunnels=true -interface=0.0.0.0 -port=9184
    ```



## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
