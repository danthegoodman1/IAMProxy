# AWS IAM Proxy

An AWS SigV4/IAM (WIP) proxy that enables you to make AWS-compatible APIs in any language.

The IAM proxy sits in front of your AWS-compatible APIs and verifies the signature of incoming requests, fetching information about the key owner from your API.

Currently, it only validates the signature, meaning that it verifies that the requester is who they say they are (owner if the key). In the future policy support may be added so that the proxy can look up a resource, and analyze user and resource policies to determine whether the user has access to the requested resource.

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