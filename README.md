# 🎯 E-Commerce Website (Golang + Gin + GORM + PostgreSQL + AWS)

## 📌 Project Overview

This is a full-fledged e-commerce website built using the **Gin** framework in **Golang**, following the **MVC architecture**. The project includes both admin and user sides, handling:  
✅ Product management  
✅ Stock management  
✅ User authentication  
✅ Order processing  
✅ Payment integration with **Razorpay**  

## 🚀 Features

✔️ **User Authentication** – Signup, login, logout, password reset 

✔️ **Admin Dashboard** – Manage products and orders  

✔️ **Product Catalog** – Categories, search, and filters  

✔️ **Shopping Cart & Checkout** – Seamless user experience  

✔️ **Secure Payments** – Integrated with **Razorpay**  

✔️ **Stock Management** – Prevents overselling  

✔️ **Order Tracking** – Status updates for users  

✔️ **Responsive UI** – Built with **HTML + Tailwind CSS**  

✔️ **Logging** – Implemented using **Zap Logger**  

✔️ **Hosting & Security** – AWS, Nginx, HTTPS/TLS  

## 🛠️ Tech Stack

- **Backend:** Golang (Gin framework)

- **Database:** PostgreSQL with GORM ORM

- **Frontend:** HTML, Tailwind Css

- **Logging:** Zap Logger

- **Payment Gateway:** Razorpay

- **Deployment:** AWS (EC2, S3, RDS, Nginx, TLS/SSL)

- **Security:** HTTPS with TLS

## Installation and Setup

### 📌 Prerequisites:

✔️ Golang installed

✔️ PostgreSQL database setup

✔️ AWS instance with Nginx configured

## Steps to Run the Project:

### 1️⃣ Clone the repository:
```sh
git clone https://github.com/anfastk/E-Commerce-Website.git
cd E-Commerce-Website 
```

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
```

### 3️⃣ Install dependencies:

```sh
go mod tidy
```

### 4️⃣ Start the server

```sh
go run main.go
```

### 5️⃣ Access the website:

- **User Panel:** [http://localhost:8080](http://localhost:8080)  
- **Admin Panel:** [http://localhost:8080/admin/login](http://localhost:8080/admin/login)  


## 🌍 Deployment on AWS with Nginx

1️⃣ Set up an EC2 instance and install Golang & PostgreSQL.

2️⃣ Clone the repository and set up environment variables.

3️⃣ Install and configure Nginx to reverse proxy the Golang server.

4️⃣ Set up SSL/TLS security using Let's Encrypt.

5️⃣ Run the application in production mode.

6️⃣ Run the Application – Start the Golang server in production mode.

