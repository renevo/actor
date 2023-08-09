# Simple Actor Framework

This is heavily based on [anthdm's Hollywood Actors](https://github.com/anthdm/hollywood).

However, the goals of this project is to take it a bit further.

TODO List:

* [x] Basic Actor Model
* [x] Port more tests from hollywood
* [x] Dead letter
* [x] Middleware
* [x] Repeaters
* [x] RPC (Request / Reply)
* [x] Context passthrough (Send needs to be modified to pass them through)
* [ ] Events
* [ ] Observability
  * [ ] Logging
  * [ ] Metrics
  * [ ] Tracing
* [ ] Go Docs
* [x] CI
* [ ] Remote Actors - this is where this will heavily deviate

## Disclaimer

This is under heavy development, and should not be used before a 1.0.0 release, as the API will most likely substantially change before then.

## What is the actor model?

The Actor Model is a computational model used to build highly concurrent and distributed systems. It was introduced by Carl Hewitt in 1973 as a way to handle complex systems in a more scalable and fault-tolerant manner.

In the Actor Model, the basic building block is an actor, called receiver in this package, which is an independent unit of computation that communicates with other actors by exchanging messages. Each actor has its own state and behavior, and can only communicate with other actors by sending messages. This message-passing paradigm allows for a highly decentralized and fault-tolerant system, as actors can continue to operate independently even if other actors fail or become unavailable.

Actors can be organized into hierarchies, with higher-level actors supervising and coordinating lower-level actors. This allows for the creation of complex systems that can handle failures and errors in a graceful and predictable way.

By using the Actor Model in your application, you can build highly scalable and fault-tolerant systems that can handle a large number of concurrent users and complex interactions.
