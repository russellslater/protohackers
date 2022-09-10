# protohackers

A collection of my solutions for the server programming challenge at https://protohackers.com.

## Smoke Test
Problem: https://protohackers.com/problem/0

To build the Docker image and test locally ...
```
$ docker build -f Dockerfile.smoke-test --tag smoke-test .
$ docker run -p 5000:5000 smoke-test
```
To test locally ...
```
nc localhost 5000
```

To deploy to [fly.io](https://fly.io/) (requires account) ...
```
flyctl deploy --dockerfile Dockerfile.smoke-test
```
To test on fly.io ...
```
nc [FLY_HOSTNAME] 5000
```
For submission at https://protohackers.com, find the public IP address of your fly.io app with `flyctl ips list`.
