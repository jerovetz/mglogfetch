package utils

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

func rootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}

func InitTestEnv() {
	dotenvErr := godotenv.Load(rootDir() + "/.env.test")
	if dotenvErr != nil {
		fmt.Println("Error loading test env")
		os.Exit(2)
	}
}
