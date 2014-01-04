package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
)

type fileData struct {
	groupNumber int
	data        string
}

type fileHeap []fileData

func (fh fileHeap) Len() int { return len(fh) }
func (fh fileHeap) Less(i, j int) bool {
	inta, err := strconv.Atoi(fh[i].data)
	if err != nil {
		panic(err)
	}
	intb, err := strconv.Atoi(fh[j].data)
	if err != nil {
		panic(err)
	}
	return inta < intb
}

func (fh fileHeap) Swap(i, j int) {
	fh[i], fh[j] = fh[j], fh[i]
}

func (h *fileHeap) Push(x interface{}) {
	*h = append(*h, x.(fileData))
}

func (h *fileHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// func merge(src, dst []int, first, mid, last int) {
// 	i, j := first, mid+1
// 	k := 0
// 	for ; i <= mid && j <= last; k++ {
// 		if src[i] < src[j] {
// 			dst[k] = src[i]
// 			i++
// 		} else {
// 			dst[k] = src[j]
// 			j++
// 		}
// 	}
// 	for ; i <= mid; k++ {
// 		dst[k] = src[i]
// 		i++
// 	}
// 	for ; j <= last; k++ {
// 		dst[k] = src[j]
// 		j++
// 	}
// 	copy(src[first:last+1], dst[:k])
// }

// func mergeSort(src, dst []int, first, last int) {
// 	if first < last {
// 		mid := (first + last) / 2
// 		mergeSort(src, dst, first, mid)
// 		mergeSort(src, dst, mid+1, last)
// 		merge(src, dst, first, mid, last)
// 	}
// }

func partition(src []int, s, e int) int {
	key := src[s]
	i := s
	for j := s + 1; j <= e; j++ {
		if src[j] < key {
			i++
			src[i], src[j] = src[j], src[i]
		}
	}
	src[s], src[i] = src[i], src[s]
	return s
}

func quickSort(src []int, s, e int) {
	if s < e {
		partition := partition(src, s, e)
		quickSort(src, s, partition-1)
		quickSort(src, partition+1, e)
	}
}

var filePrefix = "data"

func pufhRandomDataToFiles() {
	var n int
	for i := 0; i < 32; i++ {
		file, err := os.OpenFile(fmt.Sprintf("%s_%d.txt", filePrefix, i), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			panic(err)
		}
		bw := bufio.NewWriter(file)
		fileSize := 0
		for fileSize < int(float64(BlockSize)*0.8) {
			n, err = bw.WriteString(strconv.Itoa(rand.Int() % 65535))
			if err != nil {
				panic(err)
			}
			err := bw.WriteByte(' ')
			if err != nil {
				panic(err)
			}
			fileSize += n + 1
		}
		bw.Flush()
		file.Close()
	}
}

func sortDataFiles() {
	for i := 0; i < DataSize/M; i++ {
		var intArray []int
		for j := 0; j < M; j++ {
			file, err := os.Open(fmt.Sprintf("%s_%d.txt", filePrefix, i*M+j))
			if err != nil {
				panic(err)
			}
			data, err := ioutil.ReadAll(file)
			if err != nil {
				file.Close()
				panic(err)
			}
			file.Close()

			array := strings.Fields(string(data))
			for _, str := range array {
				d, err := strconv.Atoi(str)
				if err != nil {
					panic(err)
				}
				intArray = append(intArray, d)
			}
		}
		sort.Ints(intArray)
		// quickSort(intArray, 0, len(intArray)-1)
		// rewrite the data to the file
		partLen := len(intArray) / M
		var partArray []int
		for j := 0; j < M; j++ {
			file, err := os.OpenFile(fmt.Sprintf("sorted_%s_%d_%d.txt", filePrefix, i, j), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
			if err != nil {
				panic(err)
			}
			bw := bufio.NewWriter(file)
			if j != M-1 {
				partArray = intArray[j*partLen : (j+1)*partLen]
			} else {
				partArray = intArray[j*partLen:]
			}
			for _, d := range partArray {
				bw.WriteString(strconv.Itoa(d))
				bw.WriteByte(' ')
			}
			bw.Flush()
			file.Close()
		}
	}
}

var M = 8
var DataSize = 32
2
func multiMerge() {
	var err error
	num := DataSize / M
	groupFileNumbers := make([]int, num)
	files := make([]*bufio.Reader, num)
	fh := &fileHeap{}

	// init
	for i := 0; i < num; i++ {
		file, err := os.Open(fmt.Sprintf("sorted_%s_%d_%d.txt", filePrefix, i, 0))
		if err != nil {
			panic(err)
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		array := strings.Fields(string(data))
		for _, str := range array {
			fd := fileData{i, str}
			*fh = append(*fh, fd)
		}
		file.Close()
		file, err = os.Open(fmt.Sprintf("sorted_%s_%d_%d.txt", filePrefix, i, 1))
		if err != nil {
			panic(err)
		}
		defer file.Close()
		files[i] = bufio.NewReaderSize(file, 512)
		groupFileNumbers[i] = 1
	}
	heap.Init(fh)
	fmt.Println(len(*fh))

	var fileIndex int = 0
	completeMap := make(map[int]bool)
	var data fileData
	var outFile *os.File
	var bw *bufio.Writer
	var fileSize int

	for len(completeMap) < num {
		fileSize = 0
		outFile, err = os.Create(fmt.Sprintf("outfile_%d.txt", fileIndex))
		if err != nil {
			panic(err)
		}
		bw = bufio.NewWriter(outFile)
		for fileSize < int(float64(BlockSize)*0.8) {
			data = heap.Pop(fh).(fileData)
			bw.WriteString(data.data)
			bw.WriteByte(' ')
			fileSize += len(data.data) + 1
			groupNumber := data.groupNumber
			// make sure the idx group is not completed
			for {
				if _, ok := completeMap[groupNumber]; ok {
					groupNumber = (groupNumber + 1) % num
				} else {
					break
				}
			}
		ReadData:
			readData, err := files[groupNumber].ReadString(' ')
			if err != nil {
				if err == io.EOF {
					if groupFileNumbers[groupNumber] < M-1 {
						groupFileNumbers[groupNumber]++
						file, err := os.Open(fmt.Sprintf("sorted_%s_%d_%d.txt", filePrefix, groupNumber, groupFileNumbers[groupNumber]))
						if err != nil {
							panic(err)
						}
						defer file.Close()
						files[groupNumber] = bufio.NewReaderSize(file, 512)
						goto ReadData
					} else {
						completeMap[groupNumber] = true
						if len(completeMap) == num {
							break
						}
						groupNumber = (groupNumber + 1) % num
						goto ReadData
					}
				} else {
					panic(err)
				}
			} else {
				newData := fileData{groupNumber, readData[0 : len(readData)-1]}
				heap.Push(fh, newData)
			}
		}
		bw.Flush()
		outFile.Close()
		fileSize = 0
		fileIndex++
		if len(completeMap) < num {
			outFile, err = os.Create(fmt.Sprintf("outfile_%d.txt", fileIndex))
			if err != nil {
				panic(err)
			}
			bw = bufio.NewWriterSize(outFile, 512)
		}
	}
	for {
		outFile, err = os.Create(fmt.Sprintf("outfile_%d.txt", fileIndex))
		if err != nil {
			panic(err)
		}
		bw = bufio.NewWriterSize(outFile, 512)
		fileSize = 0
		for fileSize < int(float64(BlockSize)*0.8) {
			if len(*fh) == 0 {
				fmt.Println("complete")
				os.Exit(1)
			}
			data := heap.Pop(fh).(fileData)
			bw.WriteString(data.data)
			bw.WriteByte(' ')
			fileSize += len(data.data) + 1
		}
		bw.Flush()
		outFile.Close()
		fileIndex++
	}

}

func main() {

	pufhRandomDataToFiles()
	sortDataFiles()
	multiMerge()
}
