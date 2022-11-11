# Protohackers Challenges

A collection of my solutions for the server programming challenge at https://protohackers.com.

- Smoke Test: https://protohackers.com/problem/0
- Prime Time: https://protohackers.com/problem/1
- Means to an End: https://protohackers.com/problem/2
- Budget Chat: https://protohackers.com/problem/3
- Unusual Database Program: https://protohackers.com/problem/4
- Mob in the Middle: https://protohackers.com/problem/5
- Speed Daemon: https://protohackers.com/problem/6

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
## Testing Means to an End
The VS Code [Hex Editor](https://marketplace.visualstudio.com/items?itemName=ms-vscode.hexeditor) extension might be useful to you. Sending binary data might then look like `cat data.dat | nc localhost 5000` or the following (by converting a hexdump into binary) ...
```
echo '490000303900000065490000303a00000066' | xxd -r -p | nc localhost 5000
```

## Testing Unusual Database Program
To run the Docker image locally, override the `-host` flag (it defaults to a value required by fly.io otherwise) ...
```
docker run -p 5000:5000/udp unusual-database-program -host=0.0.0.0
```
You can now test UDP with Netcat using the `-u` flag ...
```
nc -u localhost 5000
```
Note that `^D` / `CTRL+D` will send `EOF`.

## Testing Speed Daemon
This could be tested locally by first creating a couple of named pipes (multiple pairs to test multiple simultaneous clients) ...
```
mkfifo out
mkfifo in
```
Making use of the pipes after starting the server ...
```
nc localhost 5000 <out >in &; cat > out
```
Sending commands from another terminal window ...
```
echo -e -n '\x80\x00\x7b\x00\x08\x00\x3c' > out # IAmCamera
echo -e -n '\x81\x03\x00\x42\x01\x70\x13\x88' > out # IAmDispatcher
echo -e -n '\x20\x04\x55\x4e\x31\x58\x00\x00\x03\xe8' > out # Plate
echo -e -n '\x40\x00\x00\x00\x0a' > out # WantHeartbeat
```
The `-n` flag will prevent trailing newlines from being sent, and the `-e` flag allows us to use hex notation.

And watching for the responses in yet another terminal window ...
```
cat in
```
A short example scenario with 3 connected clients ...
```
# Client 1: camera at mile 8
echo -e -n '\x80\x00\x7b\x00\x08\x00\x3c' > out # IAmCamera{road: 123, mile: 8, limit: 60}
echo -e -n '\x20\x04\x55\x4e\x31\x58\x00\x00\x00\x00' > out # Plate{plate: "UN1X", timestamp: 0}

# Client 2: camera at mile 9
echo -e -n '\x80\x00\x7b\x00\x09\x00\x3c' > out2 # IAmCamera{road: 123, mile: 9, limit: 60}
echo -e -n '\x20\x04\x55\x4e\x31\x58\x00\x00\x00\x2d' > out2 # Plate{plate: "UN1X", timestamp: 45}

# Client 3: ticket dispatcher (should receive a ticket after identifying)
echo -e -n '\x81\x01\x00\x7b' > out3 # IAmDispatcher{roads: [123]}
```