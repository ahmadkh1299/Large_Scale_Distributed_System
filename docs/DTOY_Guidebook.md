# DTOY: A Guide to Building Polyglot Distributed Systems

### Architecture, Concepts, and Code Walkthrough

*A complete guidebook to the Tel Aviv University Distributed Systems Workshop project — from first principles to file-by-file implementation.*

**Built with Go, Java, and Python** · gRPC · Protocol Buffers · ZeroMQ · Chord DHT · MetaFFI

*Developed by Ahmad Khalaila, Deeb Tibi, and Obaida Haj Yahya — nominated for the Outstanding Project competition.*

---

# Table of Contents

**Part I: The Blueprint (Concepts & Theory)**
1. Why Distributed Systems Exist
2. The Language Services Speak: Protobuf & gRPC
3. Welcome to DTOY City: The Three Services
4. The Polyglot Trick: Three Languages, One Process

**Part II: System Architecture & Design Patterns**
5. Architectural Style
6. Key Design Patterns
7. System Data Flow: Startup, a Write Request, and the Health Loop

**Part III: The Codebase Deep Dive**
8. The Directory Map
9. Root & Shared Foundations
10. The Framework Layer: `services/common/`
11. The Registry Service
12. The Cache Service
13. The Test Service
14. The Interop Layer: Hand-Written FFI
15. Black-Box Testing
16. Build & Deployment: `output/`

**Part IV: Performance Bottlenecks & Refactoring**
17. Performance & Scalability
18. Correctness & Concurrency
19. Design Hygiene & Maintainability
20. Security & Operational Concerns
21. The First Four Refactoring Tickets

**Part V: The Builder's Roadmap**
22. The Five-Phase Roadmap
23. Prerequisites, Tools & the "Google This" Cheat Sheet

---

---

# Part I: The Blueprint (Concepts & Theory)

*This part assumes nothing beyond basic programming: variables, loops, and functions. Every technical term is defined the moment it appears. If you already know what a microservice is, you can skim — but the analogies introduced here (the kitchen, the phonebook, the circle of friends) are used throughout the rest of the book.*

## Chapter 1: Why Distributed Systems Exist

### The single chef problem

Imagine a tiny restaurant with **one chef** who does everything: takes your order, cooks the food, washes the dishes, restocks the fridge, and handles the cash register.

This works fine when there are 3 customers. Now imagine 300 customers show up. The chef doesn't get 100× faster just because you yell at them. And here's the scary part: if the chef gets sick, **the entire restaurant shuts down**. One person, one point of failure.

A real restaurant solves this with a **kitchen staff**: a host who greets people, several line cooks, a dishwasher, a cashier. Each person has *one job* and does it well. If one line cook goes home sick, the kitchen slows down a little but keeps running. If Friday night gets busy, you hire *more line cooks* — you don't try to find one superhuman chef.

**A distributed system is exactly this: instead of one giant program doing everything on one computer, you run many small programs (often on many computers) that each do one job and talk to each other.**

### Monolith vs. microservices

Two words you'll hear constantly:

- A **monolith** is the single-chef approach: one big program that does everything. It's honestly the right way to *start* most projects — it's simpler!
- **Microservices** are the kitchen-staff approach: many small programs (called **services** — a "service" is just a program that runs forever, waiting to answer requests). Each service does one job.

Why do people split things up?

1. **Scaling.** If the "fetch web pages" part of your app is slow, you can run 10 copies of *just that part* instead of 10 copies of everything.
2. **Fault tolerance** (fancy words for "surviving failures"). If one copy crashes, the others keep serving customers.
3. **Team independence.** One teammate can rebuild the cache without touching the crawler.

But splitting things up creates a brand new problem, and this problem is basically what the whole DTOY project is about:

> **If your program is now 9 separate programs, how do they find each other and talk to each other?**

Hold that question. It is the plot of this entire book.

## Chapter 2: The Language Services Speak: Protobuf & gRPC

### The problem: computers need to agree on a format

When two programs talk over a network, they can't send Go variables or Python objects directly — networks only carry raw bytes (just 1s and 0s). So both sides need to agree, *precisely*, on what those bytes mean. Converting structured data into bytes is called **serialization** (and reading it back is **deserialization**). Don't let the words scare you — they just mean "packing a suitcase" and "unpacking a suitcase."

### Protobuf: the strict paper form

**Protocol Buffers ("Protobuf")** is a way to define the *exact shape* of the data two programs exchange. Think of it like an official government form:

- Line 1: Name (must be text)
- Line 2: Address (must be text)

No scribbling in the margins, no leaving fields blank, no writing your name where the address goes. You describe the form once in a little file ending in `.proto`, like this real one from the project:

```proto
message StoreKeyValue {
    string key = 1;
    string value = 2;
}
```

That says: "A `StoreKeyValue` message has exactly two fields, both text, in this order." A tool then **auto-generates code** in Go (or Python, or Java...) that packs and unpacks this form for you, extremely efficiently. You never hand-write the packing logic.

Why not just send plain text like JSON (the loose, human-readable format most websites use)? You could! But Protobuf forms are:

- **Tiny** — binary, no wasted characters, faster to send.
- **Strict** — if you send the wrong shape, it fails *immediately* instead of causing a weird bug three files away.
- **Type-checked** — the compiler catches your mistakes before you even run the program.

### gRPC: a direct phone call between programs

Protobuf is *what* we say. **gRPC** is *how* we say it.

You know how websites work: your browser sends an HTTP request like "GET me the page at /menu" and gets back a document. That's like **mailing a letter to a building** and hoping someone inside figures out what you want.

gRPC is different. It's like having a **direct phone line to a specific person**, with an agreed-upon script. RPC stands for **Remote Procedure Call**, which is a fancy way of saying:

> "I want to call a function... that lives in a *different program*, possibly on a *different computer*... and have it feel exactly like calling a normal local function."

In DTOY, there's a service with a function `HelloToUser(name)`. A completely different program, on a different port, just... calls it:

```go
answer := client.HelloToUser("Ahmad")   // runs on ANOTHER program!
// answer == "Hello Ahmad"
```

Behind the scenes, gRPC packs the argument into a Protobuf form, sends it over the network, the server unpacks it, runs the real function, packs the answer, and sends it back. To you, it looks like one line of code. That's the magic.

The menu of available functions is also declared in the `.proto` file — this is called the **contract**:

```proto
service TestService {
    rpc HelloToUser(StringValue) returns (StringValue);
    rpc Store(StoreKeyValue) returns (Empty);
    rpc Get(StringValue) returns (StringValue);
}
```

Read that like a restaurant menu: "here are the dishes you can order, here's what info I need from you, here's what you'll get back." Both sides generate their code from this one file, so they *literally cannot* misunderstand each other.

## Chapter 3: Welcome to DTOY City: The Three Services

