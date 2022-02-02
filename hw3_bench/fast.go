package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

type User struct {
	Browsers []string `json:"browsers"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
}

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	runtime.GOMAXPROCS(8)
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	in := bufio.NewScanner(file)
	seenBrowsers := []string{}
	uniqueBrowsers := 0
	foundUsers := ""

	for i := 0; in.Scan(); i++ {

		row := in.Bytes()
		if !bytes.Contains(row, []byte("Android")) && !bytes.Contains(row, []byte("MSIE")) {
			continue
		}
		user := User{}
		err = json.Unmarshal(row, &user)
		if err != nil {
			return
		}
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
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email)
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
