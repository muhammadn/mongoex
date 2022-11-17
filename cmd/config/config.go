package config
  
import (
        "fmt"
        "os"
	"log"
        "github.com/spf13/viper"
)

func ParseConfig() (string, string) {
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
		fmt.Printf("\ncouldn't load config: %s\n", err)
                fmt.Printf("We'll try to check ATLAS_PUBLICKEY and ATLAS_PRIVATEKEY env vars\n")
	}
	if err := v.Unmarshal(&p); err != nil {
		fmt.Printf("couldn't read config: %s\n", err)
	}

	atlasOrg := os.Getenv("ATLAS_ORG")
        if atlasOrg == "" {
                publicKey  := os.Getenv("ATLAS_PUBLICKEY")
                privateKey := os.Getenv("ATLAS_PRIVATEKEY")
                return publicKey, privateKey
        }
	return p[atlasOrg].PublicKey,  p[atlasOrg].PrivateKey
}
