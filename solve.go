package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "strings"
    "strconv"
	"sync"
	"runtime"
	"time"
	"runtime/pprof"
    "os"
)

type Sudoku struct {
  Grid[9][9]int
}

func (this Sudoku) ReadFile(ch chan Sudoku, wg *sync.WaitGroup){
  const path = "./Sudoku/";

  files, err := ioutil.ReadDir(path)
  if err != nil {
      log.Fatal(err)
  }

  for _, f := range files {
          b, err := ioutil.ReadFile(path + f.Name()) // just pass the file name
          if err != nil {
              fmt.Print(err)
          }

          str := string(b) // convert content to a 'string'
		  var valuesInFiles []string

		  if strings.ContainsAny(str, "\r") { //if .txt mac/linux or windows
			valuesInFiles = strings.Split(str, "\r\n")
		  } else {
			  valuesInFiles = strings.Split(str, "\n")
		  }

          for index, _ := range valuesInFiles {
              var valuesInArray = strings.Split(valuesInFiles[index], "")
              for o, _ := range valuesInArray {
                  var value = 0
                  if(valuesInArray[o] == ".") {
                    value = 0
                  } else {
                    v, _ := strconv.Atoi(valuesInArray[o]) //convert string to int
                    value = v
                  }
                  this.Grid[index][o] = value
              }

          }
        wg.Add(1)
  		ch <- this
  }

}


func (this *Sudoku) Display() {
	for kx, x := range this.Grid {
		for ky, y := range x {
			fmt.Print(y)
			if ky == 2 || ky == 5 {
				fmt.Print("|")
			}
		}
		fmt.Println("")
		if kx == 2 || kx == 5 {
			for i := 0; i < 11; i++ {
				fmt.Print("-")
			}
			fmt.Println("")
		}
	}
}

func (this *Sudoku) Check() bool {
	for x := 0; x < 9; x++ {
		acc := make(map[int]bool)
		for y := 0; y < 9; y++ {
			val := this.Grid[x][y]
			if acc[val] && val != 0 {
				return false
			}
			acc[val] = true
		}
	}

	for y := 0; y < 9; y++ {
		acc := make(map[int]bool)
		for x := 0; x < 9; x++ {
			val := this.Grid[x][y]
			if acc[val] && val != 0 {
				return false
			}
			acc[val] = true
		}
	}

	for cadX := 0 ; cadX < 3 ; cadX++ {
	    for cadY := 0 ; cadY < 3 ; cadY++ {
	        acc := make(map[int]bool)

	        for x := cadX*3; x < 3; x++ {
	            for y := cadY*3; y < 3; y++ {
    	            val := this.Grid[x][y]
        			if acc[val] && val != 0 {
        				return false
        			}
        			acc[val] = true
    	        }
	        }
	    }
	}

	return true
}

func (this *Sudoku) Solve() {
	coords := this.getMissingsNumbers()
	this.solveRecursion(coords)
}

func (this *Sudoku) getMissingsNumbers() (res [][2]int) {
	for ky, vy := range this.Grid {
		for kx, vx := range vy {
			if vx == 0 {
				add := [2]int{ky, kx}
				res = append(res, add)
			}
		}
	}
	return
}

func (this *Sudoku) solveRecursion(coords [][2]int) bool {
	if len(coords) == 0 {
		return true
	}
	y := coords[0][0]
	x := coords[0][1]
	for n := 1; n <= 9; n++ {
		if this.checkCoord(y, x, n) {
			this.Grid[y][x] = n
			if this.solveRecursion(coords[1:]) {
				return true
			}
			this.Grid[y][x] = 0
		}
	}
	return false
}

func (this *Sudoku) checkCoord(cy int, cx int, nVal int) bool {
	// Line
	for x := 0 ; x < 9 ; x++ {
		val := this.Grid[cy][x]
		if val == nVal {
			return false
		}
	}

	// Col
	for y := 0 ; y < 9 ; y++ {
		val := this.Grid[y][cx]
		if val == nVal {
			return false
		}
	}

	// square
	by := cy - (cy % 3)
	bx := cx - (cx % 3)
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {

			val := this.Grid[by + y][bx + x]
			if val == nVal {
				return false
			}

		}

	}

	return true
}

func solver(ch chan Sudoku, ch2 chan Sudoku) { //channel 1 for unsolved sudoku
	for {
		var val Sudoku
		val = <- ch
		val.Solve()
		ch2 <- val
	}
}

func go_display(ch2 chan Sudoku, wg *sync.WaitGroup){ //channel 2 for solved sudoku
	for {
		var val Sudoku
	    val = <- ch2
	    fmt.Println()
	    val.Display()
	    fmt.Println()
	    wg.Done()
	}
}

func main() {
	start_time := time.Now()
    f, err := os.Create("perf_cpu.prof")
    if err != nil {
        log.Fatal("Could not create CPU profile: ", err)
    }
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("Could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	var base Sudoku
	var ch chan Sudoku
	var ch2 chan Sudoku
	ch = make(chan Sudoku)
	ch2 = make(chan Sudoku)
    var wg sync.WaitGroup

	for i := 0; i < runtime.NumCPU(); i++ { //Check nb logical core
		go solver(ch, ch2)
	}
	go go_display(ch2, &wg)

	base.ReadFile(ch, &wg)

	wg.Wait()
	//fmt.Println(runtime.NumCPU())
	fmt.Print("The complete operation took : ")

	f, err = os.Create("perf_mem.prof")
	if err != nil {
		log.Fatal("Could not create memory profile: ", err)
	}
	runtime.GC() //Get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("Could not write memory profile: ", err)
	}
	end_time := time.Now()
	fmt.Print(end_time.Sub(start_time))

	f.Close()
}
