package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Static level constraints and discount/bestSeq which would otherewise be passed through each recursive dig
var (
	edgeConstraints[][]EdgeConstraint
	maxValues  []uint8
	discount   float32
	bestSeq    []uint8
)

// Defining level contraints read in by readLevels()
type Level struct {
	Edges      string
	MaxValues  string
	CarryLimit uint8
	Moves      uint8
}

type EdgeConstraint struct {
	IsLegal bool // Whether edge exists
	MinScore uint16 // Min Score for edge to be active
	MaxScore uint16 // Max Score for edge to be active
}

// Reads level data in from specified file and creates slice of pointers to structs for each level read in
func readLevels(fileName string) (levels []*Level) {
	re := regexp.MustCompile(`(?sU)levels\[(\d+)\] = "(.+)\\n.+"(.+)\\n.+"(\d)\\n.+"(\d+)\\n`)
	/*Format of levels from JS, see below for example - Flags s (. matches \n) and U (ungreedy)
		levels[0] = "1-2,1-3,1-4,1-6,2-3,3-4,3-5,4-5,4-6\n" +
	                "5,7,6,0,5,7\n" +
	                "2\n" +
	                "15\n" +*/
	if levelsrc, err := ioutil.ReadFile(fileName); err == nil {
		levelsIn := re.FindAllSubmatch(levelsrc, -1)
		levels = make([]*Level, len(levelsIn))
		for i, l := range levelsIn {
			levels[i] = &Level{Edges: string(l[2]), MaxValues: string(l[3]), CarryLimit: uint8(l[4][0] - '0'), Moves: uint8(10*(l[5][0]-'0') + l[5][1] - '0')}
		}
	} else {
		fmt.Printf("Error reading file : %v\n", err)
		return
	}
	return levels
}

// Creates the various static constraints for each level which are all set as global variables
func prepGame(s string, v string) {
	length := strings.Count(v, ",") + 1  // Number of nodes, since each node has a single point value assigned in v
	edgeConstraints = make([][]EdgeConstraint, length)
	buildLegalMoves(s)

	maxValues = make([]uint8, length) // Max value each node can reach, both initial value and necessary check each time node value increases
	for i, val := range strings.Split(v, ",") {
		t, _ := strconv.Atoi(val)
		maxValues[i] = uint8(t)
	}
}

// Takes in string of edge pairs and produces 3 2d Slices which are used to check for valid moves
func buildLegalMoves(s string) {
	//Build out 2d arrays, all the same size
	for i := range edgeConstraints {
		edgeConstraints[i] = make([]EdgeConstraint, len(edgeConstraints))
		for j := range edgeConstraints[i] {
			edgeConstraints[i][j].MaxScore = ^uint16(0) // Flips 0 bits to 1s, i.e. to max uint16 val
		}
	}

	edges := strings.Split(s, ",") // Split input string in to each pair of edges
	indices := make([]string, 2)   // The two node indices in each edge pair, originally strings from parsing
	var i, j, k int                // The two node indices and conditional parameter converted to ints

	for _, edge := range edges {
		conditionals := strings.Split(edge, "|")
		pair := conditionals[0]

		if strings.Contains(pair, "->") { // Directed edge from i to j but not in reverse
			indices = strings.Split(pair, "->")
			i, _ = strconv.Atoi(indices[0])
			j, _ = strconv.Atoi(indices[1])
			edgeConstraints[i-1][j-1].IsLegal = true // Can move from i to j
		} else if strings.Contains(pair, "-") { // Undirected edge allowing movemenet from i to j and j to i
			indices = strings.Split(pair, "-")
			i, _ = strconv.Atoi(indices[0])
			j, _ = strconv.Atoi(indices[1])
			edgeConstraints[i-1][j-1].IsLegal = true // Can move from i to j
			edgeConstraints[j-1][i-1].IsLegal = true // Can move from j to i
		}

		if len(conditionals) > 1 { // If there's a conditional piece
			k, _ = strconv.Atoi(conditionals[1][1:])
			if string(conditionals[1][0]) == ">" {
				edgeConstraints[i-1][j-1].MinScore = uint16(k)
				edgeConstraints[j-1][i-1].MinScore = uint16(k)
			} else {
				edgeConstraints[i-1][j-1].MaxScore = uint16(k)
				edgeConstraints[j-1][i-1].MaxScore = uint16(k)
			}
		}
	}
}

