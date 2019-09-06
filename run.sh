#!/bin/sh

docker build -t arca-jsonrpc .

docker run --rm arca-jsonrpc