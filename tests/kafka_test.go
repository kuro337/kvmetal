package tests

import (
	"fmt"
	"testing"

	"kvmgo/configuration/presets"
	"kvmgo/constants/kafka"
)

func TestKafkaSettingsGeneration(t *testing.T) {
	kafkaConfig := presets.NewKafkaConfig("kafkavm")

	full := kafkaConfig.GenerateKraftUserdata(
		"kraft",
		"kraft.kuro.com",
		9095,
		"192.168.1.10",
		9094,
		"192.168.1.225",
		1,
		kafka.BrokerController)

	fmt.Println(full)

	// settings := kafkaConfig.GenerateKafkaSettings(
	// 	"kafkavm"i,
	// 	"kafka.kuro.com",
	// 	//"182.55.66.99",
	// 	9095,
	// 	"192.66.55.10",
	// 	9094,
	// 	"192.168.1.225",
	// 	1,
	// 	kafka.BrokerController)

	// //	fmt.Println(settings)
	// replaceCmd := presets.ReplaceKafkaKraftSettings(settings)

	// fmt.Println(replaceCmd)
	// runCmdReplaceIndented := presets.DefineKafkaSettingsInRunCmd(replaceCmd)

	// fmt.Println(runCmdReplaceIndented)

	t.Error("Trigger")
}
