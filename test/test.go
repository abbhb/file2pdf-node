package test

import "github.com/libreofficedocker/unoserver-rest-api/mqcode"

func main() {
	var producer = mqcode.CreateProducer()

	var consumers = mqcode.CreateConsumer(producer)
	consumers.StartConsumer()

}
