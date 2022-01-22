package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type User struct {
	Browsers []string `json:"browsers"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
}

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	// j, err := ioutil.ReadFile("./data/users.txt")
	// if err != nil {
	// 	fmt.Println("Error reading file", err)
	// }
	// var lines []string
	// scanner := bufio.NewScanner(file)
	// for scanner.Scan() {
	// 	lines = append(lines, scanner.Text())
	// }
	// fileContents, err := ioutil.ReadAll(file)
	// if err != nil {
	// 	panic(err)
	// }
	in := bufio.NewScanner(file)
	seenBrowsers := []string{}
	uniqueBrowsers := 0
	foundUsers := ""

	// users := []User{}

	// for _, line := range lines {
	// 	user := User{}
	// 	err = json.Unmarshal([]byte(line), &user)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	users = append(users, user)
	// }
	i := 0
	for in.Scan() {
		i++
		row := in.Bytes()
		if bytes.Contains(row, []byte("Android")) == false && bytes.Contains(row, []byte("MSIE")) == false {
			continue
		}
		user := User{}
		json.Unmarshal(row, &user)
		isAndroid := false
		isMSIE := false

		for _, browserRaw := range user.Browsers {
			browser := browserRaw

			if strings.Contains(browser, "Android") {
				isAndroid = true

				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		for _, browserRaw := range user.Browsers {
			browser := browserRaw

			if strings.Contains(browser, "MSIE") {
				isMSIE = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		email := strings.Replace(user.Email, "@", " [at] ", -1)
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i-1, user.Name, email)
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
func main() {

	FastSearch(ioutil.Discard)

}
