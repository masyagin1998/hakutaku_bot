package MGEWiki

import (
	"errors"
	"hakutaku_bot/libhakutaku/log"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

// ParseMGEWiki parses all info from MGEWiki and pushes it into Redis.
func ParseMGEWiki(updateTime string, redisClient *redis.Client, logger log.Logger) (err error) {
	if ok := redisClient.Exists("letters"); ok.Val() == 228 {
		logger.Info("The database already has information from MGEWiki")
	} else {
		logger.Info("Starting download information from MGEWiki")
		var names map[rune][]string
		names, err = getListOfGirls()
		if err != nil {
			return
		}

		logger.Info("Starting download \"Letter : Names\" pairs from MGEWiki")
		for k, v := range names {
			if len(v) != 0 {
				err = redisClient.SAdd("letters", k).Err()
				if err != nil {
					return
				}
				letterNames := strings.Join(v, "\n")
				err = redisClient.Set("letter:"+string(k), letterNames, 0).Err()
				if err != nil {
					return
				}
				logger.Info("Letter : Name", "Letter", string(k), "Name", "\n"+letterNames)
			}
		}

		logger.Info("Starting download Name : Link pairs from MGEWiki")
		for _, v := range names {
			for _, v1 := range v {
				var link string
				link, err = getGirlsLink(v1)
				if err != nil {
					return
				}
				if link == "" {
					err = errors.New("Can't get link for " + v1)
					return
				}
				err = redisClient.Set("name:"+v1, link, 0).Err()
				if err != nil {
					return
				}
				logger.Info("Name : Link", "Name", v1, "Link", link)
				time.Sleep(time.Second * 2)
			}
		}
	}
	return
}

// getListOfGirls tries to fing list of all girls.
func getListOfGirls() (names map[rune][]string, err error) {
	names = make(map[rune][]string)

	var response *http.Response
	response, err = http.Get("http://mgewiki.com/w/Category:Monster_Girls")
	if err != nil {
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return
	}
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

	// Alphabet.
	for k := 'A'; k <= 'Z'; k++ {
		names[k] = make([]string, 0)
	}
	for {
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

func getGirlsLink(name string) (link string, err error) {
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

	linkStopIndex2 := strings.Index(bodyString2, "Original file")
	quotMarkCounter := 0
	for quotMarkCounter != 5 {
		linkStopIndex2--
		if bodyString2[linkStopIndex2] == '"' {
			quotMarkCounter++
		}
	}
	linkStartIndex2 := linkStopIndex2 - 1
	for quotMarkCounter != 6 {
		linkStartIndex2--
		if bodyString2[linkStartIndex2] == '"' {
			quotMarkCounter++
		}
	}

	var response3 *http.Response
	link = "http://mgewiki.com" + bodyString2[linkStartIndex2+1:linkStopIndex2]
	response3, err = http.Get(link)
	if err != nil {
		return
	}
	defer response3.Body.Close()
	if response3.StatusCode != http.StatusOK {
		return
	}
	var bodyBytes3 []byte
	bodyBytes3, err = ioutil.ReadAll(response3.Body)
	if err != nil {
		return
	}
	if !strings.HasPrefix(http.DetectContentType(bodyBytes3), "image/") {
		link = ""
		return
	}
	return
}
