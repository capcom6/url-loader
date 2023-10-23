# URL Loader
[![License](https://img.shields.io/badge/License-Apache-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/capcom6/url-loader)](https://goreportcard.com/report/github.com/capcom6/url-loader)
[![Codecov](https://codecov.io/gh/capcom6/url-loader/branch/master/graph/badge.svg)](https://codecov.io/gh/capcom6/url-loader)


URL Loader is a tool for loading URLs and retrieving statistics such as response time and size of the response body.

## Features

* Load URLs and retrieve statistics
* Configurable timeout for requests
* Adjustable buffer size for reading response bodies

## Installation

To install URL Loader, follow these steps:

1. Clone the repository: git clone https://github.com/capcom6/url-loader.git
2. Navigate to the project directory: cd url-loader
3. Install the dependencies: go mod download
4. Build the binary: go build -o url-loader
5. Run the application: ./url-loader

## Usage

```
Usage: url-loader [options] filename [filenames...]
  -loader-buffer uint
        buffer size in bytes (default 32768)
  -loader-parallel int
        parallel requests (default 8)
  -loader-redirects
        follow redirects (default true)
  -loader-timeout duration
        timeout (default 1s)
  -reader-skip uint
        skip N lines
```

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvement, please open an issue or submit a pull request.

## License

This project is licensed under the Apache License 2.0. See the LICENSE file for details.