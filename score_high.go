package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var legalMoves [][]bool
var minScores [][]uint16
var maxScores [][]uint16
var maxValues []uint8
var discount float32
var bestSeq []uint8

func oneUp(seq []uint8) {
	for i := range seq {
		seq[i]++
	}
}

/*func oneDown(seq []uint8) {
	for i := range seq {
		seq[i]--
	}
}*/

func min(a, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}

func isIn(want uint8, seq []uint8) bool {
	for _, v := range seq {
		if v == want {
			return true
		}
	}
	return false
}

func index(want uint8, seq []uint8) uint8 {
	for i, v := range seq {
		if v == want {
			return uint8(i)
		}
	}
	return uint8(len(seq))
}

func emptySlots(seq []uint8) (numEmpty uint8) {
	for _, v := range seq {
		if maxValues[v] == 0 {
			numEmpty++
		}
	}
	return
}

type Level struct {
	Edges      string
	MaxValues  string
	CarryLimit uint8
	Moves      uint8
}

/*func (l Level) String() string { //Pretty print level details to debug reading issues
	return fmt.Sprintf("Edges : %s\nMaxValues : %s\nCarry Limit : %d\nMoves : %d\n", l.Edges, l.MaxValues, l.CarryLimit, l.Moves)
}*/

func prepGame(s string, v string) {
	length := strings.Count(v, ",") + 1  // Number of nodes, since each node has a single point value assigned in v
	legalMoves = make([][]bool, length)  // 2D slice where legalMoves[i][j] returns whether an edge allows movement from node (i+1) to node (j+1)
	minScores = make([][]uint16, length) // 2D slice where minScore[i][j] returns the min current score for the edge (i+1) to (j+1) to be active
	maxScores = make([][]uint16, length) // 2D slice where maxScore[i][j] returns the max current score for the edge (i+1) to (j+1) to be active

	buildLegalMoves(s)

	maxValues = make([]uint8, length) // Max value each node can reach, both initial value and necessary check each time node value increases
	for i, val := range strings.Split(v, ",") {
		t, _ := strconv.Atoi(val)
		maxValues[i] = uint8(t)
	}
}

func buildLegalMoves(s string) {
	//Build out 2d arrays, all the same size
	for i := range legalMoves {
		legalMoves[i] = make([]bool, len(legalMoves))  // Defaults to false so fine without initialization
		minScores[i] = make([]uint16, len(legalMoves)) // Defaults to 0 so fine without initialization
		maxScores[i] = make([]uint16, len(legalMoves)) // Defaults to 0 so need to switch to max uint16 vak
		for j := range maxScores {
			maxScores[i][j] = ^uint16(0) // Flips 0 bits to 1s
		}
	}

	edges := strings.Split(s, ",") //Split input string in to each pair of edges
	indices := make([]string, 2)   //The two node indices in each edge pair, originally strings from parsing
	var i, j, k int                //The two node indices and conditional parameter converted to ints

	for _, edge := range edges {
		conditionals := strings.Split(edge, "|")
		pair := conditionals[0]

		if strings.Contains(pair, "->") { // Directed edge from i to j but not in reverse
			indices = strings.Split(pair, "->")
			i, _ = strconv.Atoi(indices[0])
			j, _ = strconv.Atoi(indices[1])
			legalMoves[i-1][j-1] = true // Can move from i to j
		} else if strings.Contains(pair, "-") { // Undirected edge allowing movemenet from i to j and j to i
			indices = strings.Split(pair, "-")
			i, _ = strconv.Atoi(indices[0])
			j, _ = strconv.Atoi(indices[1])
			legalMoves[i-1][j-1] = true // Can move from i to j
			legalMoves[j-1][i-1] = true // Can move from j to i
		}

		if len(conditionals) > 1 { // If there's a conditional piece
			k, _ = strconv.Atoi(conditionals[1][1:])
			if string(conditionals[1][0]) == ">" {
				minScores[i-1][j-1] = uint16(k)
				minScores[j-1][i-1] = uint16(k)
			} else {
				maxScores[i-1][j-1] = uint16(k)
				maxScores[j-1][i-1] = uint16(k)
			}
		}
	}
}

