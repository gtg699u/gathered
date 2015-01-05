package main

import (
	"fmt"
	"strings"
	"strconv"
	"time"
)

var legalMoves [][]bool
var maxValues []uint8
//var inSack []bool
var discount float32
var bestSeq []uint8
var curSeq []uint8

func oneUp(seq []uint8) {
	for i := range seq {
		seq[i]++
	}
}

func oneDown(seq []uint8) {
	for i := range seq {
		seq[i]--
	}
}

func isIn(want uint8, seq []uint8) (bool) {
	for _, v := range seq {
		if v == want {
			return true
		}
	}
	return false
}

func prepGame(s string, v string) {
	length := strings.Count(v,",") + 1 // Number of nodes, since each node has a single point value assigned in v
	legalMoves = make([][]bool, length) // 2D slice where legalMoves[i][j] returns whether an edge allows movement from node (i+1) to node (j+1)
	buildLegalMoves(s)
	maxValues = make([]uint8, length) // Max value each node can reach, both initial value and necessary check each time node value increases
	for i, val := range strings.Split(v,",") {
		 t, _ := strconv.Atoi(val)
		 maxValues[i] = uint8(t)
	}
}

func buildLegalMoves(s string) {
	//Build out 2d array
	for i := range legalMoves {
		legalMoves[i] = make([]bool, len(legalMoves))
	}

	edges := strings.Split(s, ",") //Split input string in to each pair of edges
	indices := make([]string, 2) //The two node indices in each edge pair, originally strings from parsing
	var i, j int //The two node indices converted to ints

	for _, pair := range edges { // Var p only necessary for making sure pairs are being read in correctly, should be _
		if strings.Contains(pair, "-") { // Only base case of two way paths for now, will build out alternative edges
			indices = strings.Split(pair, "-")
			i, _ = strconv.Atoi(indices[0])
			j, _ = strconv.Atoi(indices[1])
			//fmt.Printf("Pair %d:\t%v-%v\n", p, i, j) // Pretty printing optional
			legalMoves[i-1][j-1] = true // Can move from i to j
			legalMoves[j-1][i-1] = true // Can move from j to i
		}
	}
}

func runGame(startNode uint8, carryLimit uint8, moves uint8) (highScore uint32) {
	// Initial conditions
	curScore := uint32(0)
	curTurn := uint8(0)
	curNode := startNode
	carry := make([]uint8, carryLimit)
	curValues := make([]uint8, len(maxValues))
	carryTurnCap := uint8(carryLimit + 2) // Don't expect to dig more than Carry Limit + 2 turns, a model parameter
	var maxDepth uint8
	var newScore uint32

	for i := range carry { // Initialize carry to home
		carry[i] = curNode
	}

	for i := range curValues { // Initialize current values to max values
		curValues[i] = maxValues[i]
	}	

	for curTurn < moves {
		if c := uint8(carryTurnCap + curTurn); c < moves {
			curSeq = make([]uint8, c)
			_ = copy(curSeq, bestSeq[:curTurn])
			oneUp(curSeq)
			fmt.Printf("Iterating from node %d on turn %d with a current score of %d and a endScore of %d following seq %v\n", curNode+1, curTurn, curScore, newScore, curSeq)
			oneDown(curSeq)
			maxDepth = c
			carryCopy := make([]uint8, len(carry))
			curValuesCopy := make([]uint8, len(curValues))
			_ = copy(carryCopy, carry)
			_ = copy(curValuesCopy, curValues)
			newScore = dig(curNode, curScore, carryCopy, curValuesCopy, curTurn, maxDepth, 0) // Find best sequence to specified depth

			//curNode = bestSeq[curTurn] // Advance to next node in "best seq"
			//curTurn++ // Advance turn counter
			// Update play space

			for curScore < newScore { // Loop through returned sequence until we get to next Home pit stop
				curNode = bestSeq[curTurn] // Update current node to perscribed best sequence
				curTurn++ // Advance turn counter// Update Nodes
				for i := range curValues {
					if curValues[i] < maxValues[i] {
						curValues[i]++
					}
				}
				
				// Update Score/Carry
				if maxValues[curNode] == 0 { // If Home, add carried items to score, break loop and start new dig
					for i, v := range carry {
						curScore += uint32(curValues[v])
						curValues[v] = 0 // Reset node to 0
						carry[i] = curNode // Fill item with 0 value nodes (Home is always 0 )
					}
					break
				} else if !isIn(curNode, carry) { // Else, check to see if new node is already in carry, and if not, if it should be swapped in to carry
					minCarryIndex := uint8(0)
					minCarryValue := curValues[carry[0]]
					for i := uint8(1); i < uint8(len(carry)); i++ {
						if curValues[carry[i]] < minCarryValue {
							minCarryValue = curValues[carry[i]]
							minCarryIndex = i
						}
					}
					if curValues[curNode] > minCarryValue {
						carry[minCarryIndex] = curNode
					}
				}
			}

		} else {
			curSeq = make([]uint8, moves)
			_ = copy(curSeq, bestSeq[:curTurn])
			oneUp(bestSeq)
			fmt.Printf("Last iteration from node %d on turn %d with a current score of %d and a endScore of %d following seq %v\n", curNode+1, curTurn, curScore, newScore, bestSeq)
			oneDown(bestSeq)
			maxDepth = moves
			discount = 0.0
			carryCopy := make([]uint8, len(carry))
			curValuesCopy := make([]uint8, len(curValues))
			_ = copy(carryCopy, carry)
			_ = copy(curValuesCopy, curValues)
			return dig(curNode, curScore, carryCopy, curValuesCopy, curTurn, maxDepth, 0) // At end of the line so no need to update play space
		}
	}
	return 0 // Allah willing we don't end up here
}

