package main

import (
	"os"
	"fmt"
	"bufio"
	"text/template"
	"net/url"
	"bytes"
	"net/http" 
	"strings"
	"github.com/TwinProduction/go-color"
	"flag"
	"strconv"
)

type arrayFlags []string

type options struct{
	codesArr []string
	badCodesArr []string
	cookies string
	headers []string
}

func header(){
	fmt.Println(color.Ize(color.Green,`		      _               _                 
		     | |             | |                
		  ___| |__   ___  ___| | _____ _ __
		 / __| '_ \ / _ \/ __| |/ / _ \ '__/
		| (__| | | |  __/ (__|   <  __/ |  
		 \___|_| |_|\___|\___|_|\_\___|_| 
		 `))
}

func help(){
	fmt.Println(`
		Automated Checker Tool by chivato (inspired by m0rphe)
	Usage:
		checker -u 'https://test.me/'
	Options:
		-u	URL to test against (required)
		-c	Possitive response codes to highlight
		-b	Negative response codes to exclude
		-hostf	Host header file location
		-pathf	Path file location
		-header	Header file location 
		-C	Cookies to include in requests
		-H	Individual headers to be added
		-l	Content length to match in response
	Examples:
		checker -u 'http://test.me/admin' -C "Cookie1=testing; cookie2=hello; cookie3=test" -l 53
		checker -u 'http://test.me/' -b "404,500,403" -c "200"
		checker -u 'http://test.me' -H "Admin: True" -H "LoggedIn: True" -pathf 'path/to/wordlist'
			`)
	os.Exit(3)
}

func (i *arrayFlags) Set(value string) error {
    *i = append(*i, value)
    return nil
}

func (i *arrayFlags) String() string {
    return "my string representation"
}

func Find(slice []string, val string) (int, bool) {
    for i, item := range slice {
        if item == val {
            return i, true
        }
    }
    return -1, false
}

func readLines(path string) ([]string, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}

func FormatWordlist(array []string, parsedInput *url.URL)([]string){
	output_array := []string{}
	buf := new(bytes.Buffer)
	for x:=0; x < len(array); x++{
		buf = new(bytes.Buffer)
		tmpl, _ := template.New("template").Parse(array[x])
		tmpl.Execute(buf, parsedInput)
		output_array = append(output_array,buf.String())
	}
	return output_array
}

func ExploitPaths(path_checks []string,parsedInput *url.URL, currentOptions *options,requestedLength int){
	for x:=0;x<len(path_checks);x++{
		client := &http.Client{}
		sendTo := parsedInput.Scheme + "://" + parsedInput.Host + path_checks[x]
		req, _ := http.NewRequest("GET", sendTo, nil)
		req = AddHeadersToReq(req, currentOptions.headers)
		if len(currentOptions.cookies) !=0{
			req = CreateCookie(currentOptions.cookies,req)
		}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("server not responding %s", err.Error())
			os.Exit(1)
		}
		if resp.ContentLength == -1{
			resp.ContentLength = 0
		}
		inBadCodes, colour := badCodesChecker(currentOptions.badCodesArr, currentOptions.codesArr, strconv.Itoa(resp.StatusCode))
		if requestedLength != -1{
			if requestedLength == int(resp.ContentLength) && inBadCodes == false{
				fmt.Print(color.Ize(colour,fmt.Sprintf("	- %s | Status code: %s | Content Length: %d\n",path_checks[x],resp.Status,resp.ContentLength)))
			}
		}else if inBadCodes == false{
			fmt.Print(color.Ize(colour,fmt.Sprintf("	- %s | Status code: %s | Content Length: %d\n",path_checks[x],resp.Status,resp.ContentLength)))
		}
	}
	return
}

func ExploitHeaders(header_checks []string,parsedInput *url.URL, currentOptions *options,requestedLength int){
	var headerParts []string
	for x:=0;x<len(header_checks);x++{
		client := &http.Client{}
		headerParts = strings.Split(header_checks[x],":")
		req, _ := http.NewRequest("GET", string(parsedInput.Scheme + "://" + parsedInput.Host + parsedInput.Path), nil)
		req = AddHeadersToReq(req, currentOptions.headers)
		req.Header.Add(headerParts[0], headerParts[1])
		if len(currentOptions.cookies) !=0{
			req = CreateCookie(currentOptions.cookies,req)
		}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("server not responding %s", err.Error())
			os.Exit(1)
		}
		if resp.ContentLength == -1{
			resp.ContentLength = 0
		}
		inBadCodes, colour := badCodesChecker(currentOptions.badCodesArr, currentOptions.codesArr, strconv.Itoa(resp.StatusCode))
		if requestedLength != -1{
			if requestedLength == int(resp.ContentLength) && inBadCodes == false{
				fmt.Print(color.Ize(colour,fmt.Sprintf("	- %s | Status code: %s | Content Length: %d\n",header_checks[x],resp.Status,resp.ContentLength)))
			}
		}else if inBadCodes == false{
			fmt.Print(color.Ize(colour,fmt.Sprintf("	- %s | Status code: %s | Content Length: %d\n",header_checks[x],resp.Status,resp.ContentLength)))
		}
	}
	return
}

