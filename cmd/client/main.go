package main

import (
	"fmt"
	"os"
	"strings"
	"github.com/c-bata/go-prompt"
)

var LivePrefixState struct {
	LivePrefix string
	IsEnable   bool
}

func executor(in string) {

	in = strings.Trim(in, " ")

	if in == "" {
		// LivePrefixState.IsEnable = false
		// LivePrefixState.LivePrefix = in
		return
	}
	if in == "exit"{
		os.Exit(0)
	}

	// LivePrefixState.LivePrefix = in + "> "
	// LivePrefixState.IsEnable = true
	args := strings.Fields(in)	

	fun, ok := m[args[0]]
	if !ok {
		fmt.Printf("command %s is not exist!\n", in)
		return
	}
	switch fun.(type) {
		case func():
			fun.(func())()
		case func(string)string:
			param := ""
			if len(args) > 1{
				param = args[1]
			}
			if args[0] == "use"{
				accountID := fun.(func(string)string)(param)
				if len(accountID) != 0{
					LivePrefixState.LivePrefix = accountID + "> "
					LivePrefixState.IsEnable = true
				}
			}
	}

}

func completer(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "accounts", Description: "show all accounts info in the wallet"},
		{Text: "login", Description: "login"},
		{Text: "use", Description: "choose a account to do operation"},
		{Text: "create", Description: "create a new contract"},
	}
	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

func changeLivePrefix() (string, bool) {
	return LivePrefixState.LivePrefix, LivePrefixState.IsEnable
}



func main() {
	Init()
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(">>> "),
		prompt.OptionLivePrefix(changeLivePrefix),
		prompt.OptionTitle("live-prefix-example"),
	)
	p.Run()
}

