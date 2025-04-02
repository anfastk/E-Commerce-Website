# ***E-Commerce Website (Golang + Gin + GORM + PostgreSQL + AWS)***

## 📌 Project Overview

This is a full-fledged e-commerce website built using the **Gin** framework in **Golang**, following the **MVC architecture**. The project includes both admin and user sides, handling product management, stock management, user authentication, order processing, and payment integration with **Razorpay**.

## 🚀 Features

- User authentication (signup, login, logout, password reset)

- Admin dashboard for product and order management

- Product catalog with categories and filters

- Shopping cart and checkout system

- Secure payments via **Razorpay**

- Stock management to prevent overselling

- Order tracking system

- Responsive UI using **HTML**, and **Tailwind Css**

- Logging using **Zap**

- Hosted on **AWS with Nginx**, secured with **HTTPS/TLS**

## 🛠️ Tech Stack

- **Backend:** Golang (Gin framework)

- **Database:** PostgreSQL with GORM ORM

- **Frontend:** HTML, Tailwind Css

- **Logging:** Zap Logger

- **Payment Gateway:** Razorpay

- **Deployment:** AWS (EC2, S3, RDS, Nginx, TLS/SSL)

- **Security:** HTTPS with TLS

## Installation and Setup

### Prerequisites:

- Golang installed

- PostgreSQL database setup

- AWS instance with Nginx configured

## Steps to Run the Project:

### 1️⃣ Clone the repository:

```sh 
git clone https://github.com/anfastk/E-Commerce-Website.git


### 2️⃣ Set up environment variables:
Create a `.env` file and configure database credentials, AWS settings, and Razorpay keys.

```sh
PORT=8080
DB=host=your-db-host user=your-db-user password=your-db-password dbname=your-db-name port=your-db-port
SECRETKEY=your-jwt-secretkey
CLOUDINARY_CLOUD_NAME=your-cloudinary-cloud-name
CLOUDINARY_API_KEY=your-cloudinary-api-key
CLOUDINARY_API_SECRET=your-cloudinary-api-secret
SENDGRID_API_KEY=your-sendgrid-api-key
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=your-google-redirecturl
DEFAULT_PROFILE_PIC=https://res.cloudinary.com/dghzlcoco/image/upload/v1740382266/e3b0c44298fc1Default_c149afbf4c8996fb92427aImagee41e4649b934ca4959Profile91b7852b855_rlwzij.jpg
RAZORPAY_KEY_ID=your-razorpay-key-id
RAZORPAY_KEY_SECRET=your-razorpay-key-secret

### 3️⃣ Install dependencies:

```sh
go mod tidy

### 4️⃣ Start the server

```sh
go run main.go

### 5️⃣ Access the website:

- **User:**`http://localhost:8080`

- **Admin:** `http://localhost:8080/admin/login`

## Deployment on AWS with Nginx

1. Set up an EC2 instance and install Golang & PostgreSQL.

2. Clone the repository and set up environment variables.

3. Install and configure Nginx to reverse proxy the Golang server.

4. Set up SSL/TLS security using Let's Encrypt.

5. Run the application in production mode.