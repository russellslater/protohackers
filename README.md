# Protohackers Challenges

A collection of my solutions for the server programming challenge at https://protohackers.com.

- Smoke Test: https://protohackers.com/problem/0
- Prime Time: https://protohackers.com/problem/1

## Instructions
Building a solution's Docker image (`smoke-test` in this example) and testing locally ...
```
$ docker build -f Dockerfile.smoke-test --tag smoke-test .
$ docker run -p 5000:5000 smoke-test
```

Interact with solutions locally using Netcat after running an image or with, for example, `go run ./cmd/smoke-test` ...
```
$ nc localhost 5000
```

To deploy to [fly.io](https://fly.io/) (requires account) ...
```
$ flyctl deploy --dockerfile Dockerfile.smoke-test
```
To test on fly.io ...
```
$ nc [FLY_HOSTNAME] 5000
```
For submission at https://protohackers.com, find the public IP address of your fly.io app with `flyctl ips list`.

Tail logs with ...
```
flyctl logs --app [APP_NAME]
```
