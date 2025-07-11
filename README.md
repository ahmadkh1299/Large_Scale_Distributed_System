# Large-Scale Distributed Systems Project

> ⚙️ A scalable, language-agnostic microservices platform built with Go, Python, and Java — featuring service discovery, distributed caching, and cross-language interoperability via MetaFFI.

---

## 🚀 Overview

This project implements a large-scale distributed system with dynamic service discovery, load balancing, distributed caching using Chord DHT, and seamless cross-language integration.

It was developed as part of the Tel Aviv University Distributed Systems Workshop and was **nominated for the "Outstanding Project" (פרויקט מצטיין) competition**.

---

## 🧩 Architecture Highlights

- **Service Discovery & Registry**
  - Built with Go, uses a distributed hash table (DHT) in Java to register services dynamically.
  - Supports multiple registry instances for fault tolerance.

- **Distributed Cache (CacheService)**
  - Java-based implementation of a **Chord DHT ring**.
  - Supports key-value replication for fault tolerance and fast lookup.

- **CrawlerService**
  - Written in Python using BeautifulSoup for web crawling.
  - Exposed via a Go interface using MetaFFI to integrate with the platform.


- **Load Balancer**
  - Routes client requests to appropriate service instances using round-robin logic.

---

## 🔗 Technologies Used

- Go (Golang)
- Python (BeautifulSoup, Requests)
- Java (Chord DHT)
- MetaFFI (Cross-language bridge)
- Docker & gRPC


---

## 🛠️ Setup & Run

> Requires MetaFFI and language-specific runtimes.

1. Install Go, Java, Python 3.
2. Install MetaFFI CLI and compiler.
3. Run each service via provided scripts or Docker Compose.

---

## 🏁 Future Improvements

- Auto-scaling orchestrator based on system load  
- Authentication and access control for secure deployment  
- gRPC-based communication for performance

---

## 📣 Credits

Developed by Ahmad Khalaila, Deeb Tibi, Obaida Haj Yahya students at Tel Aviv University as part of the Distributed Systems Workshop (2025).  
Special thanks to our project mentor Tsvi Cherny-Shahar and faculty support.

---


