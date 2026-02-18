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
	fmt.Println("🔧 First run — immprune setup")
	fmt.Println("Create an Immich API key with the 'apiKey.read' permission.")
	fmt.Println("In Immich, open: /user-settings?isOpen=api-keys")

	prompt := promptui.Prompt{Label: "Immich URL (example: https://immich.yourdomain.com)"}
	url, _ := prompt.Run()

	keyPrompt := promptui.Prompt{Label: "Immich API key (must include apiKey.read)", Mask: '*'}
	key, _ := keyPrompt.Run()

	viper.Set("immich_url", url)
	viper.Set("immich_key", key)
	C.ImmichURL = url
	C.ImmichKey = key

	os.MkdirAll(dir, 0700)
	cfgPath := filepath.Join(dir, "config.yaml")
	viper.WriteConfigAs(cfgPath)
	os.Chmod(cfgPath, 0600)
	fmt.Printf("✅ Secure config written to %s\n", cfgPath)
	return nil
}
