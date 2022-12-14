package main

import (
	"bufio"
	"encoding/json"
	"log"
	"math"
	"math/big"
	"net"

	"github.com/russellslater/protohackers"
)

type primeRequest struct {
	Method string   `json:"method"`
	Number *float64 `json:"number"`
}

type primeResponse struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func main() {
	log.Fatal(protohackers.ListenAndAccept(5000, handle))
}

func handle(conn net.Conn) error {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Bytes()

		log.Println("received:", string(line))

		resBytes, valid, err := handleLine(line)
		if err != nil {
			return err
		}

		if _, err := conn.Write(resBytes); err != nil {
			return err
		}

		// stop processing if the request was invalid
		if !valid {
			break
		}
	}

	return nil
}

func handleLine(line []byte) ([]byte, bool, error) {
	var req primeRequest
	if err := json.Unmarshal(line, &req); err != nil || !isValidPrimeRequest(req) {
		return []byte("invalid request\n"), false, nil
	}

	resBytes, err := json.Marshal(primeResponse{Method: "isPrime", Prime: isPrime(*req.Number)})
	if err != nil {
		return nil, true, err
	}

	return append(resBytes, []byte("\n")...), true, nil
}

func isValidPrimeRequest(req primeRequest) bool {
	return req.Method == "isPrime" && req.Number != nil
}

func isPrime(n float64) bool {
	// prime numbers are positive integers
	if n < 0 || n != math.Trunc(n) {
		return false
	}
	return big.NewInt(int64(n)).ProbablyPrime(20)
}
