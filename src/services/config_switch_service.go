package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/spf13/viper"
)

type configSwitchStruct struct {
	templateCfg      string
	outputCfg        string
	placeholder      string
	router           RouterInterface
	systemctlService string
}

type ConfigSwitchInterface interface {
	Run(context.Context, *sync.WaitGroup, <-chan bool) error
}

func NewConfigSwitchService(cfg *viper.Viper, router RouterInterface) (ConfigSwitchInterface, error) {
	sub := cfg.Sub("config-switch")
	return &configSwitchStruct{
		router:           router,
		placeholder:      sub.GetString("placeholder"),
		templateCfg:      sub.GetString("template-cfg"),
		outputCfg:        sub.GetString("output-cfg"),
		systemctlService: sub.GetString("systemctl-service"),
	}, nil
}

func (cs *configSwitchStruct) Run(ctx context.Context, wg *sync.WaitGroup, triggerChan <-chan bool) error {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("ConfigSwitch -> END")
			return nil
		case <-triggerChan:
			err := cs.makeConfig()
			if err != nil {
				return err
			}

			log.Printf("ConfigSwitch -> restarting %s\n", cs.systemctlService)
			err = exec.Command("systemctl", "restart", cs.systemctlService).Run()
			if err != nil {
				return err
			}
		}
	}
}

func (cs *configSwitchStruct) makeConfig() error {
	log.Printf("ConfigSwitch -> creating new config [%s]\n", cs.router.Pick())
	input, err := os.Open(cs.templateCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer input.Close()

	output, err := os.Create(cs.outputCfg)
	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command(
		"sed",
		fmt.Sprintf("s|%s|%s|g", cs.placeholder, cs.router.Next()),
	)
	cmd.Stdin = input
	cmd.Stdout = output
	err = cmd.Run()
	return err
}
