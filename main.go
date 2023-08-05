package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"newTweetScrap/config"
	"newTweetScrap/database"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	twitterscraper "github.com/torvald2/twitter-scrap"
	translator "github.com/turk/free-google-translate"
	tele "gopkg.in/telebot.v3"
)

var t = translator.NewTranslator(&http.Client{})

func main() {

	data, err := config.New("config")
	if err != nil {
		log.Fatal(err)
	}

	options := &tele.SendOptions{ParseMode: tele.ModeHTML, DisableWebPagePreview: true}

	scraper := twitterscraper.New().WithDelay(3)
	scraper.SetSearchMode(twitterscraper.SearchLatest)
	if code := os.Getenv("CODE"); code != "" {
		err = scraper.Login(data.TwitterLogin, data.TwitterPassword, code)
		if err != nil {
			panic(err)
		}
	} else {
		err = scraper.Login(data.TwitterLogin, data.TwitterPassword, data.TwitterEmail)
		if err != nil {
			panic(err)
		}
	}

	subscribe := fmt.Sprintf(`<a href="%s">Подписаться</a> | <a href="%s">Subscribe</a>`, data.Subscribe, data.Subscribe)
	ourChannels := fmt.Sprintf(`<a href="%s">Наши каналы</a> | <a href="%s">Our channels</a>`, data.OurChannels, data.OurChannels)
	contactUs := fmt.Sprintf(`<a href="%s">Связаться с нами</a> | <a href="%s">Contact us</a>`, data.ContactUs, data.ContactUs)

	db, err := database.NewDB("tweets.db")
	if err != nil {
		log.Fatal(err)
		return
	}
	pref := tele.Settings{
		Token:     data.TOKEN,
		Poller:    &tele.LongPoller{Timeout: 10 * time.Second},
		ParseMode: tele.ModeHTML,
	}
	channelID := &tele.User{ID: data.ChannelID}

	bot, _ := tele.NewBot(pref)

	go func() {
		for {
			db.MustExec(database.DeleteQuery, time.Now().Add(-time.Hour*24))
			time.Sleep(time.Hour * 24)
		}
	}()

	for {
	USER_LOOP:
		for _, user := range data.Users {
			tweets := make([]*twitterscraper.TweetResult, 0)
			for tweet := range scraper.SearchTweets(context.Background(), fmt.Sprintf("from:%s", user), 5) {
				if tweet.Error != nil {
					//r = tweet.Error
					//i++
					log.Printf("Error while trying to get post from %s; err - %s", user, tweet.Error)
					time.Sleep(time.Minute * 45)
					goto USER_LOOP
				}
				tweets = append(tweets, tweet)
			}
			if len(tweets) == 0 {
				continue
			}
			sort.Slice(tweets, func(i, j int) bool {
				return tweets[i].Timestamp > tweets[j].Timestamp
			})

			for _, tweet := range tweets {

				//r = nil
				txt := ""
				_, err = db.Exec(database.InsertionQuery, tweet.ID, time.Now())
				if err != nil {
					continue
				}

				postUrl := tweet.PermanentURL

				txt += fmt.Sprintf("<code>%s</code>\n\n", getDate(tweet.Timestamp))
				txt += fmt.Sprintf(`<a href="https://twitter.com/%s">@%s</a>`, tweet.Username, tweet.Username) + "\n"

				tex := tweet.Text

				var retweet *twitterscraper.Tweet
				if tweet.IsRetweet {
					retweet, _ = scraper.GetTweet(tweet.RetweetedStatusID)
					if retweet != nil {
						tex = retweet.Text
					}

				}
				var quoted *twitterscraper.Tweet
				if tweet.IsQuoted {
					quoted, err = scraper.GetTweet(tweet.QuotedStatusID)
					if err != nil {
						log.Println("Error while trying to get quoted post")
					}
				}

				if len(tex) > 240 {
					tex = tex[:240] + "..."
				}

				orgText, transtext := normalizeText(&tweet.Tweet, tex, retweet)

				txt += orgText
				txt += transtext

				tags := findTxt(tweet.HTML)
				for k, v := range tags {
					txt = strings.Replace(txt, k, fmt.Sprintf(`<a href="%s">%s</a>`, v, k), -1)
				}

				media := tele.Album{}

				if tweet.IsQuoted {

					txt += "Quoted tweet:\nПроцитированный твит:\n\n"

					txt += fmt.Sprintf(`<a href="https://tweeter.com/%s">@%s</a>`, quoted.Username, quoted.Username) + "\n"

					orgTextq, transtextq := normalizeText(quoted, quoted.Text, retweet)

					txt += orgTextq
					txt += transtextq

					txt += "<code>via Twitter</code>\n" + postUrl + "\n\n\n"

					txt += subscribe + "\n"
					txt += ourChannels + "\n"
					txt += contactUs + "\n"

					media = extractMedia(tweet.Tweet, txt, quoted)

					if len(tweet.Videos) == 0 && len(tweet.Photos) == 0 && len(tweet.GIFs) == 0 && len(media) == 0 {
						_, err = bot.Send(channelID, txt, options)
						if err != nil {
							log.Println(err)
							return
						}
					}
					_, err = bot.SendAlbum(channelID, media, options)
					if err != nil {
						log.Println(err)
					}
					continue

				} else if tweet.IsRetweet {
					txt += "<code>via Twitter</code>\n" + postUrl + "\n\n\n"
					txt += subscribe + "\n"
					txt += ourChannels + "\n"
					txt += contactUs + "\n"
					if retweet != nil {
						media = extractMedia(*retweet, txt, quoted)
					} else {
						media = extractMedia(tweet.Tweet, txt, quoted)
					}

					if len(tweet.Videos) == 0 && len(tweet.Photos) == 0 && len(tweet.GIFs) == 0 && len(media) == 0 {
						_, err = bot.Send(channelID, txt, options)
						if err != nil {
							log.Println(err)
							return
						}
					}
					_, nerr := bot.SendAlbum(channelID, media, options)
					if nerr != nil {
						log.Println(nerr)
					}
					continue

				}

				if !tweet.IsQuoted {
					txt += "<code>via Twitter</code>\n" + postUrl + "\n\n\n"

					txt += subscribe + "\n"
					txt += ourChannels + "\n"
					txt += contactUs + "\n"
					media = extractMedia(tweet.Tweet, txt, quoted)
				}

				if len(tweet.Videos) == 0 && len(tweet.Photos) == 0 && len(tweet.GIFs) == 0 && len(media) == 0 {
					_, err = bot.Send(channelID, txt, options)
					if err != nil {
						log.Println(err)
						return
					}
				}
				_, err = bot.SendAlbum(channelID, media, options)
				if err != nil {
					log.Println(err)
				}
			}

			time.Sleep(time.Minute * 2)
		}
	}

}

