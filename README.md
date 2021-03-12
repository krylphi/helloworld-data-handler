# helloworld-data-handler

1. Create a HTTP server using the https://github.com/valyala/fasthttp library
2. Handle simple JSON request of 
{"text": "hello world", "content_id": x, "client_id":y, "timestamp": now}  
* where x is a counter from 1 to 1 billion  
* where y is a random number between 1 and 10  
* where now is right now with millisecond precision  
3. Stream the data to AWS S3 in http://ndjson.org/ format with gzip compression  
a.it is important the data is streamed and not stored to disk  
b. it is okay if the file ending is corrupted but better it is not  
c. the filename should be "/chat/{{date}}/content_logs_{{date}}_{{client_id}}  
    -  where date is in the format YYYY-MM-DD as in 2020-03-30  
    -  where client_id is the value from the json message and secure against injection attacks.  
4. If the server crashes, hits an exception or is terminated by the server (for example when a docker pod is scaled down) it will attempt to flush the current stream.  


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

AWS:
- AWS_ACCESS_KEY - aws Access Key ID (**required**)
- AWS_ACCESS_SECRET - aws Secret Access Key (**required**)
- AWS_REGION - aws region (**required**)
- AWS_BUCKET - aws bucket (**required**)

AZURE:
- AZURE_ACC - azure username (**required**)
- AZURE_KEY - azure access key (**required**)
- AZURE_CONTAINER - azure container (**required**)

## Routes
POST /log - send message

## Benchmarks & Profiling Result
Benchmarks were performed by Gatling run on the developer's virtual machine \
(8gb RAM, i7 (4 cores provided for Virtual Machine)).\
Full performance benchmarks are in the benchmark repo: [bencharks](https://github.com/krylphi/helloworld-data-handler-benchmark)

- max RAM consumption on the server is near 320 Mb under stress load and less than 100 Mb idle.
- server accepts approx. 15000-30000 queries per second with 0 KO rate and 86% responses in less < 800 ms under stress load (see Gatling benchmarks).

## Logged time
It took aprox. 29 hours to implement the server, including aprox 5-7 hours research.
