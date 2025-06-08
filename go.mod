module github.com/SiriusScan/app-terminal

go 1.23.0

toolchain go1.24.4

replace github.com/SiriusScan/go-api => ../go-api

require (
	github.com/SiriusScan/go-api v0.0.4
	github.com/streadway/amqp v1.1.0
)