func runGame(startNode uint8, carryLimit uint8, moves uint8, param uint8) (highScore uint16) {
	// Initial conditions
	curScore := uint16(0)
	curTurn := uint8(0)
	curNode := startNode
	carry := make([]uint8, carryLimit)
	curValues := make([]uint8, len(maxValues))
	carryTurnCap := uint8((carryLimit + 1) + param) // How many moves to look forward, a model parameter
	bestSeq = make([]uint8, moves)
	var maxDepth uint8
	//var newScore uint16

	for i := range carry { // Initialize carry to home
		carry[i] = curNode
	}

	for i := range curValues { // Initialize current values to max values
		curValues[i] = maxValues[i]
	}

	for curTurn < moves {
		if c := uint8(carryTurnCap + curTurn); c < moves { // If look ahead depth < number of moves left, use look ahead depth
			maxDepth = c
			carryCopy := make([]uint8, len(carry)) // Need to create copies to keep state spaces separate
			curValuesCopy := make([]uint8, len(curValues))
			_ = copy(carryCopy, carry)
			_ = copy(curValuesCopy, curValues)
			_ = dig(curNode, curScore, carryCopy, curValuesCopy, curTurn, maxDepth, 0) // Find best sequence to specified depth
			// Update play space

			for curTurn < c { // Loop through returned sequence until we get to next Home pit stop
				curNode = bestSeq[curTurn] // Update current node to perscribed best move
				curTurn++                  // Advance turn counter
				// Update Nodes
				for i := range curValues {
					if curValues[i] < maxValues[i] {
						curValues[i]++
					}
				}

				// Update Score/Carry
				if maxValues[curNode] == 0 { // If Home, add carried items to score, break loop and start new dig
					for i, v := range carry {
						curScore += uint16(curValues[v])
						curValues[v] = 0   // Reset node to 0
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

		} else { // If look ahead depth takes us up to or beyond the last move, return results
			maxDepth = moves
			discount = 0.0 // Discount always 0 on final stretch
			carryCopy := make([]uint8, len(carry))
			curValuesCopy := make([]uint8, len(curValues))
			_ = copy(carryCopy, carry)
			_ = copy(curValuesCopy, curValues)
			return dig(curNode, curScore, carryCopy, curValuesCopy, curTurn, maxDepth, 0) // At end of the line so no need to update play space
		}
	}
	return 0 // Allah willing we don't end up here
}

func dig(curNode uint8, curScore uint16, carry []uint8, curValues []uint8, depth uint8, maxDepth uint8, bestScore uint16) (endScore uint16) {
	// Update Score/Carry
	if maxValues[curNode] == 0 { // If Home, add carried items to score
		for i, v := range carry {
			curScore += uint16(curValues[v])
			curValues[v] = 0   // Reset node to 0
			carry[i] = curNode // Fill item with 0 value nodes (Home is always 0 )
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
		if maxValues[curNode] == 0 { // If Home, already added carried items to score, just return score
			return
		} else { // Take current score and discount carry items
			carVal := float32(0)
			numEmpty := emptySlots(carry) //Number of empty spaces in carry, good proxy for how many moves until return home
			for _, v := range carry {
				carVal += float32(min(curValues[v]+numEmpty+1, maxValues[v])) * discount // Otherwise, rounding happens items by item instead of as a set
			}
			endScore += uint16(carVal)
			return
		}
	} else { // Loop through legal moves and return best sequence/score
		depth++ // Going one level deeper
		// Update node values
		for i := range curValues {
			if curValues[i] < maxValues[i] {
				curValues[i]++
			}
		}

		for i, isLegal := range legalMoves[curNode] {
			if isLegal && curScore >= minScores[curNode][i] && curScore <= maxScores[curNode][i] { //if a legal move exists
				carryCopy := make([]uint8, len(carry))
				curValuesCopy := make([]uint8, len(curValues))
				_ = copy(carryCopy, carry)
				_ = copy(curValuesCopy, curValues)
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
	re := regexp.MustCompile(`(?sU)levels\[(\d+)\] = "(.+)\\n.+"(.+)\\n.+"(\d)\\n.+"(\d+)\\n`)
	/*Format of levels from JS, see below for example - Flags s (. matches \n) and U (ungreedy)
		levels[0] = "1-2,1-3,1-4,1-6,2-3,3-4,3-5,4-5,4-6\n" +
	                "5,7,6,0,5,7\n" +
	                "2\n" +
	                "15\n" +*/
	var levels []*Level // Below loop reads through regex results and builds level structs, collects pointers for following loop
	if levelsrc, err := ioutil.ReadFile("gathered.html"); err == nil {
		levelsIn := re.FindAllSubmatch(levelsrc, -1)
		levels = make([]*Level, len(levelsIn))
		for i, l := range levelsIn {
			levels[i] = &Level{Edges: string(l[2]), MaxValues: string(l[3]), CarryLimit: uint8(l[4][0] - byte('0')), Moves: uint8(10*(l[5][0]-byte('0')) + l[5][1] - byte('0'))}
		}
	} else {
		fmt.Printf("Error reading file : %v\n", err)
		return
	}

	for i := range levels {
		carryLimit := levels[i].CarryLimit
		moves := levels[i].Moves
		//s := levels[i].Edges
		//v := levels[i].MaxValues

		prepGame(levels[i].Edges, levels[i].MaxValues)
		stepParams := []uint8{1, 2, 3}
		//stepParams := []uint8{0, 1, 2, 3}
		discounts := []float32{.5, .66, .75}
		//discounts := []float32{.25, .33, .5, .66, .75}
		highScore := uint16(0)
		retSeq := make([]uint8, moves)
		var p, pLast uint8
		var d, dLast float32
		// Cycle through parameters
		for _, disc := range discounts {
			for _, stepParam := range stepParams {
				discount = disc // Model paramter, how much to discount carried items by versus items which actually score
				paramScore := runGame(index(0, maxValues), carryLimit, moves, stepParam)
				if paramScore > highScore {
					highScore = paramScore
					p, pLast = stepParam, stepParam // Keep track of first and last param pair so we can pare down params we check
					d, dLast = disc, disc
					_ = copy(retSeq, bestSeq)
				} else if paramScore == highScore {
					pLast = stepParam
					dLast = disc
				}
			}
		}
		oneUp(retSeq)
		fmt.Printf("Level %d\tHigh Score - %d\nFirst p-%d, d-%f\tLast p-%d, d-%f\nFrom Seq - %v\n", i+1, highScore, p, d, pLast, dLast, retSeq)
		//oneDown(retSeq)
	}
	fmt.Println(time.Since(start))
}
