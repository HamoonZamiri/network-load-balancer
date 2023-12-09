## Network Load Balancer (TCP/UDP)

### Description

This project was created with the intent of exploring Computer Networks specifically in
relation to network layer (OSI layer 4) load balancing. The project is a simple TCP/UDP load balancer that
supports the following features:

-   Round Robin Load Balancing of multiple servers
-   Health Checks which dynamically remove servers from the pool if they are not responding
-   Support for both TCP and UDP protocols
-   Error handling for invalid/failed requests, dropped connections, etc.

### Contributors

-   Ahsan Saeed
-   Hamoon Zamiri

For a detailed acount of the contributions of each of us, please see the commit history.

### Implementation Details

The load balancer is implemented in Go (Golang) and makes extensive use of the package that ships with the standard library "net". The code is relatively easy to follow, the net package handles most of the protocal connections, and the load balancer itself is implemented as a simple struct with a few methods. The load balancer is implemented as a struct, simple methods are present for round robin load balancing and health checks. The code utilizes go routines and concurrency to handle multiple connections at once.

### Usage and Testing

 **Simple Example Using One Server:**
 Run backend server:
 `go run server.go -server=localhost:8081`

 Run load balancer:
 `go run load_balancer.go -bind=localhost:8080 -balance=localhost:8081`
 
 Run client:
`go run client.go`

Result in client terminal:
`Response from load balancer:  Hello from the server!`

### Analysis

### Concluding Remarks