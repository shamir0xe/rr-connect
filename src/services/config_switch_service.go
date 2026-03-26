package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type configSwitchStruct struct {
	templateCfg      string
	outputCfg        string
	placeholder      string
	router           RouterInterface
	systemctlService string
	notifyEnabled    bool
	notifyCommand    string
	notifyArgs       []string
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
		notifyEnabled:    sub.GetBool("notify.enabled"),
		notifyCommand:    sub.GetString("notify.command"),
		notifyArgs:       sub.GetStringSlice("notify.args"),
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

			err = cs.restartService()
			if err != nil {
				return err
			}

			go cs.notify(ctx)
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

func (cs *configSwitchStruct) restartService() error {
	log.Printf("ConfigSwitch -> restarting %s\n", cs.systemctlService)
	err := exec.Command("systemctl", "restart", cs.systemctlService).Run()
	return err
}

func (cs *configSwitchStruct) notify(ctx context.Context) error {
	if !cs.notifyEnabled {
		return nil
	}

	timer := time.NewTimer(2 * time.Second)
	<-timer.C
	log.Printf("ConfigSwitch -> executing notify command after 2 seconds delay\n")

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var args []string = make([]string, len(cs.notifyArgs))
	for i, arg := range cs.notifyArgs {
		if strings.Contains(arg, cs.placeholder) {
			args[i] = strings.ReplaceAll(arg, cs.placeholder, cs.router.Previous())
		} else {
			args[i] = arg
		}
	}

	log.Printf("%s ", cs.notifyCommand)
	for _, arg := range args {
		log.Printf("%s ", arg)
	}
	log.Println()

	err := exec.CommandContext(timeoutCtx, cs.notifyCommand, args...).Run()
	if err != nil {
		log.Printf("ConfigSwitch -> notify command failed: %v\n", err)
	}

	return err
}
