# Online Judge

A full-stack web application for solving Data Structures and Algorithms (DSA) problems. This platform allows users to practice coding problems, run their code against predefined test cases, and get instant feedback. Admins can manage problem sets and test cases.

## Features

### For Users
- Register and log in
- Browse and solve coding problems
- Submit code in multiple languages (Java, C++, Python, JavaScript)
- Run submissions against pre-defined test cases
- View submission history

### For Admins
- Create, update, or delete problems
- Upload test cases
- View all user submissions and stats

## Tech Stack

### Backend
- **Spring Boot (Java)**
- **PostgreSQL**
- **Spring Security** with JWT-based authentication
- **RabbitMQ88 for communicating with the Executor

### Frontend
- **React**

### Executor
- **Dockerized microservice** for executing submitted code securely in Go
