package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chrismar035/sudoku-solver"
)

type removedSquare struct {
	index int
	value int
}

type postParams struct {
	Puzzle   solver.Grid `json:"puzzle"`
	Solution solver.Grid `json:"solution"`
}

type Streaker struct {
	count int
	which string
}

func (s *Streaker) Count(which string) {
	if s.count == 0 {
		s.which = which
		s.count = 1
		return
	}

	if s.which == which {
		s.count++
		if s.count%5 == 0 {
			message := "{\"text\": \"Currently in a streak of " + strconv.Itoa(s.count) + " " + s.which + "\"}"
			postToSlack(message)
		}
	} else {
		message := "{\"text\": \"Broke a streak of " + strconv.Itoa(s.count) + " " + s.which + "\"}"
		postToSlack(message)

		s.which = which
		s.count = 1
	}
}

func main() {
	url := os.Getenv("API_ROOT")
	logger := log.New(os.Stdout,
		"Generator: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	streaker := Streaker{}

	logger.Println("Starting loop")
	for {
		solution := getShuffledSolution()
		puzzle, err := puzzleFromSolution(solution)
		if err != nil {
			logger.Println("Error generating puzzle", solution, err)
		} else {
			params := postParams{Puzzle: puzzle, Solution: solution}
			jsonStr, err := json.Marshal(params)
			if err != nil {
				logger.Println("Unable to marshal puzzle", puzzle, solution)
				continue
			}

			req, err := http.NewRequest("POST", url+"/puzzle", bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			_, err = client.Do(req)
			if err != nil {
				logger.Println("Unable to submit puzzle", err)
				streaker.Count("duplicates")
				continue
			} else {
				streaker.Count("adds")
			}
		}
		logger.Println("Iteration")
	}
	logger.Println("Out of loop. Ending.")
}

func postToSlack(message string) {
	url := "https://hooks.slack.com/services/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"
	req, _ := http.NewRequest("POST", url, strings.NewReader(message))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	_, _ = client.Do(req)
}

func puzzleFromSolution(solution solver.Grid) (solver.Grid, error) {
	puzzle := solution
	indexes := randomizeIndexes()
	var removed []removedSquare

	multiSolver := solver.NewMultiBacktrackingSolver()

	for _, index := range indexes {
		removed = append(removed, removedSquare{index: index, value: puzzle[index]})
		puzzle[index] = 0

		if len(multiSolver.Solve(puzzle)) > 1 {
			last := removed[len(removed)-1]
			puzzle[last.index] = last.value

			return puzzle, nil
		}
	}
	return solver.Grid{}, errors.New("Couldn't find puzzle")
}

func getShuffledSolution() solver.Grid {
	var grid solver.Grid
	randomizer := solver.NewRandBacktrackingSolver()

	return randomizer.Solve(grid)
}

func randomizeIndexes() []int {
	rand.Seed(time.Now().UTC().UnixNano())

	ints := []int{}
	for i := 0; i < 81; i++ {
		ints = append(ints, i)
	}

	mixed := []int{}
	for len(ints) > 0 {
		i := rand.Int() % len(ints)
		mixed = append(mixed, ints[i])
		ints = append(ints[0:i], ints[i+1:]...)
	}

	return mixed
}
