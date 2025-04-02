# Delivery Order Assignment Service

## Overview
This project is a **microservices-based delivery order assignment system** built with **Go**. It provides a **REST API** to accept delivery orders and automatically assigns them to available couriers using an **optimized allocation algorithm** inspired by the **Knapsack problem**. 

## Features
- **Microservices Architecture** – Decoupled services for scalability and maintainability.
- **Automated Order Assignment** – Dynamically assigns orders to couriers based on an optimal selection algorithm.
- **REST API** – Exposes endpoints for order creation, courier availability, and assignment status.
- **Optimal Dispatching Algorithm** – Uses a strategy similar to **Knapsack** to maximize efficiency.
- **Unit & Integration Tests** – Comprehensive testing to ensure reliability.

## Tech Stack
- **Golang** – Backend services.
- **REST API** – Communication between services.
- **PostgreSQL** – Persistent storage.
- **Docker** – Containerization.
- **CI/CD** – Automated testing and deployment.

## Getting Started
### Prerequisites
- Go 1.20+
- Docker & Docker Compose
- PostgreSQL & Redis

### Installation
1. Clone the repository:
   ```sh
   git clone https://github.com/ametow/go-microservice.git
   cd go-microservice
   ```

2. Run the services:
   ```sh
   docker-compose up --build
   ```

3. API Documentation (if applicable):
   ```
   Open Swagger UI at http://localhost:8080/swagger
   ```

## Running Tests
```sh
go test ./...
```

## API Endpoints
| Method | Endpoint | Description |
|--------|---------|-------------|
| POST   | `/orders` | Create a new delivery order |
| GET    | `/orders/{id}` | Get order status |
| POST   | `/couriers` | Register a courier |
| GET    | `/couriers/assignments` | List courier assignments |
| GET    | `/meta-info/:courier_id` | Courier meta data |

For more refer to code.

## License
This project is licensed under the MIT License.

## Contribution
Feel free to open issues and contribute via pull requests!