func dig(curNode uint8, curScore uint32, carry []uint8, curValues []uint8, depth uint8, maxDepth uint8, bestScore uint32) (endScore uint32) {
	// Update 
	for i := range curValues {
		if curValues[i] < maxValues[i] {
			curValues[i]++
		}
	}
	
	// Update Score/Carry
	if maxValues[curNode] == 0 { // If Home, add carried items to score
		/*oneUp(carry)
		oneUp(curSeq)
		fmt.Printf("We're home after %v, old score was %d, carrying %v\n", curSeq, curScore, carry)
		oneDown(carry)
		oneDown(curSeq)*/
		for i, v := range carry {
			curScore += uint32(curValues[v])
			curValues[v] = 0 // Reset node to 0
			carry[i] = curNode // Fill item with 0 value nodes (Home is always 0 )
			/*oneUp(carry)
			fmt.Printf("Score now %d, carrying %v\n", curScore, carry)
			oneDown(carry)*/
		}
	} else if !isIn(curNode, carry) { // Else, check to see if new node is already in carry, and if not, if it should be swapped in to carry
		minCarryIndex := uint8(0)
		minCarryValue := curValues[carry[0]]

		for i := uint8(1); i < uint8(len(carry)); i++ {
			if curValues[carry[i]] < minCarryValue {
				minCarryValue = curValues[carry[i]]
				minCarryIndex = i
			}
		}
		if curValues[curNode] > minCarryValue {
			carry[minCarryIndex] = curNode
		}
	}

	if depth == maxDepth { // End of the line, find Score and return sequence/score
		endScore = curScore
		curSeq[depth-1] = curNode
		//fmt.Printf("End of the line on node %d at depth %d\n", curNode, depth)
		if maxValues[curNode] == 0 { // If Home, already added carried items to score, just return score
			/*oneUp(curSeq)
			fmt.Printf("We're home, curScore is %d, endScore is %d from seq %v\n", curScore, endScore, curSeq)
			oneDown(curSeq)*/
			return
		} else { // Take current score and discount carry items 
			carVal := float32(0)
			for _, v := range carry {
				carVal += float32(curValues[v]) * discount // Otherwise, rounding happens items by item instead of as a set
			}
			endScore += uint32(carVal)
			return
		}
	} else { // Loop through legal moves and return best sequence/score
		depth++ // Going one level deeper
		for i, isLegal := range legalMoves[curNode] {
			if isLegal { //if a legal move exists
				carryCopy := make([]uint8, len(carry))
				curValuesCopy := make([]uint8, len(curValues))
				_ = copy(carryCopy, carry)
				_ = copy(curValuesCopy, curValues)
				//fmt.Printf("Checking node %d at depth %d\n", i, depth)
				curSeq[depth-1] = uint8(i)
				thisScore := dig(uint8(i), curScore, carryCopy, curValuesCopy, depth, maxDepth, bestScore) 
				if thisScore > bestScore { // If this sequence scores higher, replace sequence/score
					bestScore = thisScore
					bestSeq[depth-1] = uint8(i) // By convention, the first move at depth 1 goes in to the 0th slot
				}
			}
		}
		return bestScore
	}
}

func main() {
	start := time.Now()	
	//Need to read these in via script
	//s := "1-2,1-3,1-4,1-6,2-3,3-4,3-5,4-5,4-6" // level 1
	s := "1-2,1-3,1-4,2-3,2-4,2-7,3-5,3-6,3-7,4-5,4-6" // level 2
	//v :=  "5,7,6,0,5,7" // level 1
	v :=  "0,6,7,7,6,6,8" // level 2
	carryLimit := uint8(3)
	moves := uint8(20)
	bestSeq = make([]uint8, moves)
	discount = .5 // Model paramter, how much to discount carried items by versus items which actually score

	prepGame(s, v)
	highScore := runGame(0, carryLimit, moves)
	
	oneUp(bestSeq)
	fmt.Printf("High Score - %d\nFrom Seq - %v\n", highScore, bestSeq)
	oneDown(bestSeq)
	fmt.Println(time.Since(start))
}