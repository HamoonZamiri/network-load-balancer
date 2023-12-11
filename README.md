## Network Load Balancer (TCP/UDP)

### Description

This project was created with the intent of exploring Computer Networks specifically in
relation to the transport layer (OSI layer 4) and load balancing. The project is a simple 
TCP/UDP load balancer that supports the following features:

-   Round Robin Load Balancing of multiple servers
-   Health Checks which dynamically remove servers from the pool if they are not responding
-   Support for both TCP and UDP protocols
-   Error handling for invalid/failed requests, dropped connections, etc.

### Contributors

-   Ahsan Saeed
-   Hamoon Zamiri

For a detailed acount of each member's contribution, please see the commit history.

### Implementation Details

The load balancer is implemented in Go (Golang) and makes extensive use of the package that ships with the standard library "net". The code is relatively easy to follow. The net package handles most of the protocal connections, and the load balancer itself is implemented as a simple struct with a few methods that are present for round robin load balancing and health checks. The code utilizes go routines and concurrency to handle multiple connections at once. Further documentation can be found in the source code itself.

### Usage and Testing

 **Simple TCP Example Using One Server:**
 1. Run backend server:
 ```bash
 go run server.go -server=localhost:8081
 ```

 2. Run load balancer:
 ```bash
 go run load_balancer.go -bind=localhost:8080 -balance=localhost:8081
 ```
 
 3. Run client:
```bash
go run client.go
```

Result in client terminal:
```bash
Response from load balancer:  Hello from server, you sent First message to load balancer
```

**Running with UDP and/or Multiple Servers:**
In order to run with UDP, simply add the `-udp` flag when running the server, load balancer, and client.
In order to connect the load balancer with several servers, run the load balancer like this: 
 ```bash
 go run load_balancer.go -bind=localhost:8080 -balance=localhost:8081,localhost:8082
 ```

### Analysis
**TCP vs UDP Handling:**
-   TCP connections involve establishing and maintaining a connection, which is well-suited for stateful communication.
-   UDP connections are stateless, making them more suitable for scenarios with lower latency requirements.

**Scalability:**
-   The load balancer architecture allows for easy scaling by adding more backend servers.
-   Scalability is achieved without impacting the existing infrastructure.

**Results:**
-   The load balancer successfully distributes requests among backend servers for both TCP and UDP.
-   The system handles multiple clients and servers concurrently without loss of functionality.

### Conclusion
This project provided valuable insights into load balancing, network programming, and protocol-specific considerations. In particular, the project involved topics like TCP and UDP communication, load balancing techniques, and troubleshooting network-related issues. Utilizing these concepts, we successfully implemented a robust load balancer that caters to both TCP and UDP scenarios, meeting the project's objectives.