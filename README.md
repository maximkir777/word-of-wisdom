# **TCP Server with Proof-of-Work**

## 1. Project Overview
This project implements a TCP server that responds with "smart phrases" (or any form of useful information) and is protected against potential DoS/DDoS attacks by a dynamic Proof-of-Work (PoW) mechanism.
A client application is provided to interact with the server using a custom-defined protocol over TCP.

**Key Features**:
- Server: Listens on a TCP port, issues challenges to clients, verifies solutions, and returns responses.
- Client: Requests challenges from the server, solves them locally, then requests resources from the server.
- Custom PoW Algorithm: Adjusts difficulty based on server load to mitigate spam or DDoS attempts.
- Custom TCP Protocol: Simplifies communication by using structured message headers and payloads.

---

## 2. Architecture & Components

### 2.1 Server
- Exposes a TCP socket.
- Handles requests using a simple request/response loop.
- Issues PoW challenges to clients and validates their solutions.
- Provides "smart phrases" (or any other resource) on successful PoW validation.

### 2.2 Client
- Connects to the server using TCP.
- Requests a PoW challenge.
- Solves the challenge locally.
- Sends the solution to the server to retrieve the resource.

### 2.3 Proof-of-Work (PoW)
- Dynamically adjusts difficulty based on server load.
- Requires the client to find a nonce that makes the hash start with a certain number of zeros.
- The difficulty level changes over time to balance server load.

### 2.4 Custom Protocol
- Defines message headers (integers) to indicate different types of messages.
- Uses a delimiter `"|"` to separate the header from the payload.

---

## 3. PoW Explanation & Choice

### 3.1 Rationale Behind PoW
- PoW (Proof-of-Work) is used to deter malicious clients from flooding the server with excessive requests.
- By requiring clients to perform a small computation (finding a valid hash), the cost of spamming the server increases significantly.
- This is especially useful in TCP-based services that might otherwise be vulnerable to DDoS or spam attacks.

### 3.2 Algorithm Details
1. **Seed Generation**
   - The server creates a string seed in the format:
     ```
     <difficulty>,<randomNumber>
     ```
   - `difficulty` indicates how many leading zeros the computed hash must have.
   - The server also returns a `challenge` string (all zeros with length equal to the difficulty), which can be used client-side for clarity (though not strictly required to solve the puzzle).

2. **Client-Side Computation**
   - The client takes the seed and computes a proof (`nonce`) such that:
     ```
     sha256(seed + "|" + proof)
     ```
     has a hexadecimal representation starting with `difficulty` zeros.
   - For example, if `difficulty = 5`, the hash must start with `"00000"`.

3. **Difficulty Adjustment**
   - The server tracks recent requests in a sliding window and estimates the current request rate.
   - Based on the rate, the server adjusts the difficulty (e.g., increasing the number of required zeroes during high load).
   - This keeps the server more resilient when attacked by many concurrent requests.

### 3.3 Why This Approach?
- A simple SHA-256 based puzzle is easy to implement and widely understood.
- The difficulty can be adjusted quickly at runtime.
- Minimal overhead for the server (the main work is offloaded to the client).
- The approach leverages well-known cryptographic hash functions, avoiding the need for more complex or exotic techniques.
