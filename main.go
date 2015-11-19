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

type Config struct {
	output          string
	columns_hash    []int
	columns_content []int
	deep_dirs       int
	delimiter       string
	iter            int
	keylength       int
	hash_dir        bool
	wg              sync.WaitGroup
}

/*
*  Read lines of text in channel and processes hashing, content and creation of files
*
*  @param {chan string} lines, chan string
*  @param {config Config} global config object
 */
func worker(lines chan string, config *Config) {

	var hash bytes.Buffer
	var content bytes.Buffer
	var dk []byte
	var f *os.File
	var cols []string
	var i int
	var salt []byte
	var filename string
	var folder string
	var path string

	//receives data in lines channel
	for l := range lines {
		hash.Reset()
		content.Reset()

		cols = strings.Split(l, config.delimiter)

		//creates hash string
		for _, v := range config.columns_hash {
			if v <= len(cols) {
				hash.WriteString(cols[v-1])
			}
		}

		//creates content for future file: two models, from one column to the end (negative integer to select column from) o list of columns
		if len(config.columns_content) == 1 && (config.columns_content)[0] < 0 {
			for i = ((config.columns_content)[0] * -1) - 1; i < len(cols); i++ {
				content.WriteString(cols[i] + ",")
			}
		} else {
			for _, v := range config.columns_content {
				if v <= len(cols) && v >= 0 {
					content.WriteString(cols[v-1] + config.delimiter)
				}
			}
		}

		//creates derived key and writes content
		salt = []byte(hash.String() + hash.String() + hash.String())
		dk = pbkdf2.Key([]byte(hash.String()+hash.String()), salt, config.iter, config.keylength, sha1.New)

		filename = hex.EncodeToString(dk)
		folder = filename[0:config.deep_dirs]

		if config.hash_dir {
			dk = pbkdf2.Key([]byte(filename[0:config.deep_dirs]), salt, config.iter, config.deep_dirs, sha1.New)
			folder = string(hex.EncodeToString(dk))
		}

		path = strings.Join(strings.Split(folder, ""), "/")
		createDir(config.output + "/" + path)

		f, _ = os.Create(config.output + "/" + path + "/" + filename)
		defer f.Close()

		f.WriteString(content.String())
		config.wg.Done()

	}

	config.wg.Done()
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
*  @param {config Config} global config object
 */
func statify(strfile string, config *Config) {

	if _, err := os.Stat(strfile); err == nil {

		if _, err := os.Stat(config.output); err != nil {
			createDir(config.output)
		}

		lines := make(chan string)

		log.Println("starting " + viper.GetString("max_threads") + " threads")
		for w := 1; w <= viper.GetInt("max_threads"); w++ {
			config.wg.Add(1)
			go worker(lines, config)
		}

		go readFile(strfile, lines, &(config.wg))

		config.wg.Wait()
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

	config := Config{}
	cols := strings.Split(viper.GetString("colums_hash"), ",")
	config.columns_hash = arrStrToInt(&cols)
	cols = strings.Split(viper.GetString("columns_content"), ",")
	config.columns_content = arrStrToInt(&cols)
	config.output = viper.GetString("outputdir")
	config.deep_dirs = viper.GetInt("deepdirs")
	config.delimiter = viper.GetString("delimiter")
	config.iter = viper.GetInt("pbkdf2_iterations")
	config.keylength = viper.GetInt("pbkdf2_keylength")
	config.hash_dir = viper.GetBool("hash_dir")

	if config.output == "" {
		config.output = "./output/"
	}

	if config.iter == 0 {
		config.iter = 100
	}

	if config.keylength == 0 {
		config.keylength = 32
	}

	if config.delimiter == "" {
		config.delimiter = ";"
	}

	log.Println("deleting " + config.output)
	os.RemoveAll(config.output)

	statify(viper.GetString("filename"), &config)
}
