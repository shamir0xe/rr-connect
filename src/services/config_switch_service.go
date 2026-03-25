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
	cfg.SetEnvPrefix("config-switch")
	return &configSwitchStruct{
		router:           router,
		placeholder:      cfg.GetString("placeholder"),
		templateCfg:      cfg.GetString("template-cfg"),
		outputCfg:        cfg.GetString("output-cfg"),
		systemctlService: cfg.GetString("systemctl-service"),
	}, nil
}

func (cs *configSwitchStruct) Run(ctx context.Context, wg *sync.WaitGroup, triggerChan <-chan bool) error {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-triggerChan:
			err := exec.Command(
				"sed",
				fmt.Sprintf("s|%s|%s|g", cs.placeholder, cs.router.Next()),
				"<", cs.templateCfg,
				">", cs.outputCfg,
			).Run()
			if err != nil {
				return err
			}

			err = exec.Command("systemctl", "restart", cs.systemctlService).Run()
			if err != nil {
				return err
			}
		}
	}
}
