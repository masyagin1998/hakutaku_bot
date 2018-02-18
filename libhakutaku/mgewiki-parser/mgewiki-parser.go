// Copyright (C) 2018 Mikhail Masyagin

/*
Package MGEWiki contains work with Redis and MGEWiki parser.
*/
package MGEWiki

import (
	"errors"
	"hakutaku_bot/libhakutaku/log"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

// InitDatabase parses all info from MGEWiki and pushes it into Redis.
func InitDatabase(redisClient *redis.Client, logger log.Logger) (err error) {
	// Checking if database is already exists.
	ok := false
	if ok, err = databaseCheck(redisClient, logger); ok || (err != nil) {
		return
	}
	// Getting list of girls.
	logger.Info("Updating database.")
	return parseMGEWiki(redisClient, logger)
}

func UpdateDatabase(updateTime string, redisClient *redis.Client, logger log.Logger) {
	// Getting current time.
	currentH, currentM, _ := time.Now().Clock()
	updateTimeArr := strings.Split(updateTime, ":")
	updateH, err := AtoI(updateTimeArr[0])
	if err != nil {
		logger.Error("Error occured, while running bot", "error", err)
		os.Exit(1)
	}
	updateM, err := AtoI(updateTimeArr[1])
	if err != nil {
		logger.Error("Error occured, while running bot", "error", err)
		os.Exit(1)
	}

	// Timer.
	update := updateH*60 + updateM
	current := currentH*60 + currentM
	if update > current {
		time.Sleep(time.Duration(update-current) * time.Minute)
	} else {
		time.Sleep(time.Duration(24*60-current+update) * time.Minute)
	}

	// Running updates.
	for {
		if pingMGEWiki(logger) {
			logger.Info("Updating database.")
			err := parseMGEWiki(redisClient, logger)
			if err != nil {
				logger.Error("Error occured, while running bot", "error", err)
				os.Exit(1)
			}
		}
		time.Sleep(time.Hour * 24)
	}
}

func AtoI(numStr string) (num int, err error) {
	var num64 int64
	num64, err = strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return
	}
	num = int(num64)
	return
}

