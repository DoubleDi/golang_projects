package main

import (
	_ "encoding/json"
	"fmt"
	"io"
	_ "io/ioutil"
	"os"
	_ "regexp"
	"strings"
	_ "log"
     "bufio"
)

type webData struct {
	Browsers []string `json:"browsers"`
	Company  string   `json:"company"`
	Country  string   `json:"country"`
	Email    string   `json:"email"`
	Job      string   `json:"job"`
	Name     string   `json:"name"`
	Phone    string   `json:"phone"`
}

func FastSearch(out io.Writer) {
    var err error
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
    
    var i int
    var isAndroid bool
    var isMSIE bool
    var user webData
    var browser string
    var notSeenBefore bool
    seenBrowsers := make(map[string]bool, 1000)
    uniqueBrowsers := 0
    foundUsers := ""

    fileReader := bufio.NewReader(file)
    fileContents, _, err := fileReader.ReadLine()
	for fileContents != nil {
    	user = webData{}
    	err = user.UnmarshalJSON(fileContents)
    	if err != nil {
    		panic(err)
    	}
        
		isAndroid = false
		isMSIE = false

		browsers := user.Browsers
		for _, browser = range browsers {
            notSeenBefore = false
			if strings.Contains(browser, "Android") {
				isAndroid = true
				notSeenBefore = true
                
			} else if strings.Contains(browser, "MSIE")  {
				isMSIE = true
				notSeenBefore = true
            }
            
            if notSeenBefore {
                if _, ok := seenBrowsers[browser]; !ok {
                    seenBrowsers[browser] = true
                    uniqueBrowsers++
                }
            }
    	}
        
        fileContents, _, err = fileReader.ReadLine()
        
        if isAndroid && isMSIE {
            foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, user.Email)
        }

        i++
    }

    foundUsers = strings.Replace(foundUsers, "@", " [at] ", -1)
	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", uniqueBrowsers)
}

func main() {
    
}
