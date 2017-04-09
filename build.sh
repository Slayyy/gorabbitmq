#!/usr/bin/env bash

go build -o a admin/main.go && go build -o d doctor/main.go && go build -o t technician/main.go
