package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"time"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/hashicorp/go-retryablehttp"
	_ "modernc.org/sqlite"
	_"github.com/mattn/go-sqlite3"
)

type MIB struct {
	Node           string
	Name           string
	SubChildren    string
	SubNodesTotal  string
	Description    string
	Information    string
	LastRowFetched int
}

func createTable(db *sql.DB) error {
    _, err := db.Exec("CREATE TABLE IF NOT EXISTS mibs (Node TEXT, Name TEXT, SubChildren TEXT, SubNodesTotal TEXT, Description TEXT, Information TEXT, LastRowFetched INTEGER DEFAULT 0)")
    if err != nil {
        return err
    }
    return nil
}

// getLastRowFetched checks for any previously fetched rows and returns the last fetched row number
func getLastRowFetched(db *sql.DB) (int, error) {
    var lastRowFetched int
    row := db.QueryRow("SELECT COALESCE(LastRowFetched, 0) FROM mibs")
    err := row.Scan(&lastRowFetched)
    if err != nil {
        if err == sql.ErrNoRows {
            lastRowFetched = 0
        } else {
            return 0, err
        }
    }
    return lastRowFetched, nil
}

func createRetryableHttpClient() *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.RetryMax = 10
	client.RetryWaitMin = 1 * time.Second
	client.RetryWaitMax = 30 * time.Second
	return client
}

func main() {
	// Парсер командной строки
	outputFileName := flag.String("output", "mibs.sqlite", "output file name")
	flag.Parse()

	// Создайте подключение к базе данных
	db, err := sql.Open("sqlite3", *outputFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// Создать таблицу для данных MIB, если она не существует
	err = createTable(db)
    if err != nil {
        log.Fatal(err)
    }

	// Проверка наличия ранее извлеченных строк
	lastRowFetched, err := getLastRowFetched(db)
	if err != nil {
		log.Fatal(err)
	}

	// Создание HTTP-клиента с возможностью повторной попытки
	client := createRetryableHttpClient()

	// Загрузка веб-сайта и получение данных MIB
	groups := map[int][]string{}
	url := "https://oidref.com/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")

	retryReq, err := retryablehttp.FromRequest(req)
	if err != nil {
		log.Fatal(err)
	}

	var resp *http.Response
    for {
        // Выполните HTTP-запрос с повторными попытками
        resp, err = client.Do(retryReq)
        if err == nil || strings.Contains(err.Error(), "net/http: request canceled") {
            break
        }

        // Если попытки исчерпаны, сделайте паузу на час, прежде чем возобновить попытки
        if strings.Contains(err.Error(), "Too many retries") {
            fmt.Println("Retries exhausted, pausing for an hour...")
            time.Sleep(time.Hour)
        } else {
            fmt.Println(err)
        }
    }
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	x:=0
	table := doc.Find("table")
	tableLen := table.Find("tr").Length()
	tableLen = tableLen - lastRowFetched - 1
	table.Find("tr").Each(func(i int, row *goquery.Selection) {
		if i > lastRowFetched{
			row.Find("td").Each(func(j int, cell *goquery.Selection) {
				groups[j] = append(groups[j], cell.Text())
			})
			// Вставка извлеченных данных MIB в базу данных
			stmt, err := db.Prepare("INSERT INTO mibs (Node, Name, SubChildren, SubNodesTotal, Description, Information) VALUES (?, ?, ?, ?, ?, ?)")
			if err != nil {
				log.Fatal(err)
			}
			defer stmt.Close()
			if lastRowFetched == 0 {
				var text []string
				doc.Find("table th").Each(func(i int, s *goquery.Selection) {
					text = append(text ,s.Text())
				})
				_, err = stmt.Exec(text[0], text[1], text[2], text[3], text[4], text[5])
				if err != nil {
					panic(err)
				}
			}
			_, err = stmt.Exec(groups[0][x], groups[1][x], groups[2][x], groups[3][x], groups[4][x], groups[5][x])
			if err != nil {
				log.Fatal(err)
			}
			x+=1
			if i!=0{
				fmt.Printf("Downloading MIB descriptions: %d/%d (%d%%)\n", i-lastRowFetched, tableLen, (i-lastRowFetched)*100/tableLen)
			}
			// Обновление последнего извлеченного ряда в базе данных
			_, err = db.Exec("UPDATE mibs SET LastRowFetched = ?", i)
			if err != nil {
				log.Fatal(err)
			}
		}
	})
}
