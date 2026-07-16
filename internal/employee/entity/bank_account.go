package entity

import "fmt"

type BankAccount struct {
	holder string
	name   string
	number string
}

func NewBankAccount(holder, name, number string) (BankAccount, error) {
	if holder == "" {
		return BankAccount{}, fmt.Errorf("bank holder is required")
	}
	if name == "" {
		return BankAccount{}, fmt.Errorf("bank name is required")
	}
	if number == "" {
		return BankAccount{}, fmt.Errorf("bank number is required")
	}
	return BankAccount{holder: holder, name: name, number: number}, nil
}

func BankAccountFromDB(holder, name, number string) BankAccount {
	return BankAccount{holder: holder, name: name, number: number}
}

func (b BankAccount) Holder() string  { return b.holder }
func (b BankAccount) Name() string    { return b.name }
func (b BankAccount) Number() string  { return b.number }

func (b BankAccount) IsEmpty() bool {
	return b.holder == "" && b.name == "" && b.number == ""
}
