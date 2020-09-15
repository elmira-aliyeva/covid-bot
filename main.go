package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Country struct {
	Country             string `json:"country"`
	Cases               int    `json:"cases"`
	TodayCases          int    `json:"todayCases"`
	Deaths              int    `json:"deaths"`
	TodayDeaths         int    `json:"todayDeaths"`
	Recovered           int    `json:"recovered"`
	Active              int    `json:"active"`
	Critical            int    `json:"critical"`
	CasesPerOneMillion  int    `json:"casesPerOneMillion"`
	DeathsPerOneMillion int    `json:"deathsPerOneMillion"`
	TotalTests          int    `json:"totalTests"`
	TestsPerOneMillion  int    `json:"testsPerOneMillion"`
}

var country Country

var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("Памятка пациенту", "https://www.coronavirus2020.kz/ru/patient"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("Домашний карантин", "https://www.coronavirus2020.kz/ru/home"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("Информация по лекарственным препаратам", "https://www.coronavirus2020.kz/ru/drug"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("Часто задаваемые вопросы", "https://www.coronavirus2020.kz/ru/faq"),
	),
)

func main() {
	bot, err := tgbotapi.NewBotAPI("944057915:AAEjhAQUN-G0rKpQ_bvdCG0vlVI8ldBALL8")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.Text == "/news" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			doc, err := goquery.NewDocument("https://www.coronavirus2020.kz/ru/news")
			if err != nil {
				log.Fatalln(err)
			}

			var last int = -1

			doc.Find(".lenta_news_block .lenta_news_time-rubric").Each(func(index int, item *goquery.Selection) {
				date := item.Text()[1:11]
				dateParsed, _ := time.Parse("02.01.2006", date)
				today := time.Now()
				if dateParsed.Year() == today.Year() && dateParsed.YearDay() == today.YearDay() {
					last = index
				}
			})

			if last == -1 {
				doc.Find(".lenta_news_block .lenta_news_time-rubric").Each(func(index int, item *goquery.Selection) {
					date := item.Text()[1:11]
					dateParsed, _ := time.Parse("02.01.2006", date)
					yesterday := time.Now().AddDate(0, 0, -1)
					if dateParsed.Year() == yesterday.Year() && dateParsed.YearDay() == yesterday.YearDay() {
						last = index
					}
				})
				msg.Text += "<b>Новости за " + time.Now().AddDate(0, 0, -1).Format("02.01.2006") + ":</b>\n"
			} else {
				msg.Text += "<b>Новости сегодня:</b>\n"
			}

			doc.Find(".lenta_news_block .lenta_news_title").Each(func(index int, item *goquery.Selection) {
				if index <= last {
					title := item.Text()
					linkTag := item.Find("a")
					link, _ := linkTag.Attr("href")
					msg.Text += "<a href=\"https://www.coronavirus2020.kz" + link + "\">🔹" + title + "</a>\n\n"
				}
			})

			msg.ParseMode = "html"
			bot.Send(msg)

		} else if update.Message.Text == "/stats" {

			resp, err := http.Get("https://coronavirus-19-api.herokuapp.com/countries/kazakhstan")
			if err != nil {
				log.Fatalln(err)
			}

			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}

			json.Unmarshal(body, &country)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.Text = "Зарегистрированных случаев: <b>" + strconv.Itoa(country.Cases) + "</b>\nЗарегистрированные случаи сегодня: <b>" + strconv.Itoa(country.TodayCases) + "</b>\nВыздоровевших: <b>" + strconv.Itoa(country.Recovered) + "</b>\nЛетальных случаев: <b>" + strconv.Itoa(country.Deaths) + "</b>"

			msg.ParseMode = "html"
			bot.Send(msg)

		} else if update.Message.Text == "/info" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Полезная информация по коронавирусу:")
			msg.ReplyMarkup = numericKeyboard
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			msg.Text = "/news - Свежие новости по ситуации с коронавирусом в Казахстане\n/stats - Данные по количеству зараженных в Казахстане\n/info -  Полезная информация по коронавирусу"
			bot.Send(msg)
		}
	}
}
