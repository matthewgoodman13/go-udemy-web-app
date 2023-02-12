# Building Web Applications with Go - Udemy
## Course
- Course: https://www.udemy.com/course/building-web-applications-with-go-intermediate-level
- Certificate of Completion: https://www.udemy.com/certificate/UC-8dfd1688-edcf-47aa-b5ea-bf12f18329e9/

## The Project
- Built an E-commerce application consisting of a front-end, back-end, and a microservice all built using the Go (or Golang) programming language. Used MariaDB as database solution and DBeaver as the database navigator.
- Platform allows customers to buy or subscribe to receive “Widgets” using Stripe Payment Integration
    - Added Virtual Terminal to receive credit card info and charge a user (as if over the phone)
    - Added Checkout to allow customer to buy directly off website or subscribe to a package
- Created Invoice Microservice that emails a PDF invoice after completion of a purchase
- Authentication done via Stateful Tokens and using private routes
- Forgot/Reset Password functionality using email reset link via protected hash in email
- Used MailTrap to create a fake email server / fake inbox
- Created Admin dashboard that displays purchases and subscriptions with pagination/chunking of data and ability to create/delete adminstrative accounts
- Added socket integration to logout “deleted” users automatically

## Running the Project
- cd into go-stripe directory
- Run web server
    - ~~`go run ./cmd/web`~~
    - `air` from within directory (live reload - https://github.com/cosmtrek/air)

    - make start or make start_front or make start_back
    - make stop or make stop_front or make stop_back
    - make start_invoice or make stop_invoice to start/stop invoice microservice
