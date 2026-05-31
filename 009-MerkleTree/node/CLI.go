package node

import (
	"flag"
	"fmt"
	"learn_DumboNG/009-MerkleTree/crypto"
	"log"
	"os"
)

type CLI struct{}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("\taddresslists -- 输出所有钱包地址列表")
	fmt.Println("\tcreatewallet -- 创建新钱包")
	fmt.Println("\tcreateblockchain -address -- 交易数据")
	fmt.Println("\tsend -from FROM -to TO -amount AMOUNT -- 交易明细")
	fmt.Println("\tprintchain -- 输出区块信息")
	fmt.Println("\tgetbalance -address -- 查询余额地址")

}

func isValidArgs() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	isValidArgs()

	addressListsCmd := flag.NewFlagSet("addresslists", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	sendBlockCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getbalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)

	flagFrom := sendBlockCmd.String("from", "", "转账来源地址")
	flagTo := sendBlockCmd.String("to", "", "转账目的地地址")
	flagAmount := sendBlockCmd.String("amount", "", "转账金额")

	flagCreateBlockchainWithAddress := createBlockchainCmd.String("address", "Genesis data.......", "创世区块地址")
	getbalanceWithAddress := getbalanceCmd.String("address", "", "查询余额地址")

	switch os.Args[1] {
	case "send":
		err := sendBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getbalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "addresslists":
		err := addressListsCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	default:
		printUsage()
		os.Exit(1)
	}

	if sendBlockCmd.Parsed() {
		if *flagFrom == "" || *flagTo == "" || *flagAmount == "" {
			printUsage()
			os.Exit(1)
		}

		from := crypto.JSONToArray(*flagFrom)
		to := crypto.JSONToArray(*flagTo)
		for index, fromAddress := range from {
			if crypto.IsValidForAddress([]byte(fromAddress)) == false || crypto.IsValidForAddress([]byte(to[index])) == false {
				fmt.Println("地址无效.....")
				printUsage()
				os.Exit(1)
			}
		}
		amount := crypto.JSONToArray(*flagAmount)
		cli.send(from, to, amount)
	}

	if printChainCmd.Parsed() {
		//fmt.Println("输出所有区块的数据...")
		cli.printchain()
	}

	if createWalletCmd.Parsed() {
		// 创建钱包
		cli.CreateWallet()
	}

	if addressListsCmd.Parsed() {
		// 输出所有钱包地址列表
		cli.addressLists()
	}

	if createBlockchainCmd.Parsed() {
		if crypto.IsValidForAddress([]byte(*flagCreateBlockchainWithAddress)) == false {
			fmt.Println("地址无效...")
			printUsage()
			os.Exit(1)
		}
		cli.createGenesisBlockchain(*flagCreateBlockchainWithAddress)
	}

	if getbalanceCmd.Parsed() {
		if crypto.IsValidForAddress([]byte(*getbalanceWithAddress)) == false {
			fmt.Println("地址无效...")
			printUsage()
			os.Exit(1)
		}
		cli.getBalance(*getbalanceWithAddress)
	}
}
