package lunatics

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"sync"
	"time"
)

func Run(endpoint string, block int64, workers int, outfileUst *os.File, outfileLuna *os.File) {
	client, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()
	defer outfileUst.Close()
	defer outfileLuna.Close()

	ctx, cancel := context.WithCancel(context.Background())
	doneChan := make(chan interface{})
	accountChan := make(chan string, 1) // stream of accounts to check

	ustBalance := make(chan string)  // stream of accounts holding $UST
	lunaBalance := make(chan string) // stream of accounts holding $LUNA

	// handle writing results to .csv files
	go func() {
		for {
			select {
			case ust := <-ustBalance:
				_, _ = outfileUst.WriteString(ust)
			case luna := <-lunaBalance:
				_, _ = outfileLuna.WriteString(luna)
			case <-doneChan:
				return
			}
		}
	}()

	wg := &sync.WaitGroup{}
	wg.Add(workers)

	// lookup each account's balance and send positive matches for luna or ust to
	// the correct channel to be written as a csv line
	go getBalances(ctx, wg, client, workers, block, accountChan, lunaBalance, ustBalance)

	// query all existing accounts and send to the getBalances function via the
	// accountChan channel. BLocks until finished, the limited channel buffer should
	// ensure all accounts are reported.
	findAccounts(ctx, client, accountChan)

	// notify workers there are no more results
	cancel()
	// wait for workers to finish
	wg.Wait()
	// inform worker writing files we are done
	close(doneChan)

	// give our routines a little extra time to finish writing to .csv files.
	time.Sleep(5 * time.Second)
}