func pingMGEWiki(logger log.Logger) (flag bool) {
	response, err := http.Get("http://mgewiki.com/w/Main_Page")
	if err != nil {
		logger.Error("Ping \"http://mgewiki.com/w/Main_Page\"", "error", err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		logger.Info("Ping \"http://mgewiki.com/w/Main_Page\"", "error", errors.New("Bad request "+string(response.StatusCode)))
		return
	}
	flag = true
	return
}

// databaseCheck checks if info about girls exists in database.
func databaseCheck(redisClient *redis.Client, logger log.Logger) (ok bool, err error) {
	// Checking if letters list exists.
	logger.Info("Checking database")
	var flag int64
	flag, err = redisClient.Exists("letters").Result()
	if err != nil {
		return
	}
	if flag == 0 {
		return
	}
	var letters []string
	letters, err = redisClient.SMembers("letters").Result()
	sort.Strings(letters)

	logger.Info("Letters", "letters", letters)
	for _, i := range letters {
		var flag int64
		flag, err = redisClient.Exists("letter:" + i).Result()
		if err != nil {
			err = errors.New("Damaged database. You should write in teminal: \"redis-cli\", \"FLUSHALL\" and restart app")
			return
		}
		if flag == 0 {
			return
		}
		var letterNames string
		letterNames, err = redisClient.Get("letter:" + i).Result()
		logger.Info("\"Letter : Names\"", "Letter", i, "Names", letterNames)
		names := strings.Split(letterNames, "\n")
		for _, name := range names {
			var link string
			link, err = redisClient.Get("name:" + strings.ToLower(name)).Result()
			if err != nil {
				err = errors.New("Damaged database. You should write in teminal: \"redis-cli\", \"FLUSHALL\" and restart app")
				return
			}
			logger.Info("\"Name : Link\"", "Name", name, "Link", link)
		}
	}

	// Database if ready.
	logger.Info("Database is ready")
	ok = true
	return
}

func parseMGEWiki(redisClient *redis.Client, logger log.Logger) (err error) {
	logger.Info("Starting download \"Letter : Names\" pairs from MGEWiki")
	var names map[rune][]string
	names, err = getListOfGirls()
	if err != nil {
		return
	}

	// Logging list of girls and pushing it to database.
	for i := 'A'; i <= 'Z'; i++ {
		if len(names[i]) == 0 {
			continue
		}
		err = redisClient.SAdd("letters", string(i)).Err()
		if err != nil {
			return
		}
		letterNames := strings.Join(names[i], "\n")
		err = redisClient.Set("letter:"+string(i), letterNames, 0).Err()
		if err != nil {
			return
		}
		logger.Info("\"Letter : Names\"", "Letter", string(i), "Names", letterNames)
		for _, name := range names[i] {
			var link string
			link, err = getGirlsLink(name)
			if err != nil {
				return
			}
			if link == "" {
				err = errors.New("Can't get link for " + name)
				return
			}
			err = redisClient.Set("name:"+strings.ToLower(name), link, 0).Err()
			if err != nil {
				return
			}
			logger.Info("\"Name : Link\"", "Name", name, "Link", link)
		}
	}
	logger.Info("Database is ready")
	return
}

// getListOfGirls tries to find list of all girls.
func getListOfGirls() (names map[rune][]string, err error) {
	// Initialization of map.
	names = make(map[rune][]string)

	// Request for start page.
	var response *http.Response
	response, err = http.Get("http://mgewiki.com/w/Category:Monster_Girls")
	if err != nil {
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return
	}
	// Searching begining of girls list in body.
	var bodyBytes []byte
	bodyBytes, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	bodyString := string(bodyBytes)
	header := "<h2>Pages in category \"Monster Girls\"</h2>"
	listStartIndex := strings.Index(bodyString, header)
	if listStartIndex == -1 {
		return
	}
	bodyString = bodyString[listStartIndex+len(header):]

	// Pushing alphabet to map.
	for k := 'A'; k <= 'Z'; k++ {
		names[k] = make([]string, 0)
	}

	// Search Loop.
	for {
		// Searching for girls in girls list.
		for k := 'A'; k <= 'Z'; k++ {
			linkStartIndex := strings.Index(bodyString, "<h3>"+string(k)+"</h3>")
			if linkStartIndex != -1 {
				bodyString = bodyString[linkStartIndex+len("<h3>_</h3><ul>")+1:]
				keyNames := bodyString[:strings.Index(bodyString, "</ul>")]
				for {
					nameStartIndex := strings.Index(keyNames, "\">")
					if nameStartIndex == -1 {
						break
					}
					keyNames = keyNames[nameStartIndex+2:]
					nameStopIndex := strings.Index(keyNames, "</a></li>")
					if nameStopIndex == -1 {
						break
					}
					names[k] = append(names[k], keyNames[:nameStopIndex])
					keyNames = keyNames[nameStopIndex+len("</a></li>"):]
				}
			}
		}
		// Searching for next page.
		linkStopIndex := strings.Index(bodyString, "next page")
		if (linkStopIndex != -1) && (bodyString[linkStopIndex-1] == '>') {
			quotMarkCounter := 0
			for quotMarkCounter != 3 {
				linkStopIndex--
				if bodyString[linkStopIndex] == '"' {
					quotMarkCounter++
				}
			}
			linkStartIndex := linkStopIndex - 1
			for quotMarkCounter != 4 {
				linkStartIndex--
				if bodyString[linkStartIndex] == '"' {
					quotMarkCounter++
				}
			}
			var response1 *http.Response
			response1, err = http.Get(strings.Replace("http://mgewiki.com"+bodyString[linkStartIndex+1:linkStopIndex], "&amp;", "&", -1))
			if err != nil {
				response1.Body.Close()
				return
			}
			if response1.StatusCode != http.StatusOK {
				response1.Body.Close()
				return
			}
			bodyBytes, err = ioutil.ReadAll(response1.Body)
			if err != nil {
				response1.Body.Close()
				return
			}
			bodyString = string(bodyBytes)
			listStartIndex := strings.Index(bodyString, header)
			if listStartIndex == -1 {
				break
			}
			bodyString = bodyString[listStartIndex+len(header):]
		} else {
			break
		}
	}
	return
}

// getGirlsLink tries to find current girl link.
func getGirlsLink(name string) (link string, err error) {
	// Request for girls page.
	var response1 *http.Response
	response1, err = http.Get("http://mgewiki.com/w/" + strings.Replace(name, " ", "_", -1))
	if err != nil {
		return
	}
	defer response1.Body.Close()
	if response1.StatusCode != http.StatusOK {
		return
	}
	var bodyBytes1 []byte
	bodyBytes1, err = ioutil.ReadAll(response1.Body)
	if err != nil {
		return
	}
	bodyString1 := string(bodyBytes1)

	linkStartIndex1 := strings.Index(bodyString1, "<ul class=\"gallery mw-gallery-traditional\">")
	bodyString1 = bodyString1[linkStartIndex1+len("<ul class=\"gallery mw-gallery-traditional\">"):]
	linkStartIndex1 = strings.Index(bodyString1, "_eng")
	linkStopIndex1 := linkStartIndex1
	for bodyString1[linkStartIndex1] != '"' {
		linkStartIndex1--
	}
	for bodyString1[linkStopIndex1] != '"' {
		linkStopIndex1++
	}

	// Request for image page.
	link = "http://mgewiki.com" + bodyString1[linkStartIndex1+1:linkStopIndex1]
	var response2 *http.Response
	response2, err = http.Get(link)
	if err != nil {
		return
	}
	defer response2.Body.Close()
	if response2.StatusCode != http.StatusOK {
		return
	}
	var bodyBytes2 []byte
	bodyBytes2, err = ioutil.ReadAll(response2.Body)
	if err != nil {
		return
	}
	bodyString2 := string(bodyBytes2)

	tag := "<div class=\"fullMedia\"><a href="
	linkStartIndex2 := strings.Index(bodyString2, tag)
	bodyString2 = bodyString2[linkStartIndex2+len(tag):]
	linkStartIndex2 = 1
	linkStopIndex2 := linkStartIndex2
	for {
		linkStopIndex2++
		if bodyString2[linkStopIndex2] == '"' {
			break
		}
	}

	// Image link.
	link = "http://mgewiki.com" + bodyString2[linkStartIndex2:linkStopIndex2]
	return
}
