package service

import (
	"github.com/lewvy/gopk/config"
)

func Reset() error {
	return config.ResetDB()

}
