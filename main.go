package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"net/http/cookiejar"
	"os"
	"regexp"
	"strconv"
	"time"

	instaddr "github.com/aiwao/instaddr_api"
	"github.com/aiwao/rik"
	"github.com/corpix/uarand"
	_ "github.com/lib/pq"
)

const registrationURL = "https://secure.sakura.ad.jp/serviceidp/api/v1/user/registration/"
var domains []string

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)
	log.Println("Connect to the database: " + dsn)
	var db *sql.DB
	for range 10 {
		tempDB, err := sql.Open("postgres", dsn)
		if err == nil {
			if err = tempDB.Ping(); err == nil {
				db = tempDB
				break
			}
		}
    	time.Sleep(2 * time.Second)
	}
	if db == nil {
		log.Fatalln("Database connection timed out")
	}
	defer db.Close()	
	
	acc, err := instaddr.NewAccount(instaddr.Options{})
	if err != nil {
		log.Fatalln(err)	
	}

	mailCnt := 0
	newAccountFlag := false
	for {
		if mailCnt >= 50 || newAccountFlag {
			newAcc, err := instaddr.NewAccount(instaddr.Options{})
			if err != nil {
				log.Println(err)
			} else {
				acc = newAcc
				mailCnt = 0
				newAccountFlag = false
			}
		}

		domain := "mail4.uk"
		if domains == nil || len(domains) == 0 {
			domains, err = acc.GetMailDomains(instaddr.Options{})
			if err != nil {
				log.Println(err)
			}
		} else {
			domain = domains[rand.IntN(len(domains))]
		}

		instaddrAuthInfo, err := acc.GetAuthInfo(instaddr.Options{})
		if err != nil {
			newAccountFlag = true
			log.Println(err)
			continue
		}

		ua := uarand.GetRandom()
		jar, err := cookiejar.New(nil)
		if err != nil {
			log.Println(err)
			continue
		}
		client := &http.Client{
			Jar: jar,
		}

		mailAcc, err := acc.CreateAddressWithDomainAndName(instaddr.OptionsWithName{Name: randStr(39, false)}, domain)
		if err != nil {
			log.Println(err)
			continue
		}
		mailCnt++
		
		email := mailAcc.Address
		log.Println("New mail address: "+mailAcc.Address)

		res, err := rik.Post(registrationURL+"email/").
			Header("User-agent", ua).
			Header("X-Csrftoken", "undefined").
			JSON(rik.NewJSON().
				Set("email", email).
				Build(),
			).			
			DoClient(client)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("Register email status: %d\n", res.StatusCode)

		verifyCode := ""
		for range 20 {
			previews, err := acc.SearchMail(instaddr.SearchOptions{Query: email})
			if err != nil {
				log.Println(err)
				goto SLEEP
			}
			for _, p  := range previews {
				mail, err := acc.ViewMail(instaddr.Options{}, p)
				if err != nil {
					log.Println(err)
					goto SLEEP
				}

				re := regexp.MustCompile(`\d+`)
				match := re.FindAllString(mail.Content, -1)
				if len(match) > 0 {
					_, err = strconv.Atoi(match[len(match)-1])
					if err != nil {
						log.Println(err)
						goto SLEEP
					}
					verifyCode = match[len(match)-1]
					goto DONE
				}
			}
			SLEEP:
				time.Sleep(1*time.Second)
			DONE:
				if verifyCode != "" {
					break
				}
		}
		if verifyCode == "" {
			log.Println("failed to verify")
			continue
		}
		log.Printf("Received verification code: %s\n", verifyCode)

		res, err = rik.Post(registrationURL+"code/").
			Header("User-agent", ua).
			Header("X-Csrftoken", "undefined").
			JSON(rik.NewJSON().
				Set("code", verifyCode).
				Build(),
			).
			DoClient(client)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("Code verification status: %d\n", res.StatusCode)
		
		password := randStr(16, true)
		res, err = rik.Post(registrationURL).
			Header("User-agent", ua).
			Header("X-Csrftoken", "undefined").
			JSON(rik.NewJSON().
				Set("full_name", "").
				Set("password", password).
				Build(),
			).
			DoClient(client)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("Registration status: %d\n", res.StatusCode)
		if res.StatusCode == http.StatusOK {
			log.Println("New Sakura ID was created")
			log.Printf("%s:%s\n", email, password)
			_, err = db.Exec(
				`INSERT INTO accounts(email, password, instaddr_id, instaddr_password) VALUES ($1, $2, $3, $4)`,
				email,
				password,
				instaddrAuthInfo.AccountID,
				instaddrAuthInfo.Password,
			)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println("Account was stored to the database\n")
		}
	}
}

const lower = "abcdefghijklmnopqrstuvwxyz"
const upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const alphabet = lower+upper
const numbers = "0123456789"

func randStr(length int, modePassword bool) string {
	b := make([]byte, length)
	alphabetLen := len(alphabet)
	numbersLen := len(numbers)
	for i := range b {
		var randChar byte
		if modePassword {
			if i > length/2 {
				randChar = numbers[rand.IntN(numbersLen)]
			} else {
				randChar = alphabet[rand.IntN(alphabetLen)]
			}
		} else {
			randChar = (lower+numbers)[rand.IntN(len(lower)+numbersLen)]
		}
		b[i] = randChar
	}
	return string(b)
}
