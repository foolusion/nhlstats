package bscnhl

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

type Game struct {
	AwayPlayer, AwayTeam           string
	AwayScore, AwayShots, AwayHits int
	AwayVerified                   bool
	HomePlayer, HomeTeam           string
	HomeScore, HomeShots, HomeHits int
	HomeVerified                   bool
	HomeWon                        bool
	Overtime                       bool
	Date                           time.Time
}

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/newgame", newgame)
	http.HandleFunc("/addgame", addgame)
	http.HandleFunc("/profile", profile)
}

type index struct {
	Games []Game
	User  string
	Login bool
	URL   string
}

func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	var loginURL string
	if u == nil {
		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		loginURL = url
	} else {
		url, err := user.LogoutURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		loginURL = url
	}
	q := datastore.NewQuery("Game").Order("-Date").Limit(10)
	games := make([]Game, 0, 10)
	if _, err := q.GetAll(c, &games); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	i := index{
		Games: games,
		URL:   loginURL,
	}
	if u != nil {
		i.User = u.Email
		i.Login = false
	} else {
		i.Login = true
	}
	if err := recentGamesTemplate.Execute(w, i); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var recentGamesTemplate = template.Must(template.ParseFiles("bscnhl/main.html", "bscnhl/gamelist.html"))

func newgame(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	if u == nil {
		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return
	}
	url, err := user.LogoutURL(c, r.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	i := index{
		User:  u.Email,
		Login: false,
		URL:   url,
	}
	if err := newGameTemplate.Execute(w, i); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var newGameTemplate = template.Must(template.ParseFiles("bscnhl/main.html", "bscnhl/newgame.html"))

func addgame(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	homescore, err := strconv.Atoi(r.FormValue("homegoals"))
	if err != nil {
		http.Error(w, "Only Integers in Home Goals", http.StatusInternalServerError)
		return
	}
	homeshots, err := strconv.Atoi(r.FormValue("homeshots"))
	if err != nil {
		http.Error(w, "Only Integers in Home Shots", http.StatusInternalServerError)
		return
	}
	homehits, err := strconv.Atoi(r.FormValue("homehits"))
	if err != nil {
		http.Error(w, "Only Integers in Home Hits", http.StatusInternalServerError)
		return
	}
	awayscore, err := strconv.Atoi(r.FormValue("awaygoals"))
	if err != nil {
		http.Error(w, "Only Integers in Away Goals", http.StatusInternalServerError)
		return
	}
	awayshots, err := strconv.Atoi(r.FormValue("awayshots"))
	if err != nil {
		http.Error(w, "Only Integers in Away Shots", http.StatusInternalServerError)
		return
	}
	awayhits, err := strconv.Atoi(r.FormValue("awayhits"))
	if err != nil {
		http.Error(w, "Only Integers in Away Hits", http.StatusInternalServerError)
		return
	}
	g := Game{
		HomeTeam:  r.FormValue("hometeam"),
		HomeScore: homescore,
		HomeShots: homeshots,
		HomeHits:  homehits,
		AwayTeam:  r.FormValue("awayteam"),
		AwayScore: awayscore,
		AwayShots: awayshots,
		AwayHits:  awayhits,
		Date:      time.Now(),
	}
	if r.FormValue("overtime") == "true" {
		g.Overtime = true
	}
	if r.FormValue("side") == "home" {
		g.AwayPlayer = r.FormValue("opponent")
		if u := user.Current(c); u != nil {
			g.HomePlayer = u.String()
			g.HomeVerified = true
		}
	} else {
		g.HomePlayer = r.FormValue("opponent")
		if u := user.Current(c); u != nil {
			g.AwayPlayer = u.String()
			g.AwayVerified = true
		}
	}
	if g.HomeScore > g.AwayScore {
		g.HomeWon = true
	} else {
		g.HomeWon = false
	}
	_, err = datastore.Put(c, datastore.NewIncompleteKey(c, "Game", nil), &g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
