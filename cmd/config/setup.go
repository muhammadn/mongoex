package config

import (
	"fmt"
	"os"
)

func SetupCreate() {
        homeDir, err := os.UserHomeDir()
        if err != nil {
                fmt.Println(err)
        }

	err = os.MkdirAll(homeDir + "/.mongoex", 0755)
	if err != nil {
                fmt.Println(err)
	}
  
        if _, err := os.Stat(homeDir + "/.mongoex/config"); err != nil {
                // If the file does not exist, create it
                fmt.Println("config does not exist, creating...")
                file, err := os.Create(homeDir + "/.mongoex/config")
                if err != nil {
	                fmt.Println(err)
                } else {
	                file.WriteString("[default]\n")
	                file.WriteString("publicKey = \"yourmongodbpublickey\"\n")
	                file.WriteString("privateKey = \"yourmongodbprivatekey\"\n")
	                fmt.Println("Done creating.")
	                fmt.Println("Configuration is at ~/.mongoex/config")
                }
                file.Close()
	} else {
                // do not create if file already exist to avoid overriding a live configuration
                fmt.Println("unable to create configuration file as ~/.mongoex/config already exists")
        }
}
