# Duplex Communication System

This project enables duplex communication where multiple users are connected to a system network. When one user sends a message, it is broadcasted to all other users along with their IPv4 address and location for easy tracking by authorized authorities. The data is stored in the server for potential use as evidence in investigations.

## Features

- Multiple user communication
- IPv4 address and location tracking
- Data storage for evidence
- Backend in Go, Frontend in React
- MongoDB for database storage

## Prerequisites

- [Go](https://golang.org/doc/install) installed on your system.
- [Node.js](https://nodejs.org/en/) for the frontend (React) development.
- [MongoDB](https://www.mongodb.com/) running locally on port 2027.

## Setup Instructions
1 Clone the project

### Backend (Go)

1. Navigate to the backend directory.
2. Install necessary Go dependencies.
3. Run the backend server with:
    ```bash
    go run main.go
    ```
4. Ensure MongoDB is running locally on port 2027.

### Frontend (React)

1. Navigate to the frontend directory.
2. Install necessary npm dependencies:
    ```bash
    npm install
    ```
3. Start the frontend server:
    ```bash
    npm start
    ```

## Database

The project uses MongoDB for storing user data and messages, which is hosted locally at `localhost:2027`.

## Running the Project

1. Start the backend server (Go).
2. Start the frontend (React).
3. The project is now running, and users can connect to send and receive messages with tracking enabled.

## License

This project is licensed under the MIT License.

