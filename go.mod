module github.com/SiriusScan/app-terminal

go 1.22.5

replace github.com/SiriusScan/go-api => ../go-api

require github.com/SiriusScan/go-api v0.0.3

require (
	github.com/KnicKnic/go-powershell v0.0.10 // indirect
	github.com/creack/pty v1.1.24 // indirect
	github.com/streadway/amqp v1.1.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
)
