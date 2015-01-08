package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"time"
)

type Level struct {
	Edges      string
	MaxValues  string
	CarryLimit uint8
	Moves      uint8
}

func (l Level) String() string {
	return fmt.Sprintf("Edges : %s\nMaxValues : %s\nCarry Limit : %d\nMoves : %d\n", l.Edges, l.MaxValues, l.CarryLimit, l.Moves)
}

func main() {
	re := regexp.MustCompile(`(?sU)levels\[(\d+)\] = "(.+)\\n.+"(.+)\\n.+"(\d)\\n.+"(\d+)\\n`)
	/*Format of levels from JS, see below for example - Flags s (. matches \n) and U (ungreedy)
		levels[0] = "1-2,1-3,1-4,1-6,2-3,3-4,3-5,4-5,4-6\n" +
	                "5,7,6,0,5,7\n" +
	                "2\n" +
	                "15\n" +*/
	start := time.Now()
	var levels []*Level
	if levelsrc, err := ioutil.ReadFile("gathered.html"); err == nil {
		levelsIn := re.FindAllSubmatch(levelsrc, -1)
		levels = make([]*Level, len(levelsIn))
		for i, l := range levelsIn {
			levels[i] = &Level{Edges: string(l[2]), MaxValues: string(l[3]), CarryLimit: uint8(l[4][0] - byte('0')), Moves: uint8(10*(l[5][0]-byte('0')) + l[5][1] - byte('0'))}
		}
	} else {
		fmt.Printf("Error reading file : %v\n", err)
	}

	for i := range levels {
		fmt.Printf("Levels %d\n%s\n", i, levels[i])
	}
	fmt.Println(time.Since(start))
}
