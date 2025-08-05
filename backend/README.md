# Online Judge - Backend

This is the backend for an online judge platform, built with **Spring Boot**. It provides RESTful APIs for managing users, problems, submissions, and authentication.

## Features

* **User Authentication:** Secure user registration and login using JWT (JSON Web Tokens).
* **Problem Management:** CRUD operations for programming problems, including test cases and tags.
* **Code Submissions:** Submit code for a specific problem and receive a verdict.
* **Role-Based Access Control:** Differentiates between regular users and admins, with specific permissions for certain actions.

## Technologies Used

* **Java 21**
* **Spring Boot**
* **Spring Security** for authentication and authorization.
* **Spring Data JPA** for database interactions.
* **PostgreSQL** as the relational database.
* **Maven** for project management.
* **JUnit 5**, **Mockito**, and **Testcontainers** for testing.
* **Spotless** for code formatting.
* **JaCoCo** for code coverage analysis.

## API Endpoints

The following are the main API endpoints provided by the backend:

### Authentication

* `POST /api/v1/auth/register`: Register a new user and receive a JWT.
* `POST /api/v1/auth/login`: Authenticate an existing user and receive a JWT.
* `POST /api/v1/auth/logout`: Log out the current user.

### Problems

* `GET /api/v1/problems/list?page={page}`: Get a list of problems for a given page.
* `GET /api/v1/problems/{problemId}`: Get a single problem by its ID.
* `POST /api/v1/problems`: Create a new problem (admin only).

### Submissions

* `GET /api/v1/submissions/list?page={page}`: Get a list of submissions for a given page.
* `GET /api/v1/submissions/{submissionId}`: Get details of a specific submission by its ID.
* `POST /api/v1/submissions`: Make a submission to a problem (authenticated user only).