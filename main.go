package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"github.com/spf13/viper"
	"golang.org/x/crypto/pbkdf2"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

/*
*  Read lines of text in channel and processes hashing, content and creation of files
*
*  @param {chan string} lines, chan string
*  @param {*sync.WaitGroup} wait group, *sync.WaitGroup
*  @param {string} output dir, string
*  @param {*[]int} columns_hash, *[]int
*  @param {*[]int} columns_content, *[]int
 */
func worker(lines chan string, wg *sync.WaitGroup, output *string, cols_hash *[]int, cols_content *[]int, deepdirs *int, delimiter *string, iter *int, keylength *int, hash_dir *bool) {

	var hash bytes.Buffer
	var content bytes.Buffer

	//receives data in lines channel
	for l := range lines {
		hash.Reset()
		content.Reset()

		cols := strings.Split(l, *delimiter)

		//creates hash string
		for _, v := range *cols_hash {
			if v <= len(cols) {
				hash.WriteString(cols[v-1])
			}
		}

		//creates content for future file: two models, from one column to the end (negative integer to select column from) o list of columns
		if len(*cols_content) == 1 && (*cols_content)[0] < 0 {
			for i := ((*cols_content)[0] * -1) - 1; i < len(cols); i++ {
				content.WriteString(cols[i] + ",")
			}
		} else {
			for _, v := range *cols_content {
				if v <= len(cols) && v >= 0 {
					content.WriteString(cols[v-1] + ",")
				}
			}
		}

		//creates derived key
		salt := []byte(hash.String() + hash.String() + hash.String())
		dk := pbkdf2.Key([]byte(hash.String()+hash.String()), salt, *iter, *keylength, sha1.New)

		filename := hex.EncodeToString(dk)
		folder := filename[0:*deepdirs]

		if *hash_dir {
			dk2 := pbkdf2.Key([]byte(filename[0:*deepdirs]), salt, *iter, *deepdirs, sha1.New)
			folder = string(hex.EncodeToString(dk2))
		}

		path := strings.Join(strings.Split(folder, ""), "/")
		createDir(*output + "/" + path)

		f, _ := os.Create(*output + "/" + path + "/" + filename)
		defer f.Close()

		f.WriteString(content.String())
		wg.Done()

	}

	wg.Done()
}

/*
*  Creates full paths from string
*
*  @param {string} path
*
 */
func createDir(path string) {
	os.MkdirAll(path, 0644)
}

/*
*  Reads csv to transform in flat files database
*
*  @param {string}file to read
*  @param {chan string} channel
*  @param {*sync.WaitGroup} wait group
 */
func readFile(strfile string, lines chan string, wg *sync.WaitGroup) {

	if file, err := os.Open(strfile); err == nil {

		log.Println("reading file")
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			wg.Add(1)
			lines <- scanner.Text()
		}
		close(lines)
		if err = scanner.Err(); err != nil {
			log.Fatal(err)
		}

	} else {
		log.Fatal(err)
	}
}

/*
*  Function that start all the jobs: workers and readfile
*
*  @param {string} file to read
*  @param {string} output directory
*  @param {*[]int} columns_hash
*  @param {*[]int} columns_content
 */
func statify(strfile string, output *string, cols_hash *[]int, cols_content *[]int, deepdirs *int, delimiter *string, iter *int, keylength *int, hash_dir *bool) {

	if _, err := os.Stat(strfile); err == nil {

		if _, err := os.Stat(*output); err != nil {
			createDir(*output)
		}

		lines := make(chan string)

		var wg sync.WaitGroup

		log.Println("starting " + viper.GetString("max_threads") + " threads")
		for w := 1; w <= viper.GetInt("max_threads"); w++ {
			wg.Add(1)
			go worker(lines, &wg, output, cols_hash, cols_content, deepdirs, delimiter, iter, keylength, hash_dir)
		}

		go readFile(strfile, lines, &wg)

		wg.Wait()
		log.Println("end")
	}

}

/*
*  Cast array elements from string to int. Comes from yaml properties
*
*  @param arr pointer to array of strings
*  @return array of integers
 */
func arrStrToInt(arr *[]string) []int {

	cparr := make([]int, len(*arr))

	for i, v := range *arr {
		v2, _ := strconv.Atoi(v)
		cparr[i] = v2
	}

	return cparr
}

func main() {

	log.SetOutput(os.Stdout)

	viper.SetConfigName("config")
	viper.AddConfigPath("./")

	err := viper.ReadInConfig()
	if err != nil {
		log.Println("missing config file")
		os.Exit(1)
	}

	cols := strings.Split(viper.GetString("colums_hash"), ",")
	cols_hash := arrStrToInt(&cols)
	cols = strings.Split(viper.GetString("columns_content"), ",")
	cols_content := arrStrToInt(&cols)
	output := viper.GetString("outputdir")
	deepdirs := viper.GetInt("deepdirs")
	delimiter := viper.GetString("delimiter")
	iter := viper.GetInt("pbkdf2_iterations")
	keylength := viper.GetInt("pbkdf2_keylength")
	hash_dir := viper.GetBool("hash_dir")

	if output == "" {
		output = "./output/"
	}

	if iter == 0 {
		iter = 100
	}

	if keylength == 0 {
		keylength = 32
	}

	if delimiter == "" {
		delimiter = ";"
	}

	log.Println("deleting " + output)
	os.RemoveAll(output)

	statify(viper.GetString("filename"), &output, &cols_hash, &cols_content, &deepdirs, &delimiter, &iter, &keylength, &hash_dir)
}