func getDate(unixTime int64) string {
	t := time.Unix(unixTime, 0)
	parsedTime := t.Format("02.01.2006")
	return parsedTime

}

func extractMedia(tweet twitterscraper.Tweet, text string, quoted *twitterscraper.Tweet) tele.Album {
	media := tele.Album{}
	fmt.Printf("%+v\n", tweet)
	fmt.Printf("%+v\n", text)
	fmt.Printf("%+v\n", quoted)
	if tweet.IsQuoted {

		tweet.Videos = append(tweet.Videos, quoted.Videos...)
		tweet.Photos = append(tweet.Photos, quoted.Photos...)
		tweet.GIFs = append(tweet.GIFs, quoted.GIFs...)

	}

	for _, video := range tweet.Videos {
		if len(media) == 0 {
			media = append(media, &tele.Video{File: tele.FromURL(video.URL), Caption: text})
			continue

		}
		media = append(media, &tele.Video{File: tele.FromURL(video.URL)})

	}
	for _, gifs := range tweet.GIFs {
		if len(media) == 0 {
			media = append(media, &tele.Video{File: tele.FromURL(gifs.URL), Caption: text})
			continue

		}
		media = append(media, &tele.Video{File: tele.FromURL(gifs.URL)})

	}

	for _, photos := range tweet.Photos {
		if len(media) == 0 {
			media = append(media, &tele.Photo{File: tele.FromURL(photos.URL), Caption: text})
			continue

		}
		media = append(media, &tele.Photo{File: tele.FromURL(photos.URL)})

	}

	return media

}

func findTxt(text string) map[string]string {
	pattern := `<a[^>]*href="([^"]*)"[^>]*>([^<]*)<\/a>`

	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(text, -1)

	tagMap := make(map[string]string)

	if len(matches) > 0 {
		for _, match := range matches {
			href := match[1]
			innerText := match[2]
			tagMap[innerText] = href
		}

	}

	return tagMap
}

func normalizeText(tweet *twitterscraper.Tweet, originalText string, retweet *twitterscraper.Tweet) (string, string) {

	var (
		orgText   string
		transText string
		err       error
	)

	tag := "<code>TW: </code>"
	if tweet.IsRetweet {
		originalText = strings.Replace(originalText, "RT ", "<code>RT </code>", 1)
		tag = ""
	}
	if tweet.IsQuoted {
		tag = "<code>CT: </code>"
	}

	urlPattern := `https://t\.co/\S+`

	r, err := regexp.Compile(urlPattern)
	if err != nil {
		fmt.Println("Error compiling regex pattern:", err)

	}

	originalText = r.ReplaceAllString(originalText, "")

	if len(originalText) == 0 {
		if len(tweet.URLs) == 0 {
			originalText = "This tweet contains only media"
		} else {
			originalText = "This tweet contains only link"

		}
	}
	var translatedText string
	if len(strings.TrimSpace(originalText)) != 0 {
		translatedText, err = t.Translate(originalText, "auto", "ru")
		if err != nil {
			log.Println("error while trying to translate post", err)
		}

	}
	originalText = strings.Trim(originalText, "\n")
	links := ""
	for _, url := range tweet.URLs {
		if len(url) > 40 {
			url = fmt.Sprintf(`<a href="%s">%s</a>...`, url, url[:40])
		}

		links += url + "\n"

	}
	if tweet.IsRetweet && len(links) == 0 && retweet != nil {
		fmt.Printf("%+v\n", retweet)
		for _, url := range retweet.URLs {
			if len(url) > 40 {
				url = fmt.Sprintf(`<a href="%s">%s</a>...`, url, url[:40])
			}
			links += url + "\n"
		}

	}

	if strings.TrimRight(originalText, "\n") == "This tweet contains only media" || strings.TrimRight(originalText, "\n") == "This tweet contains only link" {
		originalText = fmt.Sprintf("<i>%s</i>", originalText)
		translatedText = fmt.Sprintf("<i>Этот твит содержит только медиа</i>")
	}

	orgText = strings.Replace(orgText, "https", "", -1)
	transText = strings.Replace(orgText, "https", "", -1)

	if len(links) != 0 {
		orgText = tag + strings.Trim(originalText, "\n") + "\n" + links + "\n\n"
	} else {
		orgText += tag + originalText + "\n\n"
	}

	if !strings.Contains(strings.TrimSpace(translatedText), strings.TrimSpace(originalText)) {
		if len(links) != 0 {
			translatedText = strings.TrimRight(translatedText, "\n") + "\n"
			translatedText += links + "\n\n"
		}
		transText += tag + strings.TrimRight(translatedText, "\n") + "\n\n\n"
	}

	return orgText, transText
}
