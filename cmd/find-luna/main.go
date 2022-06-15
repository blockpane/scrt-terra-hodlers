package main

import (
	"flag"
	"fmt"
	lunatics "github.com/blockpane/scrt-terra-hodlers"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.Lshortfile)
	log.SetOutput(os.Stderr)

	var block int64
	var workers int
	var endpoint, outfile string
	flag.StringVar(&endpoint, "n", "scrt-rpc.blockpane.com:9090", "grpc endpoint")
	flag.StringVar(&outfile, "o", "<block-number>", ".csv file basename to write, defaults to the block number, will append type UST|LUNA to name")
	flag.Int64Var(&block, "b", 3286613, "block to index.")
	flag.IntVar(&workers, "w", 4, "number of simultaneous workers pulling account balances. Warning: too many will overload server.")
	flag.Parse()

	var fatal string
	switch true {
	case endpoint == "":
		fatal = "grpc endpoint cannot be empty"
	case block == 0:
		fatal = "block must be > 0"
	case workers == 0:
		fatal = "workers must be > 0"
	}
	if fatal != "" {
		flag.PrintDefaults()
		log.Fatal(fatal)
	}

	var outUST, outLUNA string
	if outfile == "<block-number>" {
		outUST = fmt.Sprintf("%d-UST.csv", block)
		outLUNA = fmt.Sprintf("%d-LUNA.csv", block)
	}
	fU, err := os.OpenFile(outUST, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fL, err := os.OpenFile(outLUNA, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	lunatics.Run(endpoint, block, workers, fU, fL)
}
