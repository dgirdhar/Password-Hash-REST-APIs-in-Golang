# Password-Hash-REST-APIs-in-Golang
REST APIs for password hash in Golang

Password Hash Service

This service exposes following 4 REST APIs.

1. /hash -> This end point provides a REST API to calculate the hash of given password. Hash is not calculated and returned immediately. This API returns ID and this can be used later on to find out the hash og given password.
2. /hash/{id} -> This end point provides the hash of password given in /hash API. Pass the ID returned in the /hash API in this API to find the hash of password.
3. /stats -> This end point provides the statistics of hash API.
4. /shutdown -> This end point helps caller to shutdown the service gracefully.


Note: This service runs on 7777 port. In this version it is not taking port number in command line arguments. In later versions we will enhance this.

Run Instructions: Run following command to run this service in localhost
go run main.go

TODOs are mentioned in the files.
