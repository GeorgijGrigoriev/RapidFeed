# RapidFeed: Lightweight RSS Reader Server 

RapidFeed is an open-source RSS reader server written in Go (Golang). Designed with legacy devices in mind, particularly older iOS devices like iPad 2, it offers a lightweight and efficient solution without the need for JavaScript.

## Features

- **Cross-Platform Compatibility**: Optimized for older devices such as iPad 2.
- **No JavaScript Required**: A fully server-rendered application providing a smooth user experience on any device.
- **High Performance**: Built with Go for fast and reliable performance.
- **RSS Aggregation**: Collects and displays RSS feeds from various sources.
- **Easy Installation**: Simple setup process for users.

## Getting Started

### Prerequisites

- Go 1.16 or later
- Git installed on your machine

### Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/yourusername/rapidfeed.git
   cd rapidfeed
   ```

2. **Build and Run**

   ```bash
   go build -o rapidfeed cmd/main.go
   ./rapidfeed
   ```
    OR
   ```bash
    make build
   ./rapidfeed-1.0.0-linux-amd64 (for example)
   ```
3. **Configuration**

   By default app is configured via environment variables with this default values:
   ```bash
      LISTEN: ":8080" #host:port where RapidFeed will listen for incoming connections
      SECRET_KEY: "strong-secretkey" #consider to change this before first run
      REGISTRATION_ALLOWED: true #allow or disallow self user registration on RapidFeed server
   ```
4. **Access the Application**

   Open your web browser and navigate to `http://localhost:8080`. Adjust the port number if necessary based on your configuration. Default user is **admin**, default **password is shown once on first app start**, consider add new admin and block default or change password

## Contributing

We welcome contributions from the community! Please fork the repository, make your changes, and submit a pull request. For major changes, please open an issue first to discuss what you would like to change.


## Acknowledgements

We appreciate the support and contributions from all contributors to this project.

Thank you for choosing RapidFeed as your RSS reader server solution!
