package main

import (
	"fmt"
	"kasir-api/database"
	"os"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port   string `mapstructure:"PORT"`
	DBConn string `mapstructure:"DB_CONN"`
}

func main() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		viper.SetConfigType("env") // ← Tambahkan ini
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Error reading config: %v", err)
		}
	}
	
	config := Config{
		Port:   viper.GetString("PORT"),
		DBConn: viper.GetString("DB_CONN"),
	}

	// Setup database
	db, err := database.InitDB(config.DBConn)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	fmt.Println("✅ Database connected successfully!")
	fmt.Println("Server running di localhost:" + config.Port)
}