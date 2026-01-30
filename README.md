<h1 align='center'>
    Restaurant Management Backend
</h1>

<p align='center'>
    Example of a complete backend system for restaurant management.
</p>

<p align='center'>
    <a href="https://go.dev">
        <img alt="Golang Logo" src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
    </a>
    &nbsp;
    <a href="https://www.mongodb.com">
        <img alt="MongoDB Logo" src="https://img.shields.io/badge/MongoDB-4EA94B?style=for-the-badge&logo=mongodb&logoColor=white" />
    </a>
    &nbsp;
    <a href="https://docs.docker.com/compose">
        <img alt="Docker Compose Logo" src="https://img.shields.io/badge/Docker%20Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white" />
    </a>
    &nbsp;
    <a href="https://www.jwt.io">
        <img alt="JWT Logo" src="https://img.shields.io/badge/JWT-000000?style=for-the-badge&logo=JSON%20web%20tokens&logoColor=white" />
    </a>
    &nbsp;
    <a href="https://www.postman.com">
        <img alt="Postman Logo" src="https://img.shields.io/badge/Postman-FF6C37?style=for-the-badge&logo=Postman&logoColor=white" />
    </a>
</p>

## 💡 Project Overview

This project is a complete Golang backend system for restaurant management. It's my version of the project created by [Akhil Sharma](https://www.youtube.com/@AkhilSharmaTech) with this [video tutorial playlist](https://youtube.com/playlist?list=PL5dTjWUk_cPbjazI1vRuTRZi6o5QlVAAR&si=vPB6STutIOLT570I) on YouTube.

The idea for this project was to improve my Golang skills with interesting topics such as:

* Creating an API
* Adding a login system
* Connecting Golang to a MongoDB database

## 📚 Prerequisites

To run this project, you need the following prerequisites:

* Golang (version 1.24 was used for this project)
* An API testing tool like Postman
* Docker and Docker Compose (to run MongoDB easily)

## 📂 Structure

We will describe the different folders of this project:

* **routes**: describes all of the API routes
* **models**: defines all of the struct elements
* **middleware**: defines *authMiddleware* for the authentication system
* **helpers**: utility functions
* **database**: creates the connection to the MongoDB database
* **controllers**: handles the business logic of the application and connects routes to models

## 👩‍💻 Commands

The easiest way to run this project is by using Docker Compose. 

> [!NOTE]
>If you don't want to use Docker, make sure you have MongoDB installed and running on your machine. Don't forget to update the MongoDB connection string in the `database/databaseConnection.go` file.

To run the application, you need to use the following command:

```bash
sudo docker compose down
sudo docker compose up -d
```

Now, the backend server should be running on `http://localhost:8000`, and the MongoDB database should be running on the default port `27017`.
You can test the API endpoints using Postman or any other API testing tool (some examples are visible in the Postman collection file: `golang-restaurant-management.postman_collection.json`).

> [!TIP]
> Make sure to set up token in Postman for protected routes.

To stop the application, you can use:

```bash
sudo docker compose down
```

> [!CAUTION]
> If you want to use the project for production, make sure to change the `SECRET_KEY` and use environment variables to store sensitive information like mongodb credentials.