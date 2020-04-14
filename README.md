# helloworld-data-handler

## Installation
1. compile binary.
2. Set environment variables
3. Run

## Environment Variables
- ADDR - HTTP API address
- PORT - HTTP API port (default - 8902) (optional)
- AWS_ACCESS_KEY - aws Access Key ID (**required**)
- AWS_ACCESS_SECRET - aws Secret Access Key (**required**)
- AWS_REGION - aws region (**required**)
- AWS_BUCKET - aws bucket (**required**) 

## Routes
POST /log - send message

## Benchmarks & Profiling Result
Benchmarks were performed by Gatling run on the developer's virtual machine (8gb RAM, i7 (4 cores provided for Virtual Machine)).\
Full performance benchmarks are in th benchmark folder

- max RAM consumption on the server is near 320 Mb under stress load and less than 100 Mb idle.
- server accepts approx. 15000-30000 queries per second with 0 KO rate and 86% responses in less < 800 ms (see Gatling logs).

## Logged time
It took aprox. 17 hours to implement the server, including aprox 5-6 hours research.