The system is called **DTOY** (the team's initials). It's written mostly in **Go** — a language Google made that's fantastic for network programs: clean, fast, and beginner-friendlier than C++.

Here's a fun quirk: the project compiles into **one single program**, but when you launch it, you hand it a small config file that says *which role to play* — like one actor who can play three characters depending on the costume you give them. The startup script launches **9 copies** of this program: 3 registries, 3 test services, 3 caches. Instant city.

### 🏢 The Registry Service — the city phonebook

**The problem it solves:** services start up on random addresses (the operating system literally picks a free port). So how does anyone find anyone?

**The answer:** a phonebook. When a service starts, it calls the Registry and says:

> "Hi, I'm a `TestService`, and I live at address `127.0.0.1:53211`. Put me in the book."

That's called **registering**. When someone needs a TestService, they ask the Registry:

> "Give me the addresses of everyone named `TestService`."

That's called **discovery**. The registry answers with a list, and the caller picks one — DTOY picks randomly, which spreads the work around. Spreading requests across copies is called **load balancing**.

Two extra-cool things this Registry does:

1. **It's not alone.** Three registry copies run at once, sharing one phonebook between them (using the same "circle of friends" trick the Cache uses — next section). If one registry crashes, the other two still have the book.
2. **It checks pulses.** Every 10 seconds, the registry calls every service in the book and asks `IsAlive()` — literally a function that just returns `true`. If a service fails to answer **twice in a row**, the registry assumes it's dead and erases it from the book. The phonebook *heals itself*: nobody ever gets sent to a disconnected number.

### 🧠 The Cache Service — the city's shared memory

**The problem it solves:** services need to store and share data — a giant shared dictionary: `Set("username", "Ahmad")`, `Get("username")` → `"Ahmad"`. This is called a **key-value store** (you already know this concept — it's a dictionary/map, just living on the network).

But what if the dictionary is too big for one computer? Or the computer holding it dies?

**The answer: split the dictionary across many computers.** This is where the project's favorite idea lives — the **Chord Ring**.

**The circle-of-friends analogy.** Imagine you and 9 friends need to memorize a 1,000-page phone directory. Nobody can memorize the whole thing. So you stand in a circle and split it: you take names A–B, your friend takes C–E, the next takes F–H, around the circle. Now the clever part — there's a simple *math rule* (a **hash function**: a formula that turns any name into a number) that tells anyone, instantly, *which friend* holds any given name. Ask anyone in the circle for "Menachem," they compute the rule and pass your question around the ring toward the right person. No boss, no central index. And if a friend leaves the circle, the neighbors absorb their pages and the rule updates.

This structure is called a **DHT** — a **Distributed Hash Table** — and "Chord" is a famous recipe for building one. Real systems like BitTorrent are built on ideas like this.

In DTOY, every Cache Service node joins one shared Chord ring. When you `Set("Deeb", "Tibi")`, the key gets hashed and stored on whichever node the ring says owns it. Any cache node can find it afterward. One quirky detail: the first cache node to start becomes the **root** ("hey, no ring exists yet — I'll start one"), and every later node discovers the root through the Registry and joins its ring.

### 🍽️ The Test Service — the actual restaurant

Registry and Cache are *infrastructure* — plumbing. The **Test Service** is the application people actually use. Its menu (straight from its `.proto` contract):

- `HelloWorld()` / `HelloToUser(name)` — the classic warm-up.
- `Store(key, value)` / `Get(key)` — pretends to store your data, but secretly delegates to the Cache Service. A service using another service — that's the whole microservices dance!
- `ExtractLinksFromURL(url, depth)` — give it a web page, and it returns every link on that page. This is a **web crawler** — the same core idea Google's search engine is built on. DTOY's crawler is written in *Python* (that plot twist is Chapter 4).
- `WaitAndRand(seconds)` — waits, then sends back a random number. Sounds silly, but it exists to demonstrate **streaming**: answers that arrive over time instead of all at once.

Three copies of it run at once, all registered in the phonebook, so clients get spread across them.

*(There's also a bonus second communication channel where clients can drop off a request like a note in a mailbox and pick up the answer later — "async messaging" over a library called ZeroMQ. Part II covers it properly.)*

## Chapter 4: The Polyglot Trick: Three Languages, One Process

Here's the flex of this project. The system uses:

- **Go** — all the services, networking, and glue.
- **Java** — the Chord ring implementation (the circle of friends).
- **Python** — the web crawler, using a beloved library called BeautifulSoup that makes reading web pages easy.

The obvious way to mix languages is to make each one a separate service and have them talk over the network. But network calls are *slow* (milliseconds are an eternity to a CPU), and each extra service is another thing to babysit.

The team did something sneakier. Using a tool called **MetaFFI** (FFI = **Foreign Function Interface** — "let language A call functions written in language B"), the project loads an actual **Java virtual machine** and an actual **Python interpreter *inside* the running Go program**. When Go "calls Java," there's no network involved at all — it's a direct in-memory function call, thousands of times faster.

The kitchen analogy: instead of phoning the bakery across town every time you need bread (a network call), you hired a baker to work *inside your own kitchen* (FFI). Same skills, zero delivery time.

Why is this a neat trick for a university project? Because it forces you to understand what *actually* happens beneath the pretty language syntax — memory ownership, type conversion, "who frees this string?" The team even wrote a raw, manual version of these bridges in C first (hundreds of lines of careful glue code — preserved in the `interop/` folder and dissected in Chapter 14) before adopting MetaFFI, precisely to appreciate what it automates. That kind of depth is what turns a course project into a portfolio piece.

---

---

# Part II: System Architecture & Design Patterns

*You now have the concepts: services, contracts, discovery, rings, FFI. This part puts on the architect's hat and describes the same system formally — how the pieces are shaped, which named design patterns they implement, and exactly what happens on the wire. The analogies from Part I map one-to-one onto the terms used here: the phonebook is "service discovery," the circle of friends is "the Chord DHT," and the direct phone call is "synchronous gRPC."*

## Chapter 5: Architectural Style

DTOY is a **microservices architecture delivered as a single polyglot binary** — an interesting hybrid:

- **One Go executable** (`main.go`) acts as a *service launcher*: the YAML config passed on the command line decides which of three services the process becomes (`RegistryService`, `TestService`, or `CacheService`). Deployment-wise you get microservices — many processes, each independently addressable and horizontally scalable. Build-wise you get a monorepo/monolith with shared internal packages.
- **Service-oriented topology** with three roles:
  - **RegistryService** — service discovery plus health checking (the *control plane*). Runs as a cluster of 3; registry instances share state through their own **Chord DHT ring** (Java, port 1098).
  - **CacheService** — distributed key-value store (the *data plane*). Each cache node joins a second, separate **Chord ring** (port 1097), so data is partitioned across cache nodes.
  - **TestService** — the "application" service: hello-world operations, key-value operations delegated to the cache, and a web crawler implemented in Python.
- **Polyglot via in-process FFI, not network calls.** The Chord DHT is Java (`dht/Chord.class`) and the crawler is Python (`crawler.py`), both embedded into the Go process through **MetaFFI** — the JVM and CPython interpreters are loaded *inside* the Go process. The `interop/` folder contains a parallel, hand-written **cgo/JNI + CPython C-API** implementation of the same bridges (build-tagged `//go:build interop`) — an educational artifact showing what MetaFFI abstracts away.
- **Two transport planes:**
  - **Synchronous:** gRPC/Protobuf for request-response.
  - **Asynchronous:** ZeroMQ REQ/REP carrying a hand-rolled protobuf envelope (`CallParameters` / `ReturnValue`), giving clients *future-based* async calls. Each service can expose both; the MQ endpoint registers itself under `<ServiceName>MQ`.

## Chapter 6: Key Design Patterns

| Pattern | Where | Notes |
|---|---|---|
| **Service Registry / Discovery** | `registry-service`, `ServiceClientBase.Connect()` | Clients never hardcode service addresses; they ask the registry, which stores `name → "addr1;addr2;…"` in the DHT. |
| **Client-side load balancing** | `PickNode()` in `ServiceClientBase.go` | Random selection per call. (The README says "round-robin" — the code is actually random; only the ZMQ REQ socket does true round-robin.) |
| **Service / Servant / Client / Common layering** | every service | A CORBA/ICE-style **Servant pattern**: `service/` is the transport adapter (gRPC handlers), `servant/` is transport-agnostic business logic, `client/` is a typed SDK, `common/` holds the contract (`.proto` + generated stubs). Clean separation of transport from domain logic. |
| **Template Method / Inversion of Control** | `services/common/ServiceBase.go` → `Start()` | Generic lifecycle (listen → bind gRPC → bind MQ → register → serve); each service injects `bindgRPCToService` and `messageHandler` hooks. |
| **Generic base class** | `ServiceClientBase[client_t]` | Go generics reuse discovery + connection logic across all typed clients. |
| **Factory Method** | `CreateClient func(grpc.ClientConnInterface) client_t` fields; `NewXServiceClient()` constructors | The generated gRPC stub constructor is injected as a factory. |
| **Future / Promise** | `*Async` methods in `TestServiceClient.go` | Async calls return a `func() (T, error)` that blocks on the ZMQ reply when invoked. |
| **Command / dynamic-dispatch envelope** | `CallParameters{method, data}` + `messageHandler` switch | A mini-RPC protocol over ZMQ: method name + opaque serialized payload. |
| **Adapter / Bridge (FFI)** | `dht/Chord.go`, `interop/ChordDHT.go`, `interop/Crawler.go` | Go structs wrap opaque JVM/CPython handles behind idiomatic Go methods. |
| **Singleton (package-level)** | `chord`, `serviceInstance`, `utils.Logger`, MetaFFI runtimes in `init()` | Heavy use of package-global state initialized in `init()` functions. |

## Chapter 7: System Data Flow

### Startup choreography

1. **Registries.** Instance 1 binds port 8502 and **creates** the registry Chord ring (JVM port 1098) and becomes the health-checker; instances 2–3 fail to bind 8502, retry 8503/8504, and **join** the ring. All three now share one registry keyspace via the DHT. (The OS's "port already taken" error *is* the coordination mechanism — no leader election needed.)
2. **TestServices.** Each binds an ephemeral gRPC port and an ephemeral ZMQ REP port, then registers `TestServiceMQ` and `TestService` with a random registry node.
3. **CacheServices.** Each registers, then bootstraps the *data* ring: discover peers via the registry, probe each with `IsRoot()`, and join the cache Chord ring (JVM port 1097) — or create it and become root if none exists.

### The critical write path — the life of `Store("k", "v")`

This trace scales the Chapter 3 analogy into the exact call chain:

```
 YOU (client program)
  │ 1. NewTestServiceClient() reads ./configurations/RegistryAddresses.yaml
  │ 2. Store() → ServiceClientBase.Connect()
  ▼
 REGISTRY (picked randomly from the configured 3, e.g. :8503)
  │    gRPC Discover("TestService")
  │    → RegistryServiceServant looks up the shared phonebook:
  │      GetAllKeys() + Get("TestService") ──MetaFFI/JVM──► Registry Chord ring (:1098)
  │    → returns "addr1;addr2;addr3", decoded to a node list
  │
  │ 3. PickNode() → random TestService instance → grpc.Dial
  ▼
 TEST SERVICE (one of 3 copies)
  │    gRPC handler Store() → TestServiceServant.Store
  │    "I don't store data myself — the Cache does that."
  │ 4. NewCacheServiceClient(LoadRegistryAddresses())  ← re-reads YAML from disk
  │ 5. Discover("CacheService") — a full registry round-trip AGAIN
  │ 6. PickNode() → random CacheService instance → grpc.Dial
  ▼
 CACHE SERVICE
  │    gRPC handler Set() → CacheServiceServant.Set
  │    chord.Set("k","v") ──MetaFFI/JVM──► Cache Chord ring (:1097)
  │    (Java Chord hashes the key and routes it to the owning
  │     cache node, replicating per its implementation)
  │
  │ 7. "Done!" travels all the way back up the chain
  ▼
 YOU: Store() returns. Total time: a few milliseconds.
```

One logical `Store` therefore costs: **2 registry discoveries + 3 gRPC dials + 2 JVM FFI transitions + intra-ring Chord routing** — all torn down afterward. (Part IV has opinions about this.)

The beautiful part: call `Get("k")` immediately afterward and land on a *completely different* TestService copy, which asks a *different* cache node — **you still get your value back**, because the Chord ring routes the lookup to whichever node holds the key. Nine independent programs behaving like one brain. Kill a service mid-run, and within ~20 seconds the health-checker erases it from the phonebook and traffic flows around the hole.

### The asynchronous path — `ExtractLinksFromURLAsync(url, 1)`

1. Client `ConnectMQ()`: Discover `TestServiceMQ` → one ZMQ **REQ** socket connected to *all* MQ endpoints (ZMQ round-robins requests among them).
2. `NewMarshaledCallParameter("ExtractLinksFromURL", params)` → protobuf envelope → `SendBytes` (fire and return immediately).
3. Server REP loop unmarshals the envelope; `messageHandler` switches on the method name and calls the same handler used by gRPC → servant → **MetaFFI → embedded CPython** runs `crawler.extract_links_from_url` → result marshaled into a `ReturnValue`.
4. The client later invokes the returned closure (the **future**): `RecvBytes` blocks for the reply, decodes the envelope, and surfaces either `rv.Error` or the typed inner message.

### The health/liveness control loop

Every 10 seconds the base registry walks all DHT keys, pings every registered address (`IsAlive` via gRPC, or a ZMQ connect for `*MQ` entries), counts consecutive failures in `isAliveCheck`, and after 2 strikes rewrites the DHT entry without that node — self-healing discovery.

### Communication summary

| Channel | Technology | Used for |
|---|---|---|
| Client ↔ Registry | gRPC (sync) | Register / Unregister / Discover |
| Client ↔ TestService / CacheService | gRPC (sync + one server-stream) | All service RPCs |
| Client ↔ TestService | ZeroMQ REQ/REP + protobuf envelope | Async futures |
| Registry ↔ Registry | Java Chord ring (:1098) via in-process JVM | Shared registry state |
| Cache ↔ Cache | Java Chord ring (:1097) via in-process JVM | Partitioned KV data |
| Go ↔ Java / Python | MetaFFI (in-process FFI) | DHT operations, crawler |

---

---

# Part III: The Codebase Deep Dive

*This part walks every file in the repository, organized by directory, explaining what the code actually does — function by function. Keep Part II's vocabulary in mind: each service is a four-layer sandwich (contract → transport → servant → client), and everything stands on the shared framework in `services/common/`.*

## Chapter 8: The Directory Map

```
Large_Scale_Distributed_System/
├── main.go                        # Entry point: config-driven service launcher
├── go.mod / go.sum                # Module: github.com/TAULargeScaleWorkshop/DTOY
├── config/                        # Shared YAML config schemas (structs only)
│   ├── ConfigBase.go              #   type + registry_addresses
│   ├── RegistryConfigBase.go      #   type + listenPort (registry only)
│   └── RegistryClientConfig.go    #   registry_addresses (client side)
├── utils/
│   └── Logger.go                  # Global stdout logger singleton
│
├── services/                      # ★ The heart of the system
│   ├── common/                    # Cross-service infrastructure ("framework" layer)
│   │   ├── ServiceBase.go         #   Server lifecycle: gRPC + ZMQ + registration
│   │   ├── ServiceClientBase.go   #   Generic client: discovery + LB + connect
│   │   ├── ServiceClientBaseDirect.go # Address-direct client (health checks, IsRoot)
│   │   ├── CallMessage.proto/.pb.go   # MQ envelope contract
│   │   └── CallMessageFactory.go  #   Marshal/unmarshal helpers for the envelope
│   │
│   ├── registry-service/          # Control plane: discovery + liveness
│   │   ├── common/                #   RegistryService.proto + generated stubs
│   │   ├── service/               #   gRPC transport layer + cluster bootstrap
│   │   ├── servant/               #   Register/Unregister/Discover logic + health loop
│   │   │   └── dht/               #   MetaFFI Go wrapper over Java Chord + .class files
│   │   └── client/                #   Typed registry client (used by everyone)
│   │
│   ├── cache-service/             # Data plane: distributed KV store
│   │   ├── common/                #   CacheService.proto + stubs
│   │   ├── service/               #   gRPC handlers (Get/Set/Delete/IsAlive/IsRoot)
│   │   ├── servant/               #   Chord-ring membership + KV ops
│   │   └── client/                #   Typed cache client
│   │
│   └── test-service/              # Application service (workload)
│       ├── common/                #   TestService.proto + stubs
│       ├── service/               #   gRPC handlers + MQ messageHandler dispatch
│       ├── servant/               #   Business logic + Python crawler via MetaFFI
│       │   └── crawler.py         #   BeautifulSoup link extractor
│       └── client/                #   Typed client incl. *Async (ZMQ futures)
│
├── interop/                       # Educational: raw cgo JNI / CPython bridges
│   ├── ChordDHT.go                #   Hand-written JNI bridge (build tag: interop)
│   ├── Crawler.go                 #   Hand-written CPython bridge (build tag: interop)
│   └── crawler.py, Interop_test.go, backup/*.class
│
├── testing/                       # Black-box integration tests (run against live cluster)
│   ├── testservice-testing/       #   End-to-end TestService scenarios + config
│   └── cache-testing/             #   End-to-end CacheService scenarios + config
│
└── output/                        # Deployment artifacts & runtime working dir
    ├── build.sh                   #   go build → ./output/large-scale-workshop
    ├── start.sh                   #   Boots 3× registry, 3× test, 3× cache
    ├── RunRegistry/TestService/Cache.sh  # Per-service launchers (set CLASSPATH)
    ├── configurations/*.yaml      #   Per-service runtime configs
    └── dht/*.class                #   Compiled Java Chord classes (runtime dependency)
```

**Architectural responsibility, one line each:**

- `config/` — configuration *schema* (deserialization targets), no logic.
- `services/common/` — the in-house "microservice framework": server bootstrap, service registration, discovery-aware clients, and the async MQ protocol. Everything else stands on this.
- `*/common/` — the service *contract* (proto + generated code); the only thing a consumer should need.
- `*/service/` — transport adapters: unwrap protobuf, call servant, wrap result.
- `*/servant/` — pure(ish) business logic, unaware of gRPC.
- `*/client/` — consumer SDKs that hide discovery, load balancing, and transports.
- `interop/` — a from-scratch FFI layer kept for reference; excluded from normal builds.
- `output/` — the deployable unit; scripts assume it as the working directory (relative paths like `./dht/Chord.class` and `./configurations/*.yaml` resolve here).

## Chapter 9: Root & Shared Foundations

### `main.go`

The single entry point for the entire system — all 9 processes in the cluster run this same binary.

- Checks `len(os.Args) != 2`: the program demands exactly one argument, a path to a YAML config file.
- `os.ReadFile(configFile)` reads the raw bytes; `yaml.Unmarshal(configData, &config)` parses them into a `config.ConfigBase` struct. At this stage it only cares about *one* field: `Type`.
- The `switch config.Type` is the "costume selector": depending on whether the YAML says `TestService`, `RegistryService`, or `CacheService`, it calls that package's `Start(configData)` — passing the **raw YAML bytes** along so each service can re-parse it into its own, richer config struct.
- Any unknown type exits with code 4. Each failure mode has a distinct exit code (1 = bad args, 2 = unreadable file, 3 = bad YAML), which is handy in shell scripts.

### `go.mod` / `go.sum`

The Go module manifest. Declares the module name `github.com/TAULargeScaleWorkshop/DTOY` (which is why every import in the project starts with that prefix), pins Go 1.22.2, and lists the dependencies:

- `google.golang.org/grpc` + `protobuf` — the RPC stack.
- `github.com/MetaFFI/*` (three packages) — the cross-language bridge to Java and Python.
- `github.com/pebbe/zmq4` — Go bindings for ZeroMQ, used for the async plane.
- `gopkg.in/yaml.v2` — config parsing.
- `go.starlark.net` and `routine` arrive transitively via MetaFFI.

`go.sum` is the lockfile of cryptographic checksums — never edited by hand.

### `README.md`

Project overview: what it is, the tech list, and credits. Two spots drift from the code: it claims round-robin load balancing (the code picks randomly), and lists "gRPC-based communication" as a *future* improvement (gRPC is already the main transport).

### `.gitignore` / `.gitattributes`

`.gitignore` contains just `__pycache__/` (Python bytecode junk). `.gitattributes` forces the five `.sh` scripts to keep Unix line endings (`eol=lf`) — a shell script with Windows `\r\n` endings dies with cryptic errors inside the Linux dev container.

### `large-scale-workshop.code-workspace`

A VS Code workspace file: opens the repo root and tunes the Java language server's JVM memory flags (`-Xmx2G`, etc.) — needed because the repo contains Java `.class` files and the workshop machines were memory-constrained.

### `.devcontainer/devcontainer.json`

Tells VS Code to develop inside the Docker image `tscs/large-scale-workshop:latest` — a pre-baked environment from the course staff with Go, JDK, Python 3.11, ZeroMQ, and MetaFFI already installed. That's why the code can assume paths like `/usr/lib/jvm/java-11-openjdk-amd64` exist.

### `config/` — configuration schemas

Three tiny files, each defining a struct that YAML gets parsed into (the `yaml:"..."` struct tags map YAML keys to Go fields):

- **`ConfigBase.go`** — `{Type, RegistryAddresses}`: the common denominator of every service config. `main.go` uses it to route; TestService/CacheService use it to find the phonebook.
- **`RegistryConfigBase.go`** — `{Type, ListenPort}`: the registry's own config. The registry is the only service with a *fixed* port (8502), because everyone must know where the phonebook is — there's no phonebook for the phonebook.
- **`RegistryClientConfig.go`** — `RegistryServiceConfig{RegistryAddresses}`: the client-side variant for programs that only *consume* services (like the test suites).

### `utils/Logger.go`

A 15-line global logger. `LoggerWrapper` embeds `log.Logger` (Go embedding means the wrapper automatically inherits `Printf`, `Fatalf`, etc.). The package-level `init()` — a special Go function that runs automatically when the package is first imported — configures it to write to stdout with date, time, and `file:line` prefixes. Every other file just calls `utils.Logger.Printf(...)`. A classic singleton.

## Chapter 10: The Framework Layer: `services/common/`

The most important directory in the repo: the homegrown microservice toolkit all three services stand on.

### `ServiceBase.go` — the server-side backbone

1. **`LoadRegistryAddresses()`** — re-reads the YAML file from `os.Args[1]` (the same file `main.go` read) and returns the `registry_addresses` list, fataling if missing or empty.

2. **`startgRPC(listenPort)`** — the low-level listen step:
   - `net.Listen("tcp", ":<port>")` opens a TCP socket. Passing port `0` makes the OS pick any free port — that's how TestService/CacheService instances avoid colliding.
   - It extracts the *actual* port via `lis.Addr().(*net.TCPAddr).Port` (a type assertion, needed because port 0 was resolved by the OS).
   - Crucially, it does **not** start serving. It returns a `startListening` closure wrapping `grpcServer.Serve(lis)`. This delayed-start trick lets the caller register RPC handlers and register with the registry *before* traffic flows.

3. **`Start(serviceName, port, bindgRPCToService, messageHandler)`** — the template method orchestrating a service's whole lifecycle:
   - Starts gRPC (deferred).
   - If `messageHandler != nil`, calls `bindMQToService` to open a ZeroMQ socket, then registers that MQ endpoint under `serviceName + "MQ"` (e.g. `TestServiceMQ`) — so the async endpoint is discoverable as if it were its own service.
   - Calls `bindgRPCToService(grpcServer)` — the hook where the concrete service attaches its generated handlers.
   - Spawns a goroutine that runs the MQ receive loop forever, with a deferred unregister.
   - Returns three things: `startListening`, the resolved port, and a `registerService` closure that performs the gRPC-side registration and hands back an `unregister` function. Registration is a *closure you invoke later*, so the caller controls ordering.

4. **`bindMQToService(port, messageHandler)`** — the async server:
   - Creates a ZeroMQ **REP** (reply) socket and binds to `tcp://127.0.0.1:*` (`*` = OS-assigned port; `GetLastEndpoint()` retrieves the actual address).
   - Returns a `startMQ` closure containing an infinite loop: `RecvBytes` blocks until a request arrives; then a goroutine unmarshals the bytes into a `CallParameters` protobuf (`{method, data}`), calls the injected `messageHandler(method, data)`, wraps the result (or error string) in a `ReturnValue`, marshals, and `SendBytes` it back.
   - Design wrinkle (detailed in Part IV): replying from a goroutine violates REP's strict receive→send alternation, and several error paths return *without* replying — which strands the client.

5. **`registerAddress(name, registries, address)`** — builds a `RegistryServiceClient`, calls `Register`, and returns an `unregister` closure capturing the same name/address. A neat use of closures as "undo tokens."

### `ServiceClientBase.go` — the client-side backbone

A **generic** struct (Go 1.18+ generics):

```go
type ServiceClientBase[client_t any] struct {
    RegistryAddresses []string
    ServiceName       string
    CreateClient      func(grpc.ClientConnInterface) client_t
}
```

`client_t` is the *generated gRPC stub type* (e.g. `TestServiceClient`), and `CreateClient` is the generated constructor injected as a factory. This one struct gives every typed client four capabilities:

- **`LoadRegistryAddresses()`** — client-side variant: reads the hardcoded relative path `./configurations/RegistryAddresses.yaml` (so tests must run from a directory containing that folder).
- **`Connect()`** — the discovery dance: build a registry client → `Discover(ServiceName)` → `PickNode` → `grpc.Dial(addr, WithInsecure(), WithBlock())` (insecure = no TLS; block = wait until the TCP connection is actually up) → wrap the connection with the injected factory. Returns the typed stub, a `closeFunc`, and an error.
- **`PickNode(services)`** — `rand.Intn(len(services))`: pick one instance at random. This one line is the system's load balancer.
- **`ConnectMQ()` / `getMQNodes()`** — the async variant: discovers `<ServiceName>MQ` endpoints and connects **one REQ socket to all of them**. That's idiomatic ZeroMQ — a REQ socket connected to multiple endpoints automatically round-robins requests among them.

### `ServiceClientBaseDirect.go` — the "call this exact address" client

Used by the registry's health checker and the cache bootstrap, where discovery would be circular:

- **`IsMessageQueueService(name)`** — checks for the `"MQ"` suffix.
- **`IsAlive(serviceName)`** — two branches:
  - *MQ services*: creates a ZMQ socket and `Connect`s to the address, returning `true` if that didn't error. (Weak check — ZMQ connect is asynchronous and succeeds even if nothing is listening.)
  - *gRPC services*: the clever/hacky part. It doesn't have the generated stub for an arbitrary service, so it **builds the gRPC method path by hand**: `fmt.Sprintf("/%s.%s/IsAlive", strings.ToLower(serviceName), serviceName)` yields e.g. `/testservice.TestService/IsAlive`, then uses low-level `conn.Invoke(...)` with an `Empty` request and a `BoolValue` response. It works because every service's proto follows the same naming convention and declares an `IsAlive` RPC.
- **`IsRoot()`** — same raw-invoke trick with the hardcoded path `/cacheservice.CacheService/IsRoot`; used only during cache-ring bootstrap.

### `CallMessage.proto` → `CallMessage.pb.go`

The envelope protocol for the MQ plane. Two messages: `CallParameters{method string, data bytes}` ("call this method with these serialized args") and `ReturnValue{data bytes, error string}` ("here's the serialized result, or an error message"). The `.pb.go` file (226 lines) is 100% machine-generated by `protoc` — struct definitions, getters, and serialization metadata. Generated files are never edited; they are regenerated from the `.proto`.

### `CallMessageFactory.go`

Hand-written codec helpers around the envelope:

- **`NewMarshaledCallParameter(method, kwargs...)`** — marshals each argument protobuf to bytes, concatenates them into one byte slice, wraps in `CallParameters`, and marshals the whole envelope. (Concatenation works here only because every call passes exactly one message.)
- **`UnmarshalReturnValue(bytes)`** — bytes → `ReturnValue`.
- **`(rv) ExtractInnerMessage(p)`** — a method on `ReturnValue` that unmarshals its inner `Data` bytes into whatever typed message you hand it. The "unpack the envelope, then unpack the letter inside" step.
- **`ParseParamsIntoBytes`** — the marshaling half of #1, unused elsewhere.

### `RegistryAddresses.yaml`

A sample client config listing registries `127.0.0.1:8502/8503` — the shape `ServiceClientBase.LoadRegistryAddresses` expects.

## Chapter 11: The Registry Service

### `init.go`

An empty `func init() {}` in package `registryservice` — scaffolding from the workshop template. (Same for the other two services' `init.go` files.)

### `common/` — the contract

- **`RegistryService.proto`** — declares `ServiceRequest{name, address}`, `ServiceNodes{repeated string nodes}`, and three RPCs: `Register`, `Unregister` (both take a `ServiceRequest`, return `Empty`), and `Discover` (takes a `StringValue`, returns `ServiceNodes`). It imports Google's "well-known types" (`wrappers.proto`, `empty.proto`) so simple parameters don't need custom messages.
- **`RegistryService.pb.go`** (244 lines, generated) — Go structs for the two messages plus getters and serialization tables.
- **`RegistryService_grpc.pb.go`** (185 lines, generated) — the client stub (`RegistryServiceClient` interface + `NewRegistryServiceClient(conn)`), the server interface (`RegistryServiceServer`), `UnimplementedRegistryServiceServer` (a struct you embed so future proto additions don't break compilation), and `RegisterRegistryServiceServer(grpcServer, impl)` which wires an implementation into the server's routing table.

### `service/RegistryService.go` — transport + cluster bootstrap

- `registryServiceImplementation` embeds `UnimplementedRegistryServiceServer` and implements the three RPCs — each handler is three lines: unwrap the protobuf, call the servant, wrap the reply. (Bug noted in Part IV: `Register`/`Unregister` return success even if the servant errored.)
- **`Start(configData)`** contains the port-probing loop that makes "run the same binary 3 times" work:

  ```go
  for {
      listenPort := baseListenPort + i
      err := StartServer(baseListenPort, listenPort, ...)
      if err == nil { break } else { i++ }
  }
  ```

  Instance 1 grabs 8502; instance 2 fails to bind 8502 (address in use) and succeeds on 8503; instance 3 lands on 8504.
- **`StartServer`** — binds gRPC, then `CreateChord(basePort, myPort)`: whoever holds the base port *creates* the registry's DHT ring; everyone else *joins* it. Only the base-port node launches `CheckIsAliveEvery10Seconds()` in a goroutine; then `startListening()` blocks forever.
- It has its own local `startgRPC` (a near-duplicate of the one in `ServiceBase.go`) because the registry can't use `services.Start` — that function registers with the registry, and the registry can't register with itself.

### `servant/RegistryServiceServant.go` — the phonebook logic

Package-level state: `chord *dht.Chord` (the shared phonebook storage), `isAliveCheck map[string]int` (consecutive-failure counters), and an unused `ServiceNodesMap` (leftover from a pre-DHT version that stored the book in a local map).

- **`CreateChord`** — `NewChord(":<port>", 1098)` or `JoinChord(":<port>", ":<basePort>", 1098)`. Port 1098 is the *Java-side* ring port; the node's gRPC port doubles as its ring node name.
- **`EncodeStringArray` / `DecodeStringArray`** — the storage format: the DHT only stores strings, so a service's address list is joined with `";"` (`"addr1;addr2;addr3"`) and split back on read. Empty string ↔ nil.
- **`CheckIfKeyInKeysAndSet` / `CheckIfKeyInKeysNoSet`** — existence checks done by fetching **all keys** from the ring and scanning linearly (the O(N) smell of Part IV); the `AndSet` variant initializes a missing key to `""`.
- **`Register(name, address)`** — ensure key exists → `Get` → decode → `append(addresses, address)` → encode → `Set`. A read-modify-write cycle, unsynchronized across the three registry processes.
- **`Unregister`** — same cycle, removing the address with the slice idiom `append(addresses[:i], addresses[i+1:]...)`; deletes the key entirely when the list empties.
- **`Discover(name)`** — existence check, then `Get` + decode.
- **`CheckIsAliveEvery10Seconds`** — a `time.Ticker` loop with a mutex-guarded `done` flag ensuring sweeps never overlap: if the previous sweep is still running when the ticker fires, the tick is skipped.
- **`CheckAllNodesStatus`** — the health sweep: for every service key, for every registered address, call `ServiceClientBaseDirect.IsAlive`. Failures increment `isAliveCheck[addr]`; any success resets it to 0. At **2 consecutive failures** the address is collected into `nodesToDelete` and removed from the DHT via the same read-modify-write cycle (or the whole key deleted if it was the last node). This is the self-healing loop.
- **`DeleteByValue`** — small slice-removal helper used by the sweep.

### `servant/RegistryServiceServant copy.go`

A dead file: the entire previous version of the servant, commented out line by line (its only live line is the `package` declaration). Kept as an informal backup — version control makes this unnecessary.

### `servant/dht/Chord.go` — the MetaFFI bridge to Java

The Go-side wrapper for the Java Chord class. The interesting part is `init()`:

- `metaffi.NewMetaFFIRuntime("openjdk")` boots an embedded JVM; `LoadModule("./dht/Chord.class")` loads the compiled Java class (relative path — the process must start from `output/`).
- It then "imports" each Java member into a Go function variable using MetaFFI's signature strings, e.g.:

  ```go
  set, err = chordModule.Load("class=dht.Chord,callable=set,instance_required",
      []IDL.MetaFFIType{IDL.HANDLE, IDL.STRING8, IDL.STRING8}, nil)
  ```

  Read that as: "give me Java's `Chord.set(String, String)` as a Go func; it's an instance method, so the first argument is the object handle." Both constructors are loaded separately (2-arg = create ring, 3-arg = join ring), plus `get`, `delete`, `getAllKeys` (note `LoadWithAlias` with `Dimensions: 1` for the string-array return), and a *getter for the `isFirst` field* (`field=isFirst,getter`).
- The `Chord` struct holds a single `MetaFFIHandle` — an opaque pointer to the Java object living on the JVM heap.
- Every exported method (`Set`, `Get`, `Delete`, `GetAllKeys`, `IsFirst`) is a thin wrapper: pass the handle plus arguments to the bound function variable, type-assert the `[]interface{}` result (`res[0].(string)`, etc.), return Go-typed values. `NewChord`/`JoinChord` call the constructor bindings and wrap the returned handle.
- A commented-out `leave` binding hints at a planned graceful-departure feature that never shipped.

### `servant/dht/Chord_test.go`

Three smoke tests that create a ring on `:5000` (Java port 1099), join `:5001` and `:5002` to it, and perform a few `Set` / `GetAllKeys` calls. No assertions on values — they only verify nothing errors or panics; really a "does the FFI plumbing work at all" check.

### `servant/dht/*.class` (+ `backup/`)

Compiled Java bytecode for the DHT: `Chord`, `NodeClass` (plus an anonymous inner class `NodeClass$1`), `NodeStruct`, `KeyValueTable`, `TestChord`. This is the actual Chord implementation — hashing, finger tables, successor routing, storage. **The Java source is not in the repo**, only bytecode (copies exist in four places: here, `backup/`, `interop/backup/`, and `output/dht/`). The Go side treats it as a black box.

### `client/RegistryServiceClient.go`

The one client that can't bootstrap through discovery (chicken-and-egg), so it takes explicit addresses:

- `NewRegistryServiceClient(addresses)` returns `nil` if the list is empty — callers must check; one place doesn't.
- `PickRandomRegistry()` + `_Connect()` — pick a random registry, `grpc.Dial`, wrap with the generated stub factory, return stub + close closure.
- `Discover` / `Register` / `Unregister` — one method each: connect, fire the RPC with the right wrapper type (`wrapperspb.StringValue` / `ServiceRequest`), translate errors. Note `c, closeFunc, _ := obj._Connect()` discards the connect error — if all registries are down, the following `defer closeFunc()` panics on nil.

## Chapter 12: The Cache Service

### `common/` — `CacheService.proto` + generated stubs

Contract: `SetRequest{key, value}`; RPCs `Set`, `Get`, `Delete`, `IsAlive`, and `IsRoot` — that last one exists purely for the ring-bootstrap protocol below. Generated files follow the same pattern as the registry's.

### `service/CacheService.go`

The thinnest of the three services:

- `Start`: build the handler-binding closure → `services.Start("CacheService", 0, bind, nil)` (port 0 = ephemeral; `nil` messageHandler = **no MQ plane** for the cache) → `CreateChord(port)` → `register()` → `startListening()`.
- Handlers `Get` / `Set` / `Delete` unwrap and delegate to the servant; `IsAlive` returns `true`; `IsRoot` surfaces the servant's ring-root check to the network so *other* cache nodes can find the ring during bootstrap.

### `servant/CacheServiceServant.go`

- **`_GetRegistryAddresses()`** — private re-read of the YAML in `os.Args[1]` (duplicates `common.LoadRegistryAddresses`).
- **`CreateChord(port)`** — the decentralized bootstrap algorithm:
  1. Ask the registry: `Discover("CacheService")` — who's already alive?
  2. For each peer, call `IsRoot()` directly (via `ServiceClientBaseDirect`).
  3. If a root exists → `Chord.JoinChord("CacheService<port>", "root_node", 1097)` — join the existing data ring.
  4. If nobody claims root → `Chord.NewChord("root_node", 1097)` — *become* the ring's first node.

  Since caches register *before* `CreateChord` runs (ordering in `Start`), later nodes can always find earlier ones. The ring's Java port is 1097 — deliberately different from the registry's ring (1098) so the two DHTs never mix.
- **`Get(key)`** — existence check via `_CheckIfKeyInKeys` (the `GetAllKeys` linear scan again), then `chord.Get`. On error *or* missing key it returns `("", nil)` — callers cannot tell "empty value" from "no such key."
- **`Set` / `Delete`** — direct passthroughs to the ring (`Delete` also pre-checks existence).
- **`IsRoot()`** — wraps `chord.IsFirst()`, i.e., reads the Java object's `isFirst` field through FFI.
- `GetPortFromNode` — string-split helper, effectively unused.

### `client/CacheServiceClient.go`

Typed client built on the generic base:

```go
type CacheServiceClient struct {
    services.ServiceClientBase[service.CacheServiceClient]
}
```

The constructor takes registry addresses *or* `nil` (then it loads them from the YAML — the mode the black-box tests use). `Set` / `Get` / `Delete` / `IsAlive` all follow the identical five-line recipe: `Connect()` → `defer closeFunc()` → fire RPC → translate error → unwrap value. (The copy-paste tell: all its error messages say `"could not call IsAlive"`, even in `Set` and `Get`.)

## Chapter 13: The Test Service

### `common/` — `TestService.proto` + generated stubs

The richest contract: `StoreKeyValue`, `ExtractLinksFromURLParameters{url, depth}`, `ExtractLinksFromURLReturnedValue{repeated links}`, and seven RPCs including `WaitAndRand(Int32Value) returns (stream Int32Value)` — the `stream` keyword makes it **server-streaming**: instead of one reply, the server gets a channel it can push many replies down. The generated `_grpc.pb.go` (377 lines) accordingly contains extra streaming machinery (`TestService_WaitAndRandServer` with a `.Send()` method server-side, `.Recv()` client-side).

### `service/TestService.go`

Two jobs — gRPC transport *and* the MQ dispatcher:

- **`messageHandler(method, parameters)`** — the async plane's routing table. A big `switch method` where every case follows the same template: allocate the right parameter type (`&ExtractLinksFromURLParameters{}`, `&wrappers.StringValue{}`, …), `proto.Unmarshal(parameters, p)`, then call **the same gRPC handler method** on `serviceInstance` with a background context. This is the key design move: sync (gRPC) and async (ZMQ) requests converge on one implementation, so behavior can't diverge. Unknown methods return an error that ships back in `ReturnValue.Error`.
- **`Start`** — creates the singleton `serviceInstance`, wires the handler binding and `messageHandler` into `services.Start("TestService", 0, ...)`, registers (deferring unregister), logs the port, and blocks on `startListening()`.
- **The handlers** — each unwraps protobuf and delegates to the servant:
  - `HelloWorld` / `HelloToUser` — one-liners returning wrapped strings.
  - `Store` / `Get` — delegate to the servant (which delegates to the cache cluster).
  - `WaitAndRand` — the streaming one: it wraps `streamRet.Send` into a plain callback `func(x int32) error` and hands *that* to the servant — so the servant stays 100% ignorant of gRPC streaming.
  - `IsAlive` — returns `true` (this is what the registry's health sweep calls).
  - `ExtractLinksFromURL` — unwraps url/depth, wraps the returned `[]string`.

### `servant/TestServiceServant.go`

The business logic, plus the Python bridge:

- **`init()`** — boots the embedded CPython 3.11 via MetaFFI (`NewMetaFFIRuntime("python311")` → `LoadRuntimePlugin()`), loads `./crawler.py` as a module, and binds `extract_links_from_url(string, int64) -> []string` into a Go func variable. Any failure panics — the service is useless without its crawler. Because this runs at *import time*, merely importing this package requires a working Python + MetaFFI environment.
- **`HelloWorld` / `HelloToUser`** — pure string functions (the unit-testable core of the demo).
- **`Store(key, value)` / `Get(key)`** — construct a `CacheServiceClient` (freshly, per call — a hot-path smell) with registry addresses loaded from the config file, and call `Set` / `Get` on the cache cluster. Service-to-service delegation in action.
- **`WaitAndRand(seconds, sendToClient)`** — `time.Sleep`, then push `rand.Intn(10)` through the injected callback. Blocking a goroutine with Sleep is fine in Go — each gRPC request already runs in its own goroutine.
- **`ExtractLinksFromURL(url, depth)`** — calls the bound Python function and type-asserts `res[0].([]string)`. One line of Go running a Python web crawler in-process.
- Leftover: `cacheMap map[string]string` — the original in-memory store from before the cache service existed; initialized but never used.

### `servant/crawler.py`

The 15-line Python crawler: if `depth == 0` return; `requests.get(url)`; parse HTML with BeautifulSoup; for every `<a>` tag whose `href` starts with `http`, append it and **recurse** with `depth-1`, extending the result list. Simple depth-limited DFS — no visited-set, no timeout. (The deployed copy in `output/crawler.py` adds a bare `try/except` so one broken URL doesn't kill the whole crawl.)

### `client/TestServiceClient.go`

The biggest client; three tiers of methods:

1. **Sync wrappers** (`HelloWorld`, `HelloToUser`, `Store`, `Get`, `IsAlive`, `ExtractLinksFromURL`) — the standard connect → call → unwrap recipe via the generic base. (Several copy-pasted error strings claim the failing call was `HelloToUser`.)
2. **`WaitAndRand(seconds)`** — the streaming client: it calls the RPC to obtain a stream handle, then returns a closure that does `r.Recv()` — the caller decides *when* to block for the streamed value; the connection closes when the closure runs.
3. **The `*Async` methods** (`HelloWorldAsync`, `HelloToUserAsync`, `GetAsync`, `StoreAsync`, `ExtractLinksFromURLAsync`, `IsAliveAsync`) — the **future pattern** over ZMQ. Each one:
   - `ConnectMQ()` — REQ socket connected to every `TestServiceMQ` endpoint;
   - `NewMarshaledCallParameter("MethodName", args)` — build the envelope;
   - `SendBytes` — fire and *return immediately*;
   - returns a closure that, when later invoked, does `RecvBytes` (blocking), `UnmarshalReturnValue`, checks `rv.Error`, and `ExtractInnerMessage` into the right typed result.

   Callers write: `future, _ := c.ExtractLinksFromURLAsync(url, 1)` … do other work … `links, err := future()`. The six methods are structurally identical (~40 lines each, differing only in types) — prime refactoring material; the leftover `fmt.Printf("Im here 1")` debug prints live here too.

### `client/TestServiceClient_test.go`

Per-method integration tests (require a running cluster plus a `./configurations/RegistryAddresses.yaml` next to the test's working directory): one test per RPC, each creating a client, calling once, and logging the response; `TestWaitAndRand` demonstrates the promise flow. A commented-out block at the bottom preserves the very first version of the test from before discovery existed — when the constructor took a hardcoded `"localhost:50051"`. A tiny fossil record of the project's evolution.

## Chapter 14: The Interop Layer: Hand-Written FFI

Both `.go` files here start with `//go:build interop` — a build tag that **excludes them from normal builds**. They only compile with `go test -tags interop`, so the heavy cgo dependencies (JVM headers, Python headers) don't burden the main binary. This folder is "how FFI works under the hood" — the manual counterpart to what MetaFFI automates.

### `ChordDHT.go` — the raw JNI bridge

Structure: a giant C code block inside the cgo comment, then Go wrappers.

**The C half:**

- `#cgo CFLAGS/LDFLAGS` point the compiler at the JDK's headers and `libjvm`.
- `init_jvm()` — `JNI_CreateJavaVM(&jvm, ...)` boots a JVM inside this process; maps each JNI error code to a readable message.
- `get_env()` — JNI's `JNIEnv*` is per-thread, and Go moves goroutines across OS threads, so *every* call must fetch the env and `AttachCurrentThread` if needed.
- `get_exception_message()` — when Java throws, this fetches the `jthrowable`, calls its `toString()` reflectively, copies the result to a C string, and clears the exception — converting Java exceptions into C error strings.
- `load_chord_class()` — because the class isn't on the default classpath, it builds a `java.net.URLClassLoader` pointing at the interop folder, loads `dht.Chord` through it, promotes it to a **global ref** (so Java's GC can't collect it), and caches every method ID (`GetMethodID`) for the constructors, `set`, `get`, `delete`, and `getAllKeys`.
- One C wrapper per operation — e.g. `call_method_set` converts C strings → `jstring`s, `CallVoidMethod`, checks for exceptions, frees local refs; `call_method_get_all_keys` converts a Java `String[]` into a `NULL`-terminated `char**`, with careful per-element cleanup and full rollback on partial failure (the `cleanup_loop` / `goto` dance).
- `get_is_first_field` reads the boolean `isFirst` field directly with `GetBooleanField`.

**The Go half:** `ChordDHT{instance unsafe.Pointer}` holds the `jobject`; `LoadJVM`, `NewChordDHT`, `JoinChordDHT`, `Set`, `Get`, `Delete`, `GetAllKeys`, `GetIsFirst`, `DeleteObject` each convert Go strings to C strings (`C.CString` + `defer C.free` — C memory isn't garbage-collected!), call the C wrapper, check the out-parameter error, and convert results back with `C.GoString`.

### `Crawler.go` — the raw CPython bridge

The same exercise against CPython's C API:

- `load_cpython()` — `Py_Initialize()` plus appending `"."` to `sys.path` so `import crawler` can find the local file.
- `handle_error()` — Python's version of exception translation: `PyErr_Fetch` grabs type/value/traceback, stringifies each, and formats `"Name: value\ntraceback"` into a malloc'd C string.
- `extract_links_from_url()` — the full embedding ritual, heavily commented: **`PyGILState_Ensure()`** (Python's Global Interpreter Lock — only the GIL holder may touch the interpreter; this matters because Go calls in from arbitrary threads) → import module → `PyObject_GetAttrString` to get the function → build an argument tuple → `PyObject_CallObject` → validate the result is a list of strings → copy into a `NULL`-terminated `char**` → meticulous `Py_DECREF`/`Py_XDECREF` reference-count cleanup → `PyGILState_Release`.
- Go side: `ExtractLinksFromURL` converts, calls, then uses the classic cgo idiom to view the C array as a Go slice:

  ```go
  tmpslice := (*[1 << 30]*C.char)(unsafe.Pointer(c_result))[:length:length]
  ```

  ("pretend this pointer is a gigantic array, then slice exactly `length` elements from it"), copying each C string into a Go string and freeing as it goes.

### `interop/crawler.py`

The same crawler as the servant's copy (without the `try/except`).

### `Interop_test.go`

End-to-end proof for both bridges: crawls `microsoft.com` at depth 1 and asserts links came back; boots a 3-node Chord ring (create + two joins on port 9009), sets keys through different nodes, and verifies `GetAllKeys` / `Get` / `Delete` semantics across nodes — nicely demonstrating that data written via one node is visible from another.

### `interop/backup/*.class`

Another copy of the Java bytecode, positioned where `load_chord_class`'s hardcoded `file:///workspaces/large-scale-workshop/interop/` URL expects it.

## Chapter 15: Black-Box Testing

These aren't unit tests — they're **acceptance tests** run (with `go test`) *while the full 9-process cluster is up*, acting as a real external client. Each folder carries its own `configurations/RegistryAddresses.yaml` (all three registries) because `ServiceClientBase.LoadRegistryAddresses` resolves that path relative to the test's working directory.

### `testing/testservice-testing/tester_test.go`

One long scenario (`TestTester1`) exercising nearly everything:

- Two independent clients (`c1`, `c2`) — each call re-discovers, so calls scatter across the 3 TestService instances.
- Sync calls: `HelloToUser("Obayda")`, `HelloWorld`, `ExtractLinksFromURL`.
- **The distributed-state proof:** `c1.Store("Obayda", "Production down")` then `c2.Get("Obayda")` must return the value — write through one TestService instance, read through another, with the data actually living in the cache ring. If any link in the chain is broken, this fails.
- **Async interleaving:** kicks off `ExtractLinksFromURLAsync`, does sync work, *then* redeems the future — demonstrating real overlap.
- Streaming: the `WaitAndRand(3)` promise.

### `testing/cache-testing/tester_test.go`

Talks to the cache directly through `CacheServiceClient` (constructed with `nil` → loads registry YAML): set→get roundtrip, **overwrite** semantics (set the same key twice, expect the new value), **delete** semantics (get after delete returns `""`), then a multi-key sequence (`Deeb` / `Ahmad` / `Obayda`) to ensure keys hash to — and survive across — different ring nodes.

## Chapter 16: Build & Deployment: `output/`

This folder is the *runtime working directory*; every relative path in the code (`./dht/Chord.class`, `./crawler.py`, `./configurations/*.yaml`) resolves against it.

- **`build.sh`** — `go build -buildvcs=false -o ./output/large-scale-workshop` — compiles the whole module into one binary (run from the repo root). `-buildvcs=false` skips embedding git metadata (avoids failures in containers where repo ownership looks odd to git).
- **`RunRegistry.sh` / `RunTestService.sh` / `RunCache.sh`** — each exports `CLASSPATH="./dht/:./"` (so the embedded JVM can find `dht/Chord.class`) and executes the binary with the matching YAML. Same binary, three costumes.
- **`start.sh`** — the cluster orchestrator: traps SIGINT with a `cleanup` that runs `kill 0` (kills the entire process group — every service dies with one Ctrl-C), then launches **3× registry → 3× test → 3× cache**, each backgrounded with `&` and spaced by `sleep 5`. The ordering and the sleeps are load-bearing: registries must exist before anyone can register, and the first cache must finish creating the ring before the second tries to join it. `wait` at the end keeps the script alive as the parent.
- **`configurations/*.yaml`** — the live configs: registry at base port 8502; test/cache pointing at all three registries `127.0.0.1:8502-8504`.
- **`output/crawler.py`** — the deployed crawler; this copy wraps the fetch in `try/except: pass` so unreachable links are skipped silently.
- **`output/dht/*.class`** — the deployed Java bytecode that the launch scripts' CLASSPATH points at.

### The one-paragraph mental model

If you remember nothing else: **`main.go` is a costume rack; `services/common/` is the framework everyone shares (server lifecycle + discovery-aware generic client + ZMQ future protocol); each service is a four-layer sandwich (proto contract → gRPC transport → servant logic → typed client); the registry keeps the phonebook inside a Java Chord ring, the cache keeps the data inside a *second* Java Chord ring, and the test service is the app that uses both — with Java and Python running *inside* the Go processes via MetaFFI, and `output/start.sh` conducting all nine instruments.**

---

---

# Part IV: Performance Bottlenecks & Refactoring

*No honest architecture document ends at "it works." This part is the critical-thinking chapter: where the system would strain under load, where the code has latent bugs, and what a path to production would look like. Findings are prioritized by impact.*

## Chapter 17: Performance & Scalability

1. **Discovery + dial on every single call.** `Connect()` dials a registry, runs Discover, dials the target, executes one RPC, and closes everything (`ServiceClientBase.go`). With `grpc.WithBlock()` that's ~3 full TCP + HTTP/2 handshakes per logical operation. **Fix:** cache the gRPC `ClientConn` (they're multiplexed and designed to be long-lived) and cache/TTL discovery results.
2. **`GetAllKeys()` as an existence check.** `CheckIfKeyInKeys*` (registry servant) and `_CheckIfKeyInKeys` (cache servant) enumerate *the entire DHT keyspace* on every Discover, Register, cache Get, and cache Delete. This is O(N) over the whole ring and defeats the O(log N) purpose of Chord. `Get` returning empty/null already distinguishes "missing" — use it.
3. **Config file I/O on the hot path.** `TestServiceServant.Store/Get` call `LoadRegistryAddresses()` → `os.ReadFile` + YAML parse, and construct a brand-new `CacheServiceClient`, per request. Load once at startup; keep one client.
4. **The crawler is exponential and unguarded.** `crawler.py` has no visited-set (it revisits pages and can bounce between mutually-linking pages until depth runs out), no request timeout, and runs synchronously while holding the Python GIL inside the service process. One crawl of a link-heavy site can stall a TestService instance.
5. **Health checking is centralized.** Only the base-port registry runs the liveness loop; if that process dies, dead nodes are never evicted — a single point of failure in an otherwise replicated control plane.

## Chapter 18: Correctness & Concurrency

6. **Registry read-modify-write races.** `Register` / `Unregister` / eviction all do `Get → decode → mutate → Set` on the shared `name → "a;b;c"` string with no locking or compare-and-swap across three registry processes. Two concurrent registrations can silently drop one another (a lost update). Ironically, `ServiceNodesMap` with its mutex exists but is never used — and since the DHT is shared across *processes*, a local mutex wouldn't suffice anyway. **Fix:** store one key per instance (e.g., `TestService/<addr>`) or add versioned writes.
7. **ZMQ REQ/REP protocol violations.** The server receives on the REP socket, then replies **from a spawned goroutine** (`ServiceBase.go`). ZMQ sockets are not thread-safe, and REP requires strict recv→send alternation; two in-flight requests can interleave sends and corrupt the socket. Several error paths also `return` without sending *any* reply, which permanently wedges the client's REQ socket (the future's `RecvBytes` blocks forever). **Fix:** use ROUTER/DEALER for concurrency, or keep REP strictly sequential and always reply.
8. **The MQ liveness probe is meaningless.** `IsAlive` for `*MQ` services "checks" by connecting a **REP** socket to another REP endpoint — REP-to-REP is an invalid pairing, and ZMQ `connect` is asynchronous and succeeds even when nothing is listening. It always returns `true`, so dead MQ endpoints are never evicted.
9. **Swallowed errors.** The registry's `Register`/`Unregister` handlers return success even when the servant errors; `_Connect()`'s error is discarded with `_` in all three registry-client methods — if every registry is down, `closeFunc()` nil-panics; cache `Get` maps *errors* to `("", nil)`, making failures indistinguishable from a missing key.
10. **Unsynchronized shared maps.** `isAliveCheck` is written from the health goroutine with no lock; safe today only because the done-flag prevents overlapping sweeps — fragile.
11. **Registry bootstrap assumptions.** The port-probing loop (`for { i++ }`) spins forever on any non-port error, and "base port = ring creator" breaks if instance 1 isn't first or restarts (a restarted base node will try to *create* a ring its peers already own).

## Chapter 19: Design Hygiene & Maintainability

12. **Misplaced shared dependency.** The Chord Go wrapper lives in `services/registry-service/servant/dht/` but is imported by `cache-service` — a cross-service reach-in. Promote it to a top-level shared package (e.g., `pkg/dht`).
13. **Dead code and artifacts.** `RegistryServiceServant copy.go` (fully commented out), the unused `ServiceNodesMap` and `cacheMap`, commented-out blocks in `ServiceBase.go` and `Chord.go`, duplicated `.class` files in four locations, and `interop/` duplicating `dht/Chord.go`'s role. Also debug prints (`"Im here 1..6"`) and copy-pasted error strings (`"could not call HelloToUser"` in `Get`, `Store`, `IsAlive`; `"could not call IsAlive"` in the cache client's `Set`/`Get`/`Delete`) that will actively mislead whoever debugs this.
14. **Massive duplication in async clients.** The six `*Async` methods are ~40 near-identical lines each, differing only in method name and payload types — one generic `callAsync[TReq, TResp]` helper would collapse ~200 lines.
15. **Hardcoded magic values.** Chord ports 1097/1098, `"root_node"`, the `./configurations/RegistryAddresses.yaml` path, `./dht/Chord.class`, `./crawler.py`, and the hand-built method string `"/%s.%s/IsAlive"` (breaks if a proto package doesn't match `strings.ToLower(serviceName)`). All CWD-dependent — the binary only works when launched from `output/`.
16. **Config loading via `os.Args[1]` inside library code** (`services/common/ServiceBase.LoadRegistryAddresses`) couples the framework layer to the CLI invocation; inject config instead.
17. **README drift.** The README promises round-robin load balancing (the code is random), lists "gRPC-based communication" as a future improvement (it's already the primary transport), and claims key-value replication that can't be verified — the Java source isn't in the repo, only `.class` files. The Java source should be committed.

## Chapter 20: Security & Operational Concerns

18. **No transport security or authentication.** `grpc.WithInsecure()` everywhere; no authn/authz on Register/Unregister (anyone who can reach the registry can evict or hijack a service name); the crawler will fetch arbitrary URLs (an SSRF surface). Acceptable for a workshop; must be stated as a non-goal.
19. **No persistence.** All state — phonebook and cache — is in-memory across the rings; a full-cluster restart loses everything.

## Chapter 21: The First Four Refactoring Tickets

For a university workshop project, the architectural literacy here is genuinely impressive — a clean service/servant/client layering, a real discovery-based control plane, dual sync/async transports with a future abstraction, and *two* independently implemented FFI bridges. The gap between this and production is concentrated in four places, in priority order:

1. **Connection reuse** — long-lived `ClientConn`s and cached discovery (item 1).
2. **Atomic registry state** — per-instance keys or versioned writes to end the lost-update race (item 6).
3. **A sound ZMQ concurrency model** — ROUTER/DEALER or strictly sequential REP with guaranteed replies (item 7).
4. **Error transparency** — stop swallowing errors at the transport boundary; make "not found" distinguishable from "failed" (item 9).

---

---

# Part V: The Builder's Roadmap

*Everything you've read is achievable. This part turns the book into action: a five-phase plan to build a simplified DTOY yourself, followed by the exact tools you need. Each phase is a weekend or two, and each one leaves you with something working. Don't skip ahead — the phases stack on purpose.*

## Chapter 22: The Five-Phase Roadmap

### Phase 1: The Basics — Go + a plain HTTP server

**Goal: get comfortable with Go and prove you can make two programs talk.**

1. Do the official *Tour of Go* (interactive, a few evenings). Focus on: functions, structs, slices, maps, and — Go's special sauce — goroutines (the tour explains them).
2. Write a classic HTTP server using Go's built-in `net/http` package: visiting `http://localhost:8080/hello?name=You` returns `"Hello You"`.
3. Write a *second* Go program that calls your server using `http.Get` and prints the result.

✅ **You're done when:** one terminal runs your server, another runs your client, and they talk. That's already a two-node distributed system. Seriously.

### Phase 2: The Contract — your first `.proto` file

**Goal: define a contract and generate code from it.**

1. Install the Protobuf compiler (`protoc`) and the Go plugins (Chapter 23).
2. Write `hello.proto` with a `HelloRequest { string name }`, a `HelloReply { string message }`, and a service with one `rpc SayHello(HelloRequest) returns (HelloReply)`.
3. Run `protoc` and just *look* at the generated `.go` files. You don't need to understand every line — notice that it wrote a Go struct for each message and an interface for the service, for free.

✅ **You're done when:** `protoc` runs cleanly and you can explain to a rubber duck what each message field means.

### Phase 3: Client & Server — first real gRPC call

**Goal: replace HTTP with gRPC.**

1. Follow the official gRPC-Go "Quick start" — it walks you through wiring the generated code into a server.
2. Implement `SayHello` on the server (just return `"Hello " + name`).
3. Write a client that connects and calls it like a normal function.
4. **Stretch within the phase:** add `Store(key, value)` and `Get(key)` RPCs backed by a plain Go `map[string]string`. (A real-world wrinkle you'll hit: two requests can touch the map at once — Google "go mutex" when it happens. This exact issue bit DTOY too; see Part IV, item 6.)

✅ **You're done when:** your client calls a function that runs in another process, and it feels like cheating.

### Phase 4: The Phonebook — a mini Registry

**Goal: kill the hardcoded address. This is the phase where it becomes a *distributed system*.**

1. Build a second gRPC service, `Registry`, with three RPCs: `Register(name, address)`, `Unregister(name, address)`, `Discover(name) → list of addresses`. Its storage? Just a Go map of `name → list of addresses` (with a mutex!). *You don't need a Chord ring — DTOY used one because that was the point of the workshop; a map is the honest simple version.*
2. Change your Phase-3 server: on startup, listen on port `0` (the OS picks a free port), then call `Register("HelloService", myAddress)` on the Registry.
3. Change your client: instead of a hardcoded address, call `Discover("HelloService")` first, pick a random result, then connect.
4. **The payoff moment:** launch *three copies* of your server. Run the client in a loop. Watch requests land on different copies. You built a load balancer.
5. **Stretch:** add an `IsAlive()` RPC to your server, and make the Registry ping everyone every 10 seconds, removing anyone who fails twice in a row. Then kill a server mid-run and watch the phonebook heal. (This is exactly what DTOY does.)

✅ **You're done when:** nothing except the Registry's address is hardcoded anywhere.

### Phase 5: The Stretch Goal — a second service

**Goal: make services use each other.**

Pick one (or both):

- **Option A — a Cache Service:** a standalone key-value gRPC service (`Set`/`Get`/`Delete`, map + mutex) that registers itself in your Registry. Then rewrite your first service's `Store`/`Get` to *discover the cache and delegate to it* — recreating the exact request trace from Chapter 7. Run two cache copies and think about the hard question: if `Set` goes to cache A and `Get` goes to cache B, how would they share data? (You've just discovered why Chord rings exist. You don't have to solve it — *understanding the problem* is the win.)
- **Option B — a fun worker service:** a `LinkExtractor` service that fetches a URL and returns the links on the page (Go's `net/http` + the `golang.org/x/net/html` package can do this — no Python needed).

✅ **You're done when:** a request from your client touches three different processes before coming back — and you can draw Chapter 7's diagram from memory.

Finish all five phases and you'll have built roughly **70% of the ideas** in DTOY. The remaining 30% — Chord, multi-language FFI, async messaging — is exactly what upper-year workshops are for.

## Chapter 23: Prerequisites, Tools & the "Google This" Cheat Sheet

### Install these (in order)

1. **Go** — from `go.dev/dl`. Verify with `go version`.
2. **Protocol Buffers compiler** — `brew install protobuf` on Mac (or grab a release from GitHub on Windows/Linux). Verify with `protoc --version`.
3. **The two Go plugins for protoc** (these generate the Go code):

   ```
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

4. **A good editor** — VS Code with the official Go extension is the easy choice.
5. *(Optional but delightful)* **grpcurl** — like a phone you can use to prank-call your own gRPC services from the terminal. Great for debugging.

### Your "Google this when stuck" cheat sheet

| When you're stuck on... | Search for... |
|---|---|
| Go basics | "A Tour of Go", "Go by Example" |
| Your first gRPC program | "gRPC Go quickstart" |
| Proto file syntax | "protobuf language guide proto3" |
| Two requests corrupting your map | "golang mutex map concurrent access" |
| Running things at the same time | "golang goroutines channels tutorial" |
| The phonebook idea in the real world | "service discovery", "Consul", "etcd" |
| The circle-of-friends idea | "consistent hashing explained", "Chord DHT" |
| Why microservices at all | "monolith vs microservices" |

### One last thing

Every single concept in this book confused its authors at first too. On day one, none of them could have told you what a DHT was. The trick is that you never climb the whole mountain — you just climb to the next base camp, and the view keeps getting better. Phase 1 is one weekend away.

Now go make two programs talk to each other. 🚀

---

*— End of Guidebook —*