/* Runs through game by alternatively digging for the set of best moves and updating the play space to reflect the moves taken,
returns best score found (dig() updates global variable bestSeq)*/
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

	for i, v := range maxValues { // Initialize current values to max values
		curValues[i] = v
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
				for i, v := range curValues {
					if v < maxValues[i] {
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
					for i, v := range carry[1:] {
						if curValues[v] < minCarryValue {
							minCarryValue = curValues[v]
							minCarryIndex = uint8(i+1) // Note loops over indices 1-N
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

/* Recursively digs through potential next moves down to a specified depth and heuristically scores each sequence of moves,
returns the best score (and updating global variable bestSeq)*/
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

		for i, v := range carry[1:] {
			if curValues[v] < minCarryValue {
				minCarryValue = curValues[v]
				minCarryIndex = uint8(i+1) // Note loops over indices 1-N
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
			numEmpty := emptySlots(carry) + 1 //Number of empty spaces in carry, good proxy for how many moves until return home
			for _, v := range carry {
				carVal += float32(min(curValues[v]+numEmpty, maxValues[v])) * discount // Otherwise, rounding happens items by item instead of as a set
			}
			endScore += uint16(carVal)
			return
		}
	} else { // Loop through legal moves and return best sequence/score
		depth++ // Going one level deeper
		// Update node values
		for i, v := range curValues {
			if v < maxValues[i] {
				curValues[i]++
			}
		}

		for i, v := range edgeConstraints[curNode] {
			if v.IsLegal && curScore >= v.MinScore && curScore <= v.MaxScore { // If a legal move exists
				carryCopy := make([]uint8, len(carry))
				curValuesCopy := make([]uint8, len(curValues))
				_ = copy(carryCopy, carry)
				_ = copy(curValuesCopy, curValues)
				thisScore := dig(uint8(i), curScore, carryCopy, curValuesCopy, depth, maxDepth, bestScore)
				if thisScore > bestScore { // If this sequence scores higher, replace sequence/score
					bestScore = thisScore
					bestSeq[depth-1] = uint8(i) // The first move goes in to the 0th slot
				}
			}
		}
		return bestScore
	}
}

func main() {
	start := time.Now()

	levels := readLevels("gathered.html")

	for i, l := range levels {
		carryLimit := l.CarryLimit
		moves := l.Moves

		prepGame(l.Edges, l.MaxValues)
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
				paramScore := runGame(indexOf(0, maxValues), carryLimit, moves, stepParam)
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

// Helper functions
// Increments each entry by 1 to switch from array references (0-N) to normal notation (1-N) for output
func oneUp(seq []uint8) {
	for i := range seq {
		seq[i]++
	}
}

// Reverses oneUp
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

// Checks whether a slice contains a particular value
func isIn(want uint8, seq []uint8) bool {
	for _, v := range seq {
		if v == want {
			return true
		}
	}
	return false
}

// Returns index of first appearance of a value in an slice, assumes value is present
func indexOf(want uint8, seq []uint8) uint8 {
	for i, v := range seq {
		if v == want {
			return uint8(i)
		}
	}
	return uint8(len(seq))
}

// Number of carry slots not occupied (part of scoring heuristic)
func emptySlots(seq []uint8) (numEmpty uint8) {
	for _, v := range seq {
		if maxValues[v] == 0 {
			numEmpty++
		}
	}
	return
}
