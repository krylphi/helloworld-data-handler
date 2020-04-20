# helloworld-data-handler
## Installation
1. compile binary.
2. Set environment variables
3. Run

## Environment Variables
- ADDR - HTTP API address (default=0.0.0.0) (optional)
- STORAGE - storage type (default=aws) (**aws/azure**) (optional)
- PORT - HTTP API port (default=8902) (optional)
- QUEUE_TIMEOUT_MIN - queue timeout before flushing (default=60) (optional)
- NO_UPLOAD - disable AWS uploading (_debug_) (default=0) (optional)
- AWS_ACCESS_KEY - aws Access Key ID (**required**) - aws
- AWS_ACCESS_SECRET - aws Secret Access Key (**required**) -aws
- AWS_REGION - aws region (**required**) -aws
- AWS_BUCKET - aws bucket (**required**)  -aws
- AZURE_ACC - azure username (**required**) - azure
- AZURE_KEY - azure access key (**required**) - azure
- AZURE_CONTAINER - azure container (**required**) - azure 

## Routes
POST /log - send message

## Benchmarks & Profiling Result
Benchmarks were performed by Gatling run on the developer's virtual machine \
(8gb RAM, i7 (4 cores provided for Virtual Machine)).\
Full performance benchmarks are in th benchmark folder

- max RAM consumption on the server is near 320 Mb under stress load and less than 100 Mb idle.
- server accepts approx. 15000-30000 queries per second with 0 KO rate and 86% responses in less < 800 ms under stress load (see Gatling benchmarks).

## Logged time
It took aprox. 29 hours to implement the server, including aprox 5-7 hours research.
