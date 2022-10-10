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

	err = os.MkdirAll(homeDir + "/.mongoex", 0644)
	if err != nil {
                fmt.Println(err)
	}
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
}
