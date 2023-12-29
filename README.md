# AWS IAM Proxy

<!-- TOC -->
* [AWS IAM Proxy](#aws-iam-proxy)
  * [Running](#running)
  * [Added Headers](#added-headers)
<!-- TOC -->

## Running

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Added Headers

`x-req-id`: The request ID, for log association

`x-span-id`: The OTLP span ID of the request if tracing is enabled.