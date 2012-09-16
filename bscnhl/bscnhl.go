package bscnhl

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

type Game struct {
	HomePlayer, AwayPlayer     string
	HomeTeam, AwayTeam         string
	HomeScore, AwayScore       int
	HomeShots, AwayShots       int
	HomeHits, AwayHits         int
	Overtime                   bool
	HomeVerified, AwayVerified bool
	Winner                     bool
	Date                       time.Time
}

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/newgame", newgame)
	http.HandleFunc("/addgame", addgame)
}

type Index struct {
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
	i := Index{
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

var recentGamesTemplate = template.Must(template.ParseFiles("bscnhl/main.html"))

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
	fmt.Fprint(w, newGameTemplateHTML)
}

const newGameTemplateHTML = `
<html>
	<body>
		<form action="/addgame" method="post">
			<div><input type="radio" name="side" value="home"></input> Home</div>
			<div><input type="radio" name="side" value="away"></input> Away</div>
			<div><input type="checkbox" name="overtime" value="true"></input> Overtime</div>
			<div>Opponent: <input type="text" name="opponent"></input></div>
			<div>Home Team: <input type="text" name="hometeam"></input></div>
			<div>Away Team: <input type="text" name="awayteam"></input></div>
			<div>Home Goals: <input type="number" min="0" name="homegoals"></input></div>
			<div>Away Goals: <input type="number" min="0" name="awaygoals"></input></div>
			<div>Home Shots: <input type="number" min="0" name="homeshots"></input></div>
			<div>Away Shots: <input type="number" min="0" name="awayshots"></input></div>
			<div>Home Hits: <input type="number" min="0" name="homehits"></input></div>
			<div>Away Hits: <input type="number" min="0" name="awayhits"></input></div>
			<div><input type="submit" value="Add Game"></div>
		</form>
	</body>
</html>
`

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
		g.Winner = true
	} else {
		g.Winner = false
	}
	_, err = datastore.Put(c, datastore.NewIncompleteKey(c, "Game", nil), &g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
