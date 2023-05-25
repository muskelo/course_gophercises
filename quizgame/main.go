package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
)

// creates a sequence of numbers and shuffles
func createShuffledSeq(lenght int) []int {
	nums := make([]int, lenght)
	for i := range nums {
		nums[i] = i
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(lenght, func(i, j int) {
		nums[i], nums[j] = nums[j], nums[i]
	})
	return nums
}

//
// Quiz game
//

type Problem struct {
	Question string
	Answer   string
}

type QuizGame struct {
	Problems []Problem
	Timeout  time.Duration
}

func (game *QuizGame) ParseCSV(fn string) error {
	file, err := os.Open(fn)
	if err != nil {
		return err
	}
	reader := csv.NewReader(file)

	line := 0
	for {
		line++

		record, err := reader.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if len(record) != 2 {
			return fmt.Errorf("Bad format in line %v", line)
		}

		problem := Problem{Question: record[0], Answer: record[1]}
		game.Problems = append(game.Problems, problem)
	}
}

func (game *QuizGame) Run() error {
	// prepare
	userIn := bufio.NewReader(os.Stdin)
	shuffledSeq := createShuffledSeq(len(game.Problems))
	answerCh := make(chan string)
	errorCh := make(chan error)

    // set timeout
	ctxTimeout, cancle := context.WithTimeout(context.TODO(), game.Timeout)
	defer cancle()

	for i, j := range shuffledSeq {
		fmt.Printf("#%v %v: ", i+1, game.Problems[j].Question)

		go func() {
			answer, err := userIn.ReadString('\n')
			if err != nil {
				errorCh <- err
                return
			}
			answer = strings.TrimSpace(answer)
			answerCh <- answer
		}()

		select {
		case answer := <-answerCh:
			if !strings.Contains(answer, game.Problems[j].Answer) {
				fmt.Printf("Game over, wrong answer\n")
				return nil
			}
		case err := <-errorCh:
			return err
		case <-ctxTimeout.Done():
			fmt.Println("\nGame over, time out")
			return nil
		}
	}

	fmt.Printf("You win!\n")
	return nil
}

//
// Main
//

type Flags struct {
	CSV     string
	Help    bool
	Timeout int64
	FS      *flag.FlagSet
}

func ParseFlags() (Flags, error) {
	fs := flag.NewFlagSet("Quiz game", flag.ContinueOnError)
	flags := Flags{FS: fs}

	fs.StringVar(&flags.CSV, "csv", "problems.csv", "CSV file with problems for game")
	fs.BoolVar(&flags.Help, "help", false, "print this message")
	fs.Int64Var(&flags.Timeout, "timeout", 30, "timeout in secconds")

	return flags, fs.Parse(os.Args[1:])
}

func main() {
	flags, err := ParseFlags()
	if err != nil {
		os.Exit(1)
	}
	if flags.Help == true {
		flags.FS.Usage()
		os.Exit(0)
	}

	game := new(QuizGame)
	game.Timeout = time.Duration(flags.Timeout) * time.Second
	if err = game.ParseCSV(flags.CSV); err != nil {
		fmt.Printf("Can't parse csv file: %v\n", err)
		os.Exit(1)
	}

	err = game.Run()
	if err != nil {
		fmt.Printf("Something went wrong: %v\n", err)
		os.Exit(1)
	}
}
