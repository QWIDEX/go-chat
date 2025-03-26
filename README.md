# go-chat

A simple chat application built with Go, Gin, and MongoDB.

## Features

- User registration and authentication
- Create and join chat rooms
- Send and receive messages in real-time using WebSockets
- MongoDB integration for storing user and chat data

## Prerequisites

- Go 1.23 or later
- Docker (for running MongoDB)
- Make

## Setup

1. Clone the repository:

    ```bash
    git clone https://github.com/yourusername/go-chat.git
    cd go-chat
    ```

2. Copy the .env.default file to .env and update the environment variables as needed:

    ```bash
    cp .env.default .env
    ```

3. Install dependencies:

    ```bash
    go mod download
    ```

## Running the Application

### Using Docker

1. Build and run the Docker containers:

    ```bash
    make docker-run
    ```

2. Build and run the application:

    ```bash
    make build
    make run
    ```

## Makefile Commands

- Build the application:

    ```bash
    make build
    ```

- Run the application:

    ```bash
    make run
    ```

- Live reload the application:

    ```bash
    make watch
    ```

- Clean up binary from the last build:

    ```bash
    make clean
    ```

## API Endpoints

### Authentication

- `POST /auth/register` - Register a new user
- `POST /auth/login` - Login a user
- `GET /auth/refresh-token` - Refresh access token

### Users

- `GET /users` - Get all users
- `GET /users/:uid` - Get user data by UID

### Chats

- `GET /chats` - Get all chat rooms for the authenticated user
- `POST /chats` - Create a new chat room
- `PATCH /chats` - Add a member to a chat room
- `GET /chats/:chatId` - Get chat history by chat ID

### WebSocket

- `GET /ws` - Connect to WebSocket for real-time messaging
