package MGEWikiParser

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// FindGirl tries to find image with girl description.
func FindGirl(name string) (about string, err error) {
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

	linkStartIndex1 := strings.Index(strings.ToLower(bodyString1), "/w/file:"+strings.ToLower(strings.Replace(name, " ", "_", -1))+"_eng")
	linkStopIndex1 := linkStartIndex1
	for bodyString1[linkStopIndex1] != '"' {
		linkStopIndex1++
	}

	var response2 *http.Response
	link := "http://mgewiki.com" + bodyString1[linkStartIndex1:linkStopIndex1]
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
	if strings.HasPrefix(http.DetectContentType(bodyBytes3), "image/") {
		about = link
	}

	return
}

// GetListOfGirls tries to fing list of all girls.
func GetListOfGirls() (list []string, err error) {
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
	names := make(map[rune]string)
	for k := 'A'; k <= 'Z'; k++ {
		names[k] = ""
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
					names[k] += (keyNames[:nameStopIndex] + "\n")
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
	for k := 'A'; k <= 'Z'; k++ {
		if names[k] != "" {
			list = append(list, string(k)+"\n"+names[k])
		}
	}
	fmt.Println(list)
	return
}
