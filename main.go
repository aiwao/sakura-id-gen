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

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)
	log.Println("Database: " + dsn)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	acc, err := instaddr.NewAccount(instaddr.Options{})
	if err != nil {
		log.Fatalln(err)	
	}

	for {
		ua := uarand.GetRandom()
		jar, err := cookiejar.New(nil)
		if err != nil {
			log.Println(err)
			continue
		}
		client := &http.Client{
			Jar: jar,
		}

		mailAcc, err := acc.CreateAddressWithExpiration(instaddr.Options{})
		if err != nil {
			log.Println(err)
			continue
		}
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
		for {
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
		
		password := randStr()
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
			_, err = db.Exec(`INSERT INTO accounts(email, password) VALUES ($1, $2)`, email, password)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println("Account was stored to the database\n")
		}
	}
}

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const numbers = "0123456789"
const passwordLength = 16

func randStr() string {
	b := make([]byte, passwordLength)
	for i := range b {
		var randChar byte
		if i > passwordLength/2 {
			randChar = numbers[rand.IntN(len(numbers))]
		} else {
			randChar = alphabet[rand.IntN(len(alphabet))]
		}
		b[i] = randChar
	}
	return string(b)
}
