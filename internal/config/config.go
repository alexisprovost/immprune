package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

type Config struct {
	ImmichURL string `mapstructure:"immich_url"`
	ImmichKey string `mapstructure:"immich_key"`
}

var C Config

func Load() error {
	configDir := filepath.Join(xdg.ConfigHome, "immprune")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return setup(configDir)
		}
		return err
	}
	return viper.Unmarshal(&C)
}

func setup(dir string) error {
	fmt.Println("🔧 Première exécution — configuration immprune")

	prompt := promptui.Prompt{Label: "Immich URL (ex: https://immich.mondomaine.com)"}
	url, _ := prompt.Run()

	keyPrompt := promptui.Prompt{Label: "Clé API Immich (lecture seule)", Mask: '*'}
	key, _ := keyPrompt.Run()

	viper.Set("immich_url", url)
	viper.Set("immich_key", key)
	C.ImmichURL = url
	C.ImmichKey = key

	os.MkdirAll(dir, 0700)
	cfgPath := filepath.Join(dir, "config.yaml")
	viper.WriteConfigAs(cfgPath)
	os.Chmod(cfgPath, 0600)
	fmt.Printf("✅ Config sécurisée dans %s\n", cfgPath)
	return nil
}
