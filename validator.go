package main

import (
	"github.com/eoscanada/validator"
	"github.com/thedevsaddam/govalidator"
)

func configureValidators() {
	govalidator.AddCustomRule("eos.blockNum", validator.EOSBlockNumRule)
}
