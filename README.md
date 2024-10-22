# Go-JMeter-API

Go-JMeter-API is a simple web application that integrates JMeter with a Go backend, allowing users to run JMeter tests and view results through a web interface.

## Features

- Run JMeter tests via a web interface
- View test results in real-time
- Generate and access JMeter reports
- List all previous test reports

## Prerequisites

- Go (version 1.16 or later)
- JMeter (version 5.0 or later)

## Setup

1. Clone the repository:
   ```
   git clone https://github.com/mkilic20/Go-JMeter-Api.git
   ```

2. Navigate to the project directory:
   ```
   cd Go-JMeter-Api
   ```

3. Install dependencies:
   ```
   go mod tidy
   ```

4. Ensure JMeter is installed and the path is correctly set in the `main.go` file.

## Running the Application

1. Start the server:
   ```
   go run main.go
   ```

2. Open a web browser and navigate to `http://localhost:8080`

## Usage

1. Fill out the test configuration form on the web interface.
2. Click "Run Test" to start a JMeter test.
3. View real-time results as the test progresses.
4. Access detailed reports after the test completes.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
