package config
  
import (
        "fmt"
        "os"
	"log"
        "github.com/spf13/viper"
)

func Run() (string, string) {
	type Profile struct {
	        PrivateKey string `mapstructure:"privateKey"`
		PublicKey string `mapstructure:"publicKey"`
	}

	var p map[string]Profile

        homeDir, err := os.UserHomeDir()
        if err != nil {
                log.Fatal(err)
        }

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(homeDir + "/.mongoex/")

	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("couldn't load config: %s", err)
		os.Exit(1)
	}
	if err := v.Unmarshal(&p); err != nil {
		fmt.Printf("couldn't read config: %s", err)
	}

	mongoex := os.Getenv("ATLAS_ORG")
	return p[mongoex].PublicKey,  p[mongoex].PrivateKey
}
