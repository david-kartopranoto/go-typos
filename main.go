package main

import (
	"log"
	"time"
)

var typo = "INV/20180426/XVIII / IV/154859122sya ingin komplain mengenai invoice ini ... hingga saat ini sya blm terima paket"
var src = "./raw"
var doGen = false
var tlFile = "./term_list.csv"
var biFile = "./bigram.csv"

func main() {
	var start time.Time
	var elapsed time.Duration

	if doGen {
		log.Printf("----------------Load file source to string \n")
		start = time.Now()
		fileStr := loadDirAsStr(src)
		log.Printf("result %d\n", len(fileStr))
		elapsed = time.Now().Sub(start)
		log.Printf("time taken %v s\n", elapsed.Seconds())

		log.Printf("----------------clean and split files \n")
		start = time.Now()
		termList := getWordTokens(fileStr)
		// log.Printf("%+v\n", termList)
		elapsed = time.Now().Sub(start)
		log.Printf("time taken %v s\n", elapsed.Seconds())

		log.Printf("----------------generate uni grams \n")
		start = time.Now()
		unigram := ngrams(termList, 1)
		log.Printf("result %T s\n", unigram)
		elapsed = time.Now().Sub(start)
		log.Printf("time taken %v s\n", elapsed.Seconds())

		log.Printf("----------------generate bi grams \n")
		start = time.Now()
		bigram := ngrams(termList, 2)
		log.Printf("result %T s\n", bigram)
		elapsed = time.Now().Sub(start)
		log.Printf("time taken %v s\n", elapsed.Seconds())

		log.Printf("----------------write uni grams \n")
		start = time.Now()
		generateCsv(tlFile, unigram)
		elapsed = time.Now().Sub(start)
		log.Printf("time taken %v s\n", elapsed.Seconds())

		log.Printf("----------------write bi grams \n")
		start = time.Now()
		generateCsv(biFile, bigram)
		elapsed = time.Now().Sub(start)
		log.Printf("time taken %v s\n", elapsed.Seconds())
	}

	// TODO: tripples and levenshein
	log.Printf("----------------load CSV \n")
	start = time.Now()
	tl, words := loadCsv(tlFile)
	elapsed = time.Now().Sub(start)
	log.Printf("time taken %v s\n", elapsed.Seconds())

	log.Println(len(words))
	log.Printf("%T\n", tl)

	// log.Printf("----------------load Fuzzy \n")
	// start = time.Now()
	// model := loadFuzzy(words)
	// elapsed = time.Now().Sub(start)
	// log.Printf("time taken %v s\n", elapsed.Seconds())

	// log.Printf("----------------suggest ex false \n")
	// start = time.Now()
	// log.Println("ex false", model.Suggestions("komplain", false))
	// elapsed = time.Now().Sub(start)
	// log.Printf("time taken %v s\n", elapsed.Seconds())

	// log.Printf("----------------suggest ex true \n")
	// start = time.Now()
	// log.Println("ex true", model.Suggestions("komplain", true))
	// elapsed = time.Now().Sub(start)
	// log.Printf("time taken %v s\n", elapsed.Seconds())

	log.Printf("----------------generate tripples \n")
	start = time.Now()
	res := generateTripples(tl, "komplain")
	elapsed = time.Now().Sub(start)
	log.Printf("time taken %v s\n", elapsed.Seconds())
	log.Printf("Tripples (%v) %v \n", len(res), res)

	log.Printf("----------------generate tripples 2 \n")
	start = time.Now()
	res2 := generateTripples2(tl, "komplain")
	elapsed = time.Now().Sub(start)
	log.Printf("time taken %v s\n", elapsed.Seconds())

	log.Printf("Tripples 2 (%v) %v \n", len(res2), res2)

}