func ExploitHostHeaders(host_header_checks []string, parsedInput *url.URL, currentOptions *options, requestedLength int){
	for x:=0;x<len(host_header_checks);x++{
		client := &http.Client{}
		req, _ := http.NewRequest("GET", string(parsedInput.Scheme + "://" + parsedInput.Host + parsedInput.Path), nil)
		req = AddHeadersToReq(req, currentOptions.headers)
		req.Host = host_header_checks[x]
		if len(currentOptions.cookies) !=0{
			req = CreateCookie(currentOptions.cookies,req)
		}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Server not responding: %s", err.Error())
			os.Exit(1)
		}
		if resp.ContentLength == -1{
			resp.ContentLength = 0
		}
		inBadCodes, colour := badCodesChecker(currentOptions.badCodesArr, currentOptions.codesArr, strconv.Itoa(resp.StatusCode))
		if requestedLength != -1{
			if requestedLength == int(resp.ContentLength) && inBadCodes == false{
				fmt.Print(color.Ize(colour,fmt.Sprintf("	- %s | Status code: %s | Content Length: %d\n",host_header_checks[x],resp.Status,resp.ContentLength)))
			}
		}else if inBadCodes == false{
			fmt.Print(color.Ize(colour,fmt.Sprintf("	- %s | Status code: %s | Content Length: %d\n",host_header_checks[x],resp.Status,resp.ContentLength)))
		}
	}
	return
}

func CreateCookie(cookies string,req *http.Request)*http.Request{
	eachOne := strings.Split(cookies, "; ")
	for x := 0; x < len(eachOne); x++{
		individual := strings.Split(eachOne[x],"=")
		req.AddCookie(&http.Cookie{Name: individual[0], Value: individual[1]})
	}
	return req
}

func AddHeadersToReq(req *http.Request, headers []string)(*http.Request){
	headerParts := []string{}
	if len(headers) != 0 {
		for y:=0;y<len(headers);y++{
			headerParts = strings.Split(headers[y],":")
			req.Header.Add(headerParts[0], headerParts[1])
		}
	}
	return req
}

func badCodesChecker(badCodesArr []string, codesArr []string, statusCode string)(bool,string){
	inBadCodes := false
	colour := color.White
	found := false
	if len(badCodesArr) != 0 {
		_, found = Find(badCodesArr,statusCode)
		if found == true{
			inBadCodes = true
		}
	}

	if len(codesArr) != 0 {
		_, found = Find(codesArr,statusCode)
		if found == true{
			colour = color.Green
		}else{
			colour = color.White
		}
	}
	_ = found
	return inBadCodes, colour
}

func GetCommandLineArgs()(int,string,string,string,string,arrayFlags,string,string,string){
	var myHeaders arrayFlags
	urlPtr := flag.String("u", "", "Required: URL to test against.")
	codePtr := flag.String("c", "", "Possitive codes to look out for")
	badCodePtr := flag.String("b", "", "Negative codes to exclude")
	hostHeaderLocationPtr := flag.String("hostf", "host_header_checks", "Host header file location")
	headerLocationPtr := flag.String("header", "header_checks", "Header file location")
	pathLocationPtr := flag.String("pathf", "path_checks", "Path file location")
	cookiesPtr := flag.String("C", "", "Cookies to include")
	contentLengthPtr := flag.Int("l",-1,"Content Length to match in request")
	flag.Var(&myHeaders, "H", "Additional headers to include")
	flag.Parse()
	return *contentLengthPtr, *urlPtr, *codePtr, *badCodePtr, *cookiesPtr, myHeaders, *hostHeaderLocationPtr, *headerLocationPtr, *pathLocationPtr
}

func ExploitChecks(path_checks []string, header_checks []string, host_header_checks []string, parsedInput *url.URL, currentOptions *options, contentLength int){
	tick:=0
	if len(path_checks) != 0{
		fmt.Print("\n= Checking path exploits =\n\n")
		ExploitPaths(path_checks,parsedInput,currentOptions,contentLength)
	}else{
		fmt.Println(color.Ize(color.Red,"Path check wordlist not specified or does not exist (default: path_checks)"))
		tick++
	}
	if len(header_checks) != 0{
		fmt.Print("\n= Checking header exploits =\n\n")
		ExploitHeaders(header_checks,parsedInput,currentOptions,contentLength)
	}else{
		fmt.Println(color.Ize(color.Red,"Header check wordlist not specified or does not exist (default: header_checks)"))
		tick++
	}
	if len(host_header_checks) != 0{
		fmt.Print("\n= Checking host header exploits =\n\n")
		ExploitHostHeaders(host_header_checks,parsedInput,currentOptions,contentLength)
	}else{
		fmt.Println(color.Ize(color.Red,"Host header check wordlist not specified or does not exist (default: host_header_checks)"))
		tick++
	}
	if tick == 3{
		fmt.Println("No valid input files found")
	}else{
	fmt.Println("")
	}
}

func main(){
	header()

	currentOptions := new(options)
	var contentLength int
	var input,hostHeadersFile,headersFile,pathsFile,codes,badCodes string
	contentLength, input,codes,badCodes,currentOptions.cookies,currentOptions.headers,hostHeadersFile,headersFile,pathsFile = GetCommandLineArgs()
	
	if input == ""{
		help()
	}

	if badCodes != "" {
		currentOptions.badCodesArr = strings.Split(badCodes,",")
	}
	if codes != "" {
		currentOptions.codesArr = strings.Split(codes,",")
	}

	from_file_host_header_checks,_ := readLines(hostHeadersFile)
	from_file_header_checks,_ := readLines(headersFile)
	from_file_path_checks,_ := readLines(pathsFile)

	parsedInput, err:= url.Parse(input)
	if err != nil{
		fmt.Printf("Invalid URL: %s", err.Error())
	}
	if parsedInput.Path == "" {
		parsedInput.Path = "/"
	}
	path_checks := FormatWordlist(from_file_path_checks, parsedInput)
	header_checks := FormatWordlist(from_file_header_checks, parsedInput)
	host_header_checks := FormatWordlist(from_file_host_header_checks, parsedInput)

	ExploitChecks(path_checks, header_checks, host_header_checks, parsedInput, currentOptions, contentLength)
}