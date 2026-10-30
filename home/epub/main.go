package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/raitucarp/epub"
	"golang.org/x/net/proxy"

	"github.com/anton2920/gofa/bools"
)

const (
	BaseURLFmt         = "https://www.virlyce.com/the-blue-mage-raised-by-dragons/book-%d-"
	ChapterSuffixFmt   = "chapter-%d/"
	AfterwordSuffixFmt = "afterword/"
)

var HR = []byte("<hr />")

func DownloadChapter(client *http.Client, book int, chapter int, sema chan struct{}) ([]byte, error) {
	var url string
	if chapter > 0 {
		url = fmt.Sprintf(BaseURLFmt+ChapterSuffixFmt, book, chapter)
	} else {
		url = fmt.Sprintf(BaseURLFmt+AfterwordSuffixFmt, book)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %v", err)
	}
	req.Header.Set("User-Agent", "curl/8.23.0")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do GET request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %q", resp.Status, body)
	}

	begin := bytes.Index(body, HR)
	if begin == -1 {
		return nil, errors.New("failed to find begin position")
	}

	end := bytes.Index(body[begin+1:], HR)
	if end == -1 {
		return nil, errors.New("failed to find end position")
	}
	end += begin + 1

	return body[begin+len(HR) : end], nil
}

func DownloadChapterWorker(wg *sync.WaitGroup, client *http.Client, result *[]byte, book int, chapter int, retries int, sema chan struct{}, errorChannel chan<- error) {
	defer wg.Done()

	sema <- struct{}{}
	defer func() {
		<-sema
	}()

	var err error
	for attempt := 1; attempt <= retries; attempt++ {
		log.Printf("[ INFO]: Downloading book %d chapter %d (attempt %d)...", book, chapter, attempt)
		*result, err = DownloadChapter(client, book, chapter, sema)
		if err != nil {
			errorChannel <- fmt.Errorf("failed to download book %d chapter %d (attempt %d): %v", book, chapter, attempt, err)
			continue
		}
		log.Printf("[ INFO]: Completed downloading book %d chapter %d (attempt %d)!", book, chapter, attempt)
		return
	}

	log.Fatalf("[FATAL]: exhausted all attempts to download book %d chapter %d!", book, chapter)
}

func DownloadBook(book int, chapters int, afterword bool, retries int, timeout time.Duration, threads int, delay time.Duration) ([][]byte, error) {
	var wg sync.WaitGroup

	dialer, err := proxy.SOCKS5("tcp", "localhost:6585", nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to establish proxy connection: %v")
	}
	_ = dialer

	f, err := os.Create("ssl-key.log")
	if err != nil {
		return nil, fmt.Errorf("failed to create (pre-)master shared key file: %v", err)
	}
	defer f.Close()

	var client http.Client
	client.Timeout = timeout
	client.Transport = &http.Transport{
		Dial: dialer.Dial,
	}

	result := make([][]byte, chapters+bools.ToInt(afterword))
	sema := make(chan struct{}, threads)
	errorChannel := make(chan error)
	done := make(chan struct{})

	wg.Add(len(result))
	for i := 0; i < chapters; i++ {
		go DownloadChapterWorker(&wg, &client, &result[i], book, i+1, retries, sema, errorChannel)
		time.Sleep(delay)
	}
	if afterword {
		go DownloadChapterWorker(&wg, &client, &result[len(result)-1], book, -1, retries, sema, errorChannel)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	var quit bool
	for !quit {
		select {
		case err := <-errorChannel:
			log.Printf("[ERROR]: %v", err)
		case <-done:
			quit = true
		}
	}

	return result, nil
}

func CreateBook(name string, book int, chapters [][]byte) error {
	var toc epub.TOC
	toc.Title = "Contents"

	w := epub.New("urn:isbn:9780000000001")
	w.Title("The Blue Mage Raised by Dragons Book #" + strconv.Itoa(book))
	w.Author("Virlyce")
	w.Languages("en")

	for i := 0; i < len(chapters); i++ {
		href := fmt.Sprintf("chapter-%d.xhtml", i+1)
		w.AddContent(href, []byte(fmt.Sprintf(`<html xmlns="http://www.w3.org/1999/xhtml"><body><h1>Chapter %d</h1>%s</body></html>`, i+1, chapters[i])))
		toc.Items = append(toc.Items, epub.TOC{Title: fmt.Sprintf("Chapter %d", i+1), Href: href})
	}

	if err := w.TableOfContents("toc", toc); err != nil {
		return fmt.Errorf("failed to create table of contents: %v", err)
	}

	if err := w.Write(name); err != nil {
		return fmt.Errorf("failed to write book to file: %v", err)
	}

	return nil
}

func main() {
	output := flag.String("output", "TheBlueMageRaisedByDragons.epub", "name of the output book file")
	book := flag.Int("book", 2, "specifies the book number to download")
	chapters := flag.Int("chapters", 130, "specifies the number of chapters in the book")
	afterword := flag.Bool("afterword", false, "specifies whether \"afterword\" requires special treatment")
	retries := flag.Int("retries", 10, "specifies the number of retries for downloading each chapter")
	timeout := flag.Duration("timeout", 5*time.Second, "amount of time to wait for chapter download before giving up")
	threads := flag.Int("t", runtime.GOMAXPROCS(0), "amount of simultaneous downloads")
	delay := flag.Duration("delay", 500*time.Millisecond, "delay time before scheduling next chapter download")
	flag.Parse()

	result, err := DownloadBook(*book, *chapters, *afterword, *retries, *timeout, *threads, *delay)
	if err != nil {
		s := [...]string{"regular", "special"}[bools.ToInt(*afterword)]
		log.Fatalf("Failed to download book %d with %d chapters and %s afterword: %v", *book, *chapters, s, err)
	}

	if err := CreateBook(*output, *book, result); err != nil {
		log.Fatalf("Failed to create book: %v", err)
	}

	log.Printf("DONE! ChatGPT can s**ck my d**ck! ;)")
}
