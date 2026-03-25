package services

import (
	"context"
	"fmt"
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
			fmt.Println("ConfigSwitch -> END")
			return nil
		case <-triggerChan:
			fmt.Printf("ConfigSwitch -> creating new config [%s]\n", cs.router.Pick())
			err := exec.Command(
				"sed",
				fmt.Sprintf("s|%s|%s|g", cs.placeholder, cs.router.Next()),
				"<", cs.templateCfg,
				">", cs.outputCfg,
			).Run()
			if err != nil {
				return err
			}

			fmt.Printf("ConfigSwitch -> restarting %s\n", cs.systemctlService)
			err = exec.Command("systemctl", "restart", cs.systemctlService).Run()
			if err != nil {
				return err
			}
		}
	}
}
